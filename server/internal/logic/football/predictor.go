package football

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	v1 "server/api/football/v1"
)

type historicalMatch struct {
	date      int
	homeTeam  string
	awayTeam  string
	homeGoals float64
	awayGoals float64
	market    probabilitySet
	hasMarket bool
}

type teamGoalStats struct {
	homeGames        int
	homeGoalsFor     float64
	homeGoalsAgainst float64
	awayGames        int
	awayGoalsFor     float64
	awayGoalsAgainst float64
}

type leagueHistory struct {
	code    string
	matches []historicalMatch
}

var leagueAliases = map[string]string{
	"英超": "E0", "英格兰超级联赛": "E0", "premier league": "E0",
	"英冠": "E1", "英格兰冠军联赛": "E1", "championship": "E1",
	"英甲": "E2", "英格兰甲级联赛": "E2",
	"英乙": "E3", "英格兰乙级联赛": "E3",
	"德甲": "D1", "德国甲级联赛": "D1", "bundesliga": "D1",
	"德乙": "D2", "德国乙级联赛": "D2",
	"意甲": "I1", "意大利甲级联赛": "I1", "serie a": "I1",
	"意乙": "I2", "意大利乙级联赛": "I2",
	"西甲": "SP1", "西班牙甲级联赛": "SP1", "la liga": "SP1",
	"西乙": "SP2", "西班牙乙级联赛": "SP2",
	"法甲": "F1", "法国甲级联赛": "F1", "ligue 1": "F1",
	"法乙": "F2", "法国乙级联赛": "F2",
	"苏超": "SC0", "苏格兰超级联赛": "SC0",
	"比甲": "B1", "比利时甲级联赛": "B1",
	"荷甲": "N1", "荷兰甲级联赛": "N1",
	"葡超": "P1", "葡萄牙超级联赛": "P1",
	"土超": "T1", "土耳其超级联赛": "T1",
	"希超": "G1", "希腊超级联赛": "G1",
}

func resolveLeagueCode(value string) string {
	value = strings.TrimSpace(value)
	upper := strings.ToUpper(value)
	if _, ok := leagueName[upper]; ok {
		return upper
	}
	if separator := strings.IndexAny(upper, " -_"); separator > 0 {
		candidate := upper[:separator]
		if _, ok := leagueName[candidate]; ok {
			return candidate
		}
	}
	if code, ok := leagueAliases[strings.ToLower(value)]; ok {
		return code
	}
	for alias, code := range leagueAliases {
		if strings.Contains(strings.ToLower(value), alias) {
			return code
		}
	}
	return ""
}

func loadLeagueHistory(ctx context.Context, league string) (*leagueHistory, error) {
	code := resolveLeagueCode(league)
	history := &leagueHistory{code: code, matches: make([]historicalMatch, 0)}
	if code == "" {
		return history, nil
	}

	seasons := GetSeasons(ctx)
	if len(seasons) > 10 {
		seasons = seasons[:10]
	}
	for _, season := range seasons {
		path := filepath.Join(DataDir(), season.Value, code+".csv")
		columns, rows := GetMatchesByPath(ctx, path)
		if len(columns) == 0 {
			continue
		}
		for _, row := range rows {
			homeGoals, homeErr := strconv.ParseFloat(row["FTHG"], 64)
			awayGoals, awayErr := strconv.ParseFloat(row["FTAG"], 64)
			if homeErr != nil || awayErr != nil || row["HomeTeam"] == "" || row["AwayTeam"] == "" {
				continue
			}
			day, month, year := parseCsvDate(row["Date"])
			if year == 0 || month == 0 || day == 0 {
				continue
			}
			date := year*10000 + month*100 + day
			market, hasMarket := marketProbabilitiesFromOdds(row)
			history.matches = append(history.matches, historicalMatch{
				date: date, homeTeam: teamKey(row["HomeTeam"]), awayTeam: teamKey(row["AwayTeam"]),
				homeGoals: homeGoals, awayGoals: awayGoals, market: market, hasMarket: hasMarket,
			})
		}
	}
	return history, nil
}

func GetMatchesByPath(ctx context.Context, path string) (columns []string, rows []map[string]string) {
	season := filepath.Base(filepath.Dir(path))
	league := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return GetMatches(ctx, season, league, 0, 0)
}

// teamAliases 常见球队的中文/英文别名映射，目标为 football-data.co.uk 的规范名称
var teamAliases = map[string]string{
	"阿森纳": "arsenal", "arsenal": "arsenal",
	"切尔西": "chelsea", "chelsea": "chelsea",
	"利物浦": "liverpool", "liverpool": "liverpool",
	"曼城": "mancity", "曼彻斯特城": "mancity", "man city": "mancity", "manchester city": "mancity",
	"曼联": "manutd", "曼彻斯特联": "manutd", "man utd": "manutd", "manchester united": "manutd",
	"热刺": "tottenham", "托特纳姆热刺": "tottenham", "tottenham": "tottenham", "tottenham hotspur": "tottenham",
	"纽卡斯尔": "newcastle", "纽卡斯尔联": "newcastle", "newcastle": "newcastle", "newcastle united": "newcastle",
	"阿斯顿维拉": "astonvilla", "aston villa": "astonvilla",
	"布莱顿": "brighton", "brighton": "brighton", "brighton & hove albion": "brighton",
	"西汉姆": "westham", "西汉姆联": "westham", "west ham": "westham", "west ham united": "westham",
	"埃弗顿": "everton", "everton": "everton",
	"水晶宫": "crystalpalace", "crystal palace": "crystalpalace",
	"富勒姆": "fulham", "fulham": "fulham",
	"布伦特福德": "brentford", "brentford": "brentford",
	"狼队": "wolves", "伍尔弗汉普顿": "wolves", "wolves": "wolves", "wolverhampton": "wolves", "wolverhampton wanderers": "wolves",
	"伯恩茅斯": "bournemouth", "bournemouth": "bournemouth",
	"诺丁汉森林": "nottmforest", "nottingham forest": "nottmforest", "nott'm forest": "nottmforest",
	"利兹联": "leeds", "leeds": "leeds", "leeds united": "leeds",
	"伯恩利": "burnley", "burnley": "burnley",
	"桑德兰": "sunderland", "sunderland": "sunderland",
	"莱斯特城": "leicester", "leicester": "leicester", "leicester city": "leicester",
	"南安普顿": "southampton", "southampton": "southampton",
	"皇家马德里": "realmadrid", "皇马": "realmadrid", "real madrid": "realmadrid",
	"巴塞罗那": "barcelona", "巴萨": "barcelona", "barcelona": "barcelona",
	"马德里竞技": "athmadrid", "马竞": "athmadrid", "atletico madrid": "athmadrid", "atlético madrid": "athmadrid", "ath madrid": "athmadrid",
	"塞维利亚": "sevilla", "sevilla": "sevilla",
	"皇家社会": "sociedad", "real sociedad": "sociedad",
	"比利亚雷亚尔": "villarreal", "villarreal": "villarreal",
	"瓦伦西亚": "valencia", "valencia": "valencia",
	"毕尔巴鄂竞技": "athletic", "毕尔巴鄂": "athletic", "athletic bilbao": "athletic", "athletic": "athletic",
	"皇家贝蒂斯": "betis", "贝蒂斯": "betis", "real betis": "betis", "betis": "betis",
	"拜仁慕尼黑": "bayernmunich", "拜仁": "bayernmunich", "bayern munich": "bayernmunich", "bayern": "bayernmunich",
	"多特蒙德": "dortmund", "多特": "dortmund", "borussia dortmund": "dortmund", "dortmund": "dortmund",
	"勒沃库森": "leverkusen", "bayer leverkusen": "leverkusen", "leverkusen": "leverkusen",
	"莱比锡红牛": "leipzig", "莱比锡": "leipzig", "rb leipzig": "leipzig", "leipzig": "leipzig",
	"法兰克福": "frankfurt", "eintracht frankfurt": "frankfurt", "frankfurt": "frankfurt",
	"斯图加特": "stuttgart", "stuttgart": "stuttgart",
	"沃尔夫斯堡": "wolfsburg", "wolfsburg": "wolfsburg",
	"门兴格拉德巴赫": "mgladbach", "门兴": "mgladbach", "mönchengladbach": "mgladbach", "mgladbach": "mgladbach",
	"尤文图斯": "juventus", "尤文": "juventus", "juventus": "juventus",
	"国际米兰": "inter", "国米": "inter", "inter": "inter", "inter milan": "inter", "internazionale": "inter",
	"ac米兰": "milan", "ac米兰队": "milan", "ac milan": "milan", "milan": "milan",
	"那不勒斯": "napoli", "napoli": "napoli",
	"罗马": "roma", "as roma": "roma", "roma": "roma",
	"拉齐奥": "lazio", "ss lazio": "lazio", "lazio": "lazio",
	"亚特兰大": "atalanta", "atalanta": "atalanta",
	"佛罗伦萨": "fiorentina", "fiorentina": "fiorentina",
	"巴黎圣日耳曼": "parissg", "巴黎": "parissg", "paris saint-germain": "parissg", "paris st germain": "parissg", "paris sg": "parissg",
	"马赛": "marseille", "marseille": "marseille",
	"里昂": "lyon", "lyon": "lyon",
	"摩纳哥": "monaco", "monaco": "monaco",
	"里尔": "lille", "lille": "lille",
	"阿贾克斯": "ajax", "ajax": "ajax",
	"埃因霍温": "psv", "psv": "psv", "psv eindhoven": "psv",
	"费耶诺德": "feyenoord", "feyenoord": "feyenoord",
	"波尔图": "porto", "porto": "porto",
	"本菲卡": "benfica", "benfica": "benfica",
	"里斯本竞技": "sporting", "sporting": "sporting", "sporting cp": "sporting",
	"加拉塔萨雷": "galatasaray", "galatasaray": "galatasaray",
	"费内巴切": "fenerbahce", "fenerbahçe": "fenerbahce", "fenerbahce": "fenerbahce",
	"凯尔特人": "celtic", "celtic": "celtic",
	"格拉斯哥流浪者": "rangers", "rangers": "rangers",
}

// teamKey 将中文/英文队名统一成历史数据使用的键；无法识别时保留原始中文避免不同球队互相污染。
func teamKey(value string) string {
	trimmed := strings.TrimSpace(value)
	if key, ok := teamAliases[strings.ToLower(trimmed)]; ok {
		return key
	}
	normalized := normalizeTeamName(trimmed)
	if normalized != "" {
		return normalized
	}
	return "raw:" + strings.ToLower(trimmed)
}

func normalizeTeamName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var normalized strings.Builder
	for _, char := range value {
		if char >= 'a' && char <= 'z' || char >= '0' && char <= '9' {
			normalized.WriteRune(char)
		}
	}
	return normalized.String()
}

func PredictMatch(history *leagueHistory, match v1.ImportMatch) (v1.MatchPrediction, error) {
	day, month, year := parseCsvDate(match.Date)
	if year == 0 || month < 1 || month > 12 || day < 1 || day > 31 {
		return v1.MatchPrediction{}, fmt.Errorf("比赛日期格式无法识别：%s", match.Date)
	}
	cutoff := year*10000 + month*100 + day
	homeKey := teamKey(match.HomeTeam)
	awayKey := teamKey(match.AwayTeam)
	teamStats := make(map[string]*teamGoalStats)
	var homeGoalsTotal, awayGoalsTotal float64
	var historyMatches int

	for _, historical := range history.matches {
		if historical.date >= cutoff {
			continue
		}
		historyMatches++
		homeGoalsTotal += historical.homeGoals
		awayGoalsTotal += historical.awayGoals
		homeStats := teamStats[historical.homeTeam]
		if homeStats == nil {
			homeStats = &teamGoalStats{}
			teamStats[historical.homeTeam] = homeStats
		}
		homeStats.homeGames++
		homeStats.homeGoalsFor += historical.homeGoals
		homeStats.homeGoalsAgainst += historical.awayGoals
		awayStats := teamStats[historical.awayTeam]
		if awayStats == nil {
			awayStats = &teamGoalStats{}
			teamStats[historical.awayTeam] = awayStats
		}
		awayStats.awayGames++
		awayStats.awayGoalsFor += historical.awayGoals
		awayStats.awayGoalsAgainst += historical.homeGoals
	}

	leagueHomeMean := 1.45
	leagueAwayMean := 1.15
	if historyMatches > 0 {
		leagueHomeMean = homeGoalsTotal / float64(historyMatches)
		leagueAwayMean = awayGoalsTotal / float64(historyMatches)
	}
	priorGames := 8.0
	homeStats := teamStats[homeKey]
	awayStats := teamStats[awayKey]
	attackHome := shrinkRate(homeStats, true, true, leagueHomeMean, priorGames)
	defenseAway := shrinkRate(awayStats, false, false, leagueHomeMean, priorGames)
	attackAway := shrinkRate(awayStats, false, true, leagueAwayMean, priorGames)
	defenseHome := shrinkRate(homeStats, true, false, leagueAwayMean, priorGames)
	expectedHome := clamp(leagueHomeMean*attackHome*defenseAway, 0.15, 4.5)
	expectedAway := clamp(leagueAwayMean*attackAway*defenseHome, 0.1, 4.0)

	homeProbabilities := poissonProbabilities(expectedHome, 10)
	awayProbabilities := poissonProbabilities(expectedAway, 10)
	var homeWin, draw, awayWin, over25, total float64
	bestProbability := -1.0
	bestScore := "0 - 0"
	for homeGoals, homeProbability := range homeProbabilities {
		for awayGoals, awayProbability := range awayProbabilities {
			probability := homeProbability * awayProbability
			total += probability
			if homeGoals > awayGoals {
				homeWin += probability
			} else if homeGoals == awayGoals {
				draw += probability
			} else {
				awayWin += probability
			}
			if homeGoals+awayGoals > 2 {
				over25 += probability
			}
			if probability > bestProbability {
				bestProbability = probability
				bestScore = fmt.Sprintf("%d - %d", homeGoals, awayGoals)
			}
		}
	}
	if total > 0 {
		homeWin /= total
		draw /= total
		awayWin /= total
		over25 /= total
	}

	modelBasis := fmt.Sprintf("%s 联赛历史 %d 场", history.code, historyMatches)
	homeTeamHistory := teamHistoryCount(homeStats)
	awayTeamHistory := teamHistoryCount(awayStats)
	if history.code == "" {
		modelBasis = "未匹配到联赛历史，使用通用进球先验"
	} else if historyMatches == 0 {
		modelBasis = fmt.Sprintf("%s 暂无赛前历史样本，使用通用进球先验", history.code)
	} else if homeStats == nil || awayStats == nil {
		modelBasis += "；球队名未匹配，按联赛基准估算"
	} else {
		modelBasis += fmt.Sprintf("；主队样本 %d 场，客队样本 %d 场", homeTeamHistory, awayTeamHistory)
	}

	return v1.MatchPrediction{
		MatchDate: match.Date, KickoffTime: match.Time, League: match.League, HomeTeam: match.HomeTeam, AwayTeam: match.AwayTeam,
		PredictedScore: bestScore, ExpectedHomeGoals: round(expectedHome, 2), ExpectedAwayGoals: round(expectedAway, 2),
		HomeWinProbability: round(homeWin*100, 1), DrawProbability: round(draw*100, 1), AwayWinProbability: round(awayWin*100, 1),
		Over25Probability: round(over25*100, 1), HistoryMatches: historyMatches,
		HomeTeamHistory: homeTeamHistory, AwayTeamHistory: awayTeamHistory,
		ModelBasis: modelBasis,
	}, nil
}

func teamHistoryCount(stats *teamGoalStats) int {
	if stats == nil {
		return 0
	}
	return stats.homeGames + stats.awayGames
}

func shrinkRate(stats *teamGoalStats, atHome, scoring bool, prior, priorGames float64) float64 {
	if stats == nil {
		return 1
	}
	games := stats.awayGames
	total := stats.awayGoalsAgainst
	if scoring {
		total = stats.awayGoalsFor
	}
	if atHome {
		games = stats.homeGames
		total = stats.homeGoalsAgainst
		if scoring {
			total = stats.homeGoalsFor
		}
	}
	if games == 0 || prior <= 0 {
		return 1
	}
	mean := (total + prior*priorGames) / (float64(games) + priorGames)
	return mean / prior
}

func poissonProbabilities(lambda float64, limit int) []float64 {
	probabilities := make([]float64, limit+1)
	probabilities[0] = math.Exp(-lambda)
	for goals := 1; goals <= limit; goals++ {
		probabilities[goals] = probabilities[goals-1] * lambda / float64(goals)
	}
	return probabilities
}

func clamp(value, minimum, maximum float64) float64 {
	return math.Max(minimum, math.Min(maximum, value))
}

func round(value float64, places int) float64 {
	power := math.Pow10(places)
	return math.Round(value*power) / power
}
