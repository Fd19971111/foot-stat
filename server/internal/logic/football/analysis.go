package football

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	v1 "server/api/football/v1"
)

const (
	analysisHalfLifeDays = 365.0 // 历史比赛权重半衰期：一年
	analysisMinWeight    = 0.05
	analysisPriorGames   = 6.0   // 攻防收缩先验场次
	analysisRho          = -0.10 // Dixon-Coles 低比分相关系数
	analysisMaxGoals     = 10
	eloInitialRating     = 1500.0
	eloKFactor           = 20.0
	eloHomeAdvantage     = 65.0
)

type weightedTeamStats struct {
	homeGames        int
	awayGames        int
	homeWeight       float64
	awayWeight       float64
	homeGoalsFor     float64
	homeGoalsAgainst float64
	awayGoalsFor     float64
	awayGoalsAgainst float64
}

type formResult struct {
	goalsFor     int
	goalsAgainst int
}

type leagueSnapshot struct {
	code     string
	matches  int
	homeMean float64
	awayMean float64
	teams    map[string]*weightedTeamStats
	ratings  map[string]float64
	lastFive map[string][]formResult
}

type probabilitySet struct {
	home float64
	draw float64
	away float64
}

type poissonResult struct {
	probabilitySet
	over25     float64
	btts       float64
	topScore   string
	scoreLines []v1.AnalysisScore
}

// AnalyzeImport 对导入批次逐场应用 Dixon-Coles 泊松模型、Elo 等级分和近期状态，输出综合分析结果。
func AnalyzeImport(ctx context.Context, id int64) (*v1.GetAnalysisRes, error) {
	db, err := openImportDB(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var batch v1.ImportBatch
	err = db.QueryRowContext(ctx, `SELECT id, file_name, match_count,
		DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s')
		FROM footstat_import_batches WHERE id = ?`, id).
		Scan(&batch.Id, &batch.FileName, &batch.MatchCount, &batch.ImportedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("导入批次不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("读取导入批次失败: %w", err)
	}

	rows, err := db.QueryContext(ctx, `SELECT match_date, kickoff_time, league, home_team, away_team
		FROM footstat_import_matches WHERE batch_id = ? ORDER BY match_date, id`, id)
	if err != nil {
		return nil, fmt.Errorf("读取导入比赛失败: %w", err)
	}
	defer rows.Close()

	matches := make([]v1.ImportMatch, 0, batch.MatchCount)
	for rows.Next() {
		var match v1.ImportMatch
		if err = rows.Scan(&match.Date, &match.Time, &match.League, &match.HomeTeam, &match.AwayTeam); err != nil {
			return nil, fmt.Errorf("解析导入比赛失败: %w", err)
		}
		matches = append(matches, match)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历导入比赛失败: %w", err)
	}

	type leagueCache struct {
		history   *leagueHistory
		snapshots map[int]*leagueSnapshot
	}
	caches := make(map[string]*leagueCache)
	today := time.Now().Format("20060102")
	todayCutoff, _ := strconv.Atoi(today)

	results := make([]v1.MatchAnalysis, 0, len(matches))
	var totalConfidence float64
	var highConfidence, homeLean, overLean int
	for _, match := range matches {
		cutoff := matchCutoff(match.Date)
		if cutoff == 0 {
			cutoff = todayCutoff
		}

		key := analysisLeagueKey(match.League)
		cache, ok := caches[key]
		if !ok {
			history, loadErr := loadLeagueHistory(ctx, match.League)
			if loadErr != nil {
				return nil, fmt.Errorf("加载 %s 历史数据失败: %w", match.League, loadErr)
			}
			sort.Slice(history.matches, func(i, j int) bool {
				return history.matches[i].date < history.matches[j].date
			})
			cache = &leagueCache{history: history, snapshots: make(map[int]*leagueSnapshot)}
			caches[key] = cache
		}

		snapshot, ok := cache.snapshots[cutoff]
		if !ok {
			snapshot = buildLeagueSnapshot(cache.history, cutoff)
			cache.snapshots[cutoff] = snapshot
		}

		analysis := analyzeMatch(snapshot, match)
		results = append(results, analysis)
		totalConfidence += analysis.Confidence
		if analysis.ConfidenceLevel == "高" {
			highConfidence++
		}
		if analysis.RecommendationType == "home" {
			homeLean++
		}
		if analysis.GoalsType == "over" {
			overLean++
		}
	}

	summary := v1.AnalysisSummary{
		MatchCount:     len(results),
		HighConfidence: highConfidence,
		HomeLean:       homeLean,
		OverLean:       overLean,
		Algorithm:      "Dixon-Coles 泊松模型 + Elo 等级分 + 近期状态",
	}
	if len(results) > 0 {
		summary.AverageConfidence = round(totalConfidence/float64(len(results)), 0)
	}

	return &v1.GetAnalysisRes{Batch: batch, Summary: summary, Matches: results}, nil
}

func analysisLeagueKey(league string) string {
	if code := resolveLeagueCode(league); code != "" {
		return code
	}
	return "raw:" + strings.ToLower(strings.TrimSpace(league))
}

func matchCutoff(date string) int {
	day, month, year := parseCsvDate(date)
	if year == 0 || month < 1 || month > 12 || day < 1 || day > 31 {
		return 0
	}
	return year*10000 + month*100 + day
}

// buildLeagueSnapshot 汇总赛前样本：时间衰减的攻防数据、Elo 等级分与最近五场状态。
func buildLeagueSnapshot(history *leagueHistory, cutoff int) *leagueSnapshot {
	snapshot := &leagueSnapshot{
		code:     history.code,
		homeMean: 1.45,
		awayMean: 1.15,
		teams:    make(map[string]*weightedTeamStats),
		ratings:  make(map[string]float64),
		lastFive: make(map[string][]formResult),
	}

	var homeGoalsWeighted, awayGoalsWeighted, totalWeight float64
	for _, item := range history.matches {
		if item.date >= cutoff {
			continue
		}
		days := daysBetween(item.date, cutoff)
		weight := math.Pow(0.5, float64(days)/analysisHalfLifeDays)
		if weight < analysisMinWeight {
			weight = analysisMinWeight
		}

		homeStats := snapshot.team(item.homeTeam)
		homeStats.homeGames++
		homeStats.homeWeight += weight
		homeStats.homeGoalsFor += weight * item.homeGoals
		homeStats.homeGoalsAgainst += weight * item.awayGoals

		awayStats := snapshot.team(item.awayTeam)
		awayStats.awayGames++
		awayStats.awayWeight += weight
		awayStats.awayGoalsFor += weight * item.awayGoals
		awayStats.awayGoalsAgainst += weight * item.homeGoals

		homeGoalsWeighted += weight * item.homeGoals
		awayGoalsWeighted += weight * item.awayGoals
		totalWeight += weight
		snapshot.matches++

		updateElo(snapshot.ratings, item)
		snapshot.addForm(item.homeTeam, int(item.homeGoals), int(item.awayGoals))
		snapshot.addForm(item.awayTeam, int(item.awayGoals), int(item.homeGoals))
	}

	if totalWeight > 0 {
		snapshot.homeMean = homeGoalsWeighted / totalWeight
		snapshot.awayMean = awayGoalsWeighted / totalWeight
	}
	return snapshot
}

func (s *leagueSnapshot) team(name string) *weightedTeamStats {
	stats, ok := s.teams[name]
	if !ok {
		stats = &weightedTeamStats{}
		s.teams[name] = stats
	}
	return stats
}

func (s *leagueSnapshot) addForm(name string, goalsFor, goalsAgainst int) {
	results := append(s.lastFive[name], formResult{goalsFor: goalsFor, goalsAgainst: goalsAgainst})
	if len(results) > 5 {
		results = results[len(results)-5:]
	}
	s.lastFive[name] = results
}

func updateElo(ratings map[string]float64, match historicalMatch) {
	homeRating, ok := ratings[match.homeTeam]
	if !ok {
		homeRating = eloInitialRating
	}
	awayRating, ok := ratings[match.awayTeam]
	if !ok {
		awayRating = eloInitialRating
	}

	expected := 1 / (1 + math.Pow(10, -(homeRating+eloHomeAdvantage-awayRating)/400))
	actual := 0.5
	if match.homeGoals > match.awayGoals {
		actual = 1
	} else if match.homeGoals < match.awayGoals {
		actual = 0
	}

	difference := math.Abs(match.homeGoals - match.awayGoals)
	multiplier := 1.0
	if difference == 2 {
		multiplier = 1.5
	} else if difference >= 3 {
		multiplier = (11 + difference) / 8
	}

	change := eloKFactor * multiplier * (actual - expected)
	ratings[match.homeTeam] = homeRating + change
	ratings[match.awayTeam] = awayRating - change
}

func daysBetween(from, to int) int {
	start := dateTime(from)
	end := dateTime(to)
	return int(end.Sub(start).Hours() / 24)
}

func dateTime(value int) time.Time {
	return time.Date(value/10000, time.Month((value/100)%100), value%100, 0, 0, 0, 0, time.UTC)
}

func analyzeMatch(snapshot *leagueSnapshot, match v1.ImportMatch) v1.MatchAnalysis {
	homeKey := teamKey(match.HomeTeam)
	awayKey := teamKey(match.AwayTeam)
	homeStats := snapshot.teams[homeKey]
	awayStats := snapshot.teams[awayKey]

	homeAttack := attackRate(homeStats, true, snapshot.homeMean)
	awayDefense := defenseRate(awayStats, false, snapshot.homeMean)
	awayAttack := attackRate(awayStats, false, snapshot.awayMean)
	homeDefense := defenseRate(homeStats, true, snapshot.awayMean)

	expectedHome := clamp(snapshot.homeMean*homeAttack*awayDefense, 0.15, 4.5)
	expectedAway := clamp(snapshot.awayMean*awayAttack*homeDefense, 0.1, 4.0)

	poissonResult := poissonOutcome(expectedHome, expectedAway)
	homeElo := ratingOf(snapshot.ratings, homeKey)
	awayElo := ratingOf(snapshot.ratings, awayKey)
	eloResult := eloProbabilities(homeElo, awayElo)

	final := probabilitySet{
		home: 0.65*poissonResult.home + 0.35*eloResult.home,
		draw: 0.65*poissonResult.draw + 0.35*eloResult.draw,
		away: 0.65*poissonResult.away + 0.35*eloResult.away,
	}
	final = normalize(final)

	confidence := analysisConfidence(snapshot, homeStats, awayStats, final, poissonResult.probabilitySet, eloResult)
	recommendationType, recommendation := recommendationOf(final)
	goalsType, goalsAdvice := "under", "小于 2.5 球"
	if poissonResult.over25 >= 0.5 {
		goalsType, goalsAdvice = "over", "大于 2.5 球"
	}
	bttsAdvice := "非双方进球"
	if poissonResult.btts >= 0.5 {
		bttsAdvice = "双方进球"
	}

	return v1.MatchAnalysis{
		MatchDate:         match.Date,
		KickoffTime:       match.Time,
		League:            match.League,
		HomeTeam:          match.HomeTeam,
		AwayTeam:          match.AwayTeam,
		ExpectedHomeGoals: round(expectedHome, 2),
		ExpectedAwayGoals: round(expectedAway, 2),
		Poisson:           percentSet(poissonResult.probabilitySet),
		Elo:               percentSet(eloResult),
		Final:             percentSet(final),
		Over25Probability: round(poissonResult.over25*100, 1),
		BttsProbability:   round(poissonResult.btts*100, 1),
		FairOdds: v1.AnalysisOdds{
			Home: fairOdds(final.home),
			Draw: fairOdds(final.draw),
			Away: fairOdds(final.away),
		},
		PredictedScore:     poissonResult.topScore,
		Recommendation:     recommendation,
		RecommendationType: recommendationType,
		GoalsAdvice:        goalsAdvice,
		GoalsType:          goalsType,
		BttsAdvice:         bttsAdvice,
		Confidence:         confidence,
		ConfidenceLevel:    confidenceLevel(confidence),
		HistoryMatches:     snapshot.matches,
		HomeTeamHistory:    teamSamples(homeStats),
		AwayTeamHistory:    teamSamples(awayStats),
		HomeForm:           buildTeamForm(snapshot.lastFive[homeKey], homeElo),
		AwayForm:           buildTeamForm(snapshot.lastFive[awayKey], awayElo),
		ScoreLines:         poissonResult.scoreLines,
		ModelBasis:         modelBasis(snapshot, homeStats, awayStats),
	}
}

func attackRate(stats *weightedTeamStats, atHome bool, leagueMean float64) float64 {
	if stats == nil || leagueMean <= 0 {
		return 1
	}
	weight, goals := stats.awayWeight, stats.awayGoalsFor
	if atHome {
		weight, goals = stats.homeWeight, stats.homeGoalsFor
	}
	mean := (goals + analysisPriorGames*leagueMean) / (weight + analysisPriorGames)
	return mean / leagueMean
}

func defenseRate(stats *weightedTeamStats, atHome bool, opponentMean float64) float64 {
	if stats == nil || opponentMean <= 0 {
		return 1
	}
	weight, goals := stats.awayWeight, stats.awayGoalsAgainst
	if atHome {
		weight, goals = stats.homeWeight, stats.homeGoalsAgainst
	}
	mean := (goals + analysisPriorGames*opponentMean) / (weight + analysisPriorGames)
	return mean / opponentMean
}

func ratingOf(ratings map[string]float64, team string) float64 {
	if rating, ok := ratings[team]; ok {
		return rating
	}
	return eloInitialRating
}

func teamSamples(stats *weightedTeamStats) int {
	if stats == nil {
		return 0
	}
	return stats.homeGames + stats.awayGames
}

func eloProbabilities(homeElo, awayElo float64) probabilitySet {
	difference := homeElo + eloHomeAdvantage - awayElo
	expected := 1 / (1 + math.Pow(10, -difference/400))
	draw := clamp(0.28-math.Abs(difference)*0.00035, 0.16, 0.30)
	home := expected - draw/2
	away := 1 - home - draw
	return normalize(probabilitySet{home: home, draw: draw, away: away})
}

func poissonOutcome(homeLambda, awayLambda float64) poissonResult {
	homeGoals := poissonProbabilities(homeLambda, analysisMaxGoals)
	awayGoals := poissonProbabilities(awayLambda, analysisMaxGoals)
	matrix := make([][]float64, len(homeGoals))
	total := 0.0
	for home := range homeGoals {
		matrix[home] = make([]float64, len(awayGoals))
		for away := range awayGoals {
			value := homeGoals[home] * awayGoals[away] * dixonColesTau(home, away, homeLambda, awayLambda)
			matrix[home][away] = value
			total += value
		}
	}

	var result poissonResult
	result.scoreLines = make([]v1.AnalysisScore, 0, len(homeGoals)*len(awayGoals))
	if total <= 0 {
		result.topScore = "0 - 0"
		return result
	}

	for home := range matrix {
		for away := range matrix[home] {
			probability := matrix[home][away] / total
			switch {
			case home > away:
				result.home += probability
			case home == away:
				result.draw += probability
			default:
				result.away += probability
			}
			if home+away > 2 {
				result.over25 += probability
			}
			if home > 0 && away > 0 {
				result.btts += probability
			}
			result.scoreLines = append(result.scoreLines, v1.AnalysisScore{
				Score:       fmt.Sprintf("%d - %d", home, away),
				Probability: probability,
			})
		}
	}

	sort.SliceStable(result.scoreLines, func(i, j int) bool {
		return result.scoreLines[i].Probability > result.scoreLines[j].Probability
	})
	if len(result.scoreLines) > 5 {
		result.scoreLines = result.scoreLines[:5]
	}
	for index := range result.scoreLines {
		result.scoreLines[index].Probability = round(result.scoreLines[index].Probability*100, 1)
	}
	result.topScore = result.scoreLines[0].Score
	return result
}

// dixonColesTau 对 0-0、0-1、1-0、1-1 四个低比分做相关性修正。
func dixonColesTau(home, away int, homeLambda, awayLambda float64) float64 {
	switch {
	case home == 0 && away == 0:
		return 1 - homeLambda*awayLambda*analysisRho
	case home == 0 && away == 1:
		return 1 + homeLambda*analysisRho
	case home == 1 && away == 0:
		return 1 + awayLambda*analysisRho
	case home == 1 && away == 1:
		return 1 - analysisRho
	default:
		return 1
	}
}

func normalize(set probabilitySet) probabilitySet {
	total := set.home + set.draw + set.away
	if total <= 0 {
		return probabilitySet{home: 1.0 / 3, draw: 1.0 / 3, away: 1.0 / 3}
	}
	return probabilitySet{home: set.home / total, draw: set.draw / total, away: set.away / total}
}

func percentSet(set probabilitySet) v1.AnalysisProbability {
	return v1.AnalysisProbability{
		HomeWin: round(set.home*100, 1),
		Draw:    round(set.draw*100, 1),
		AwayWin: round(set.away*100, 1),
	}
}

func fairOdds(probability float64) float64 {
	if probability <= 0 {
		return 0
	}
	return round(1/probability, 2)
}

func analysisConfidence(snapshot *leagueSnapshot, homeStats, awayStats *weightedTeamStats, final, poisson, elo probabilitySet) float64 {
	leagueFactor := clamp(float64(snapshot.matches)/300, 0, 1)
	minimumSamples := teamSamples(homeStats)
	if awaySamples := teamSamples(awayStats); awaySamples < minimumSamples {
		minimumSamples = awaySamples
	}
	teamFactor := clamp(float64(minimumSamples)/12, 0, 1)
	dataFactor := 0.5*leagueFactor + 0.5*teamFactor

	topProbability := math.Max(final.home, math.Max(final.draw, final.away))
	certainty := clamp((topProbability-1.0/3)/(2.0/3), 0, 1)

	agreement := 1 - (math.Abs(poisson.home-elo.home)+math.Abs(poisson.draw-elo.draw)+math.Abs(poisson.away-elo.away))/2
	agreement = clamp(agreement, 0, 1)

	score := 100 * (0.35*dataFactor + 0.65*certainty)
	if agreement < 0.85 {
		score -= (0.85 - agreement) * 40
	}
	if snapshot.matches == 0 {
		score = clamp(score, 20, 35)
	} else if minimumSamples == 0 {
		score = clamp(score, 25, 45)
	}
	return round(score, 0)
}

func confidenceLevel(confidence float64) string {
	switch {
	case confidence >= 75:
		return "高"
	case confidence >= 55:
		return "中"
	default:
		return "低"
	}
}

func recommendationOf(set probabilitySet) (string, string) {
	switch {
	case set.home >= set.draw && set.home >= set.away:
		return "home", "主胜"
	case set.away >= set.home && set.away >= set.draw:
		return "away", "客胜"
	default:
		return "draw", "平局"
	}
}

func buildTeamForm(results []formResult, elo float64) v1.AnalysisTeamForm {
	form := v1.AnalysisTeamForm{Matches: len(results), Elo: round(elo, 0)}
	var sequence strings.Builder
	for _, item := range results {
		switch {
		case item.goalsFor > item.goalsAgainst:
			form.Wins++
			form.Points += 3
			sequence.WriteString("W")
		case item.goalsFor == item.goalsAgainst:
			form.Draws++
			form.Points++
			sequence.WriteString("D")
		default:
			form.Losses++
			sequence.WriteString("L")
		}
		form.GoalsFor += item.goalsFor
		form.GoalsAgainst += item.goalsAgainst
	}
	form.Sequence = sequence.String()
	return form
}

func modelBasis(snapshot *leagueSnapshot, homeStats, awayStats *weightedTeamStats) string {
	homeSamples, awaySamples := teamSamples(homeStats), teamSamples(awayStats)
	switch {
	case snapshot.code == "":
		return "未匹配联赛历史，使用通用进球先验与 1500 初始等级分"
	case snapshot.matches == 0:
		return fmt.Sprintf("%s 暂无赛前历史样本，使用通用先验估算", snapshot.code)
	case homeSamples == 0 || awaySamples == 0:
		return fmt.Sprintf("%s · 近 10 赛季 %d 场样本 · 球队样本不足，按联赛基准与 Elo 估算", snapshot.code, snapshot.matches)
	default:
		return fmt.Sprintf("%s · 近 10 赛季 %d 场样本 · 主队 %d 场 / 客队 %d 场 · 泊松与 Elo 融合", snapshot.code, snapshot.matches, homeSamples, awaySamples)
	}
}
