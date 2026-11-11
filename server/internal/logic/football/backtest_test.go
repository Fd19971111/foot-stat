package football

import "testing"

func TestMarketProbabilitiesFromOddsRemovesOverround(t *testing.T) {
	probabilities, ok := marketProbabilitiesFromOdds(map[string]string{
		"AvgH": "2.0",
		"AvgD": "3.0",
		"AvgA": "4.0",
	})
	if !ok {
		t.Fatal("expected valid market odds")
	}
	total := probabilities.home + probabilities.draw + probabilities.away
	if total < 0.999999 || total > 1.000001 {
		t.Fatalf("market probabilities sum to %.6f, want 1", total)
	}
	if probabilities.home <= probabilities.draw || probabilities.draw <= probabilities.away {
		t.Fatalf("unexpected normalized market probabilities: %+v", probabilities)
	}
}

func TestEvaluateBacktestUsesEarlierDatesOnly(t *testing.T) {
	history := &leagueHistory{code: "E0", matches: []historicalMatch{
		{date: 20250101, homeTeam: "arsenal", awayTeam: "chelsea", homeGoals: 2, awayGoals: 0},
		{date: 20250101, homeTeam: "liverpool", awayTeam: "wolves", homeGoals: 1, awayGoals: 1},
		{date: 20250108, homeTeam: "arsenal", awayTeam: "chelsea", homeGoals: 0, awayGoals: 1, market: probabilitySet{home: 0.5, draw: 0.3, away: 0.2}, hasMarket: true},
	}}

	report := evaluateBacktest(history, 2025)
	if report.Model.Matches != 3 {
		t.Fatalf("model evaluated %d matches, want 3", report.Model.Matches)
	}
	if report.Market.Matches != 1 {
		t.Fatalf("market evaluated %d matches, want 1", report.Market.Matches)
	}
	if report.Model.LogLoss <= 0 || report.Model.Brier <= 0 {
		t.Fatalf("expected positive scoring losses, got %+v", report.Model)
	}
}

func TestMarketProbabilitiesPreferClosingOdds(t *testing.T) {
	probabilities, ok := marketProbabilitiesFromOdds(map[string]string{
		"AvgCH": "1.5",
		"AvgCD": "4.0",
		"AvgCA": "6.0",
		"AvgH":  "6.0",
		"AvgD":  "4.0",
		"AvgA":  "1.5",
	})
	if !ok || probabilities.home <= probabilities.away {
		t.Fatalf("expected closing odds to take priority, got %+v (valid: %t)", probabilities, ok)
	}
}
