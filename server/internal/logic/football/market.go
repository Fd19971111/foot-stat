package football

import (
	"context"
	"math"
	"strings"
	"time"

	v1 "server/api/football/v1"
)

// enrichOddsFromRemote 为缺少赔率的历史批次按「日期 + 主队 + 客队」补回竞彩网在售赔率。
// 返回被补齐的比赛下标，便于调用方把赔率持久化；拉取失败时静默降级，不影响原有流程。
func enrichOddsFromRemote(ctx context.Context, matches []v1.ImportMatch) []int {
	missing := false
	for _, match := range matches {
		if match.HomeOdds <= 1 || match.DrawOdds <= 1 || match.AwayOdds <= 1 {
			missing = true
			break
		}
	}
	if !missing {
		return nil
	}

	remoteCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	remote, err := GetRemoteMatches(remoteCtx)
	if err != nil || remote == nil || len(remote.Matches) == 0 {
		return nil
	}

	byFixture := make(map[string]v1.RemoteMatch, len(remote.Matches))
	for _, item := range remote.Matches {
		if item.HomeOdds <= 1 || item.DrawOdds <= 1 || item.AwayOdds <= 1 {
			continue
		}
		byFixture[fixtureKey(item.Date, item.HomeTeam, item.AwayTeam)] = item
	}
	if len(byFixture) == 0 {
		return nil
	}
	filled := make([]int, 0)
	for index := range matches {
		match := &matches[index]
		if match.HomeOdds > 1 && match.DrawOdds > 1 && match.AwayOdds > 1 {
			continue
		}
		if item, ok := byFixture[fixtureKey(match.Date, match.HomeTeam, match.AwayTeam)]; ok {
			match.HomeOdds, match.DrawOdds, match.AwayOdds = item.HomeOdds, item.DrawOdds, item.AwayOdds
			filled = append(filled, index)
		}
	}
	return filled
}

func fixtureKey(date, homeTeam, awayTeam string) string {
	return strings.ToLower(strings.TrimSpace(date) + "|" + strings.TrimSpace(homeTeam) + "|" + strings.TrimSpace(awayTeam))
}

// marketProbabilitySet 把胜平负赔率去掉水位后换算成隐含概率。
// 竞彩网、football-data.co.uk 的历史 CSV 都提供这三个赔率，是球队未匹配历史时最可靠的差异化信号。
func marketProbabilitySet(match v1.ImportMatch) (probabilitySet, bool) {
	return marketProbabilities(match.HomeOdds, match.DrawOdds, match.AwayOdds)
}

func marketProbabilities(homeOdds, drawOdds, awayOdds float64) (probabilitySet, bool) {
	odds := [3]float64{homeOdds, drawOdds, awayOdds}
	var inverse [3]float64
	for index, value := range odds {
		if value <= 1 || math.IsNaN(value) || math.IsInf(value, 0) {
			return probabilitySet{}, false
		}
		inverse[index] = 1 / value
	}
	return normalize(probabilitySet{home: inverse[0], draw: inverse[1], away: inverse[2]}), true
}

// poissonProbabilitySet 由主客队期望进球计算胜平负概率。
func poissonProbabilitySet(homeLambda, awayLambda float64) probabilitySet {
	homeGoals := poissonProbabilities(homeLambda, 10)
	awayGoals := poissonProbabilities(awayLambda, 10)
	var result probabilitySet
	for home, homeProbability := range homeGoals {
		for away, awayProbability := range awayGoals {
			probability := homeProbability * awayProbability
			switch {
			case home > away:
				result.home += probability
			case home == away:
				result.draw += probability
			default:
				result.away += probability
			}
		}
	}
	return normalize(result)
}

// calibrateLambdas 反推一组期望进球，使泊松模型的胜平负概率尽量贴近市场隐含概率。
// baseHome/baseAway 来自历史统计，作为总进球水平的软约束；无历史时只约束在合理区间。
func calibrateLambdas(target probabilitySet, baseHome, baseAway float64) (float64, float64) {
	baseTotal := baseHome + baseAway
	if baseTotal <= 0 {
		baseTotal = 2.6
	}
	errorOf := func(homeLambda, awayLambda float64) float64 {
		probabilities := poissonProbabilitySet(homeLambda, awayLambda)
		diffHome := probabilities.home - target.home
		diffDraw := probabilities.draw - target.draw
		diffAway := probabilities.away - target.away
		goalPenalty := 0.004 * math.Pow(homeLambda+awayLambda-baseTotal, 2)
		return diffHome*diffHome + diffDraw*diffDraw + diffAway*diffAway + goalPenalty
	}

	bestHome, bestAway, bestError := baseHome, baseAway, math.MaxFloat64
	for home := 0.30; home <= 3.60; home += 0.10 {
		for away := 0.20; away <= 3.10; away += 0.10 {
			if value := errorOf(home, away); value < bestError {
				bestError, bestHome, bestAway = value, home, away
			}
		}
	}

	// 在网格最优点附近细化，避免不同比赛被量化成同一组期望进球。
	refinedHome, refinedAway, refinedError := bestHome, bestAway, bestError
	for home := bestHome - 0.10; home <= bestHome+0.101; home += 0.02 {
		if home < 0.15 {
			continue
		}
		for away := bestAway - 0.10; away <= bestAway+0.101; away += 0.02 {
			if away < 0.10 {
				continue
			}
			if value := errorOf(home, away); value < refinedError {
				refinedError, refinedHome, refinedAway = value, home, away
			}
		}
	}
	return refinedHome, refinedAway
}

// marketWeight 历史样本越少，越依赖市场赔率；样本充足时保留历史模型的差异。
func marketWeight(historyMatches int) float64 {
	if historyMatches <= 0 {
		return 1
	}
	return clamp(1-float64(historyMatches)/float64(historyMatches+150), 0.35, 0.85)
}
