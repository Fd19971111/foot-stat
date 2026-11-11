package football

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"

	v1 "server/api/football/v1"
)

type BacktestScores struct {
	Matches  int     `json:"matches"`
	Correct  int     `json:"correct"`
	Accuracy float64 `json:"accuracy"`
	LogLoss  float64 `json:"logLoss"`
	Brier    float64 `json:"brier"`
}

type BacktestReport struct {
	League   string         `json:"league"`
	FromYear int            `json:"fromYear"`
	Model    BacktestScores `json:"model"`
	Market   BacktestScores `json:"market"`
}

func BacktestLeague(ctx context.Context, league string, fromYear int) (*BacktestReport, error) {
	history, err := loadLeagueHistory(ctx, league)
	if err != nil {
		return nil, err
	}
	if history.code == "" {
		return nil, fmt.Errorf("无法识别联赛：%s", league)
	}

	report := evaluateBacktest(history, fromYear)
	if report.Model.Matches == 0 {
		return nil, fmt.Errorf("%d 年以后没有可回测的有效比赛", fromYear)
	}
	return &report, nil
}

func evaluateBacktest(history *leagueHistory, fromYear int) BacktestReport {
	sort.Slice(history.matches, func(i, j int) bool {
		return history.matches[i].date < history.matches[j].date
	})

	report := BacktestReport{League: history.code, FromYear: fromYear}
	var modelLogLoss, modelBrier, marketLogLoss, marketBrier float64
	for start := 0; start < len(history.matches); {
		end := start + 1
		date := history.matches[start].date
		for end < len(history.matches) && history.matches[end].date == date {
			end++
		}

		if date/10000 >= fromYear {
			snapshot := buildLeagueSnapshot(&leagueHistory{code: history.code, matches: history.matches[:start]}, date)
			for _, match := range history.matches[start:end] {
				prediction := analyzeMatch(snapshot, v1.ImportMatch{
					Date:     fmt.Sprintf("%04d-%02d-%02d", date/10000, date/100%100, date%100),
					League:   history.code,
					HomeTeam: match.homeTeam,
					AwayTeam: match.awayTeam,
				})
				actual := outcomeIndex(match.homeGoals, match.awayGoals)
				model := probabilitySet{
					home: prediction.Final.HomeWin / 100,
					draw: prediction.Final.Draw / 100,
					away: prediction.Final.AwayWin / 100,
				}
				addBacktestScore(&report.Model, &modelLogLoss, &modelBrier, model, actual)
				if match.hasMarket {
					addBacktestScore(&report.Market, &marketLogLoss, &marketBrier, match.market, actual)
				}
			}
		}
		start = end
	}

	finishBacktestScores(&report.Model, modelLogLoss, modelBrier)
	finishBacktestScores(&report.Market, marketLogLoss, marketBrier)
	return report
}

func outcomeIndex(homeGoals, awayGoals float64) int {
	if homeGoals > awayGoals {
		return 0
	}
	if homeGoals == awayGoals {
		return 1
	}
	return 2
}

func addBacktestScore(scores *BacktestScores, logLoss, brier *float64, probabilities probabilitySet, actual int) {
	values := [3]float64{probabilities.home, probabilities.draw, probabilities.away}
	best := 0
	for index := 1; index < len(values); index++ {
		if values[index] > values[best] {
			best = index
		}
	}
	if best == actual {
		scores.Correct++
	}
	for index, probability := range values {
		target := 0.0
		if index == actual {
			target = 1
			*logLoss -= math.Log(math.Max(probability, 1e-15))
		}
		difference := probability - target
		*brier += difference * difference
	}
	scores.Matches++
}

func finishBacktestScores(scores *BacktestScores, logLoss, brier float64) {
	if scores.Matches == 0 {
		return
	}
	count := float64(scores.Matches)
	scores.Accuracy = round(float64(scores.Correct)/count*100, 2)
	scores.LogLoss = round(logLoss/count, 4)
	scores.Brier = round(brier/count, 4)
}

func marketProbabilitiesFromOdds(row map[string]string) (probabilitySet, bool) {
	for _, columns := range [][3]string{
		{"AvgCH", "AvgCD", "AvgCA"},
		{"B365CH", "B365CD", "B365CA"},
		{"AvgH", "AvgD", "AvgA"},
		{"B365H", "B365D", "B365A"},
	} {
		var inverse [3]float64
		valid := true
		for index, column := range columns {
			odds, err := strconv.ParseFloat(row[column], 64)
			if err != nil || odds <= 1 || math.IsNaN(odds) || math.IsInf(odds, 0) {
				valid = false
				break
			}
			inverse[index] = 1 / odds
		}
		if valid {
			return normalize(probabilitySet{home: inverse[0], draw: inverse[1], away: inverse[2]}), true
		}
	}
	return probabilitySet{}, false
}
