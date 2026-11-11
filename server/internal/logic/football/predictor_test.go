package football

import (
	"math"
	"strings"
	"testing"

	v1 "server/api/football/v1"
)

func TestPredictMatchUsesOnlyPreMatchHistory(t *testing.T) {
	history := &leagueHistory{
		code: "E0",
		matches: []historicalMatch{
			{date: 20250801, homeTeam: "arsenal", awayTeam: "wolves", homeGoals: 2, awayGoals: 0},
			{date: 20250901, homeTeam: "chelsea", awayTeam: "liverpool", homeGoals: 1, awayGoals: 2},
			{date: 20261001, homeTeam: "arsenal", awayTeam: "liverpool", homeGoals: 0, awayGoals: 4},
		},
	}

	prediction, err := PredictMatch(history, v1.ImportMatch{
		Date: "2026-02-01", League: "E0", HomeTeam: "Arsenal", AwayTeam: "Liverpool",
	})
	if err != nil {
		t.Fatalf("PredictMatch returned error: %v", err)
	}
	if prediction.HistoryMatches != 2 {
		t.Fatalf("expected future history to be excluded, got %d matches", prediction.HistoryMatches)
	}
	probabilityTotal := prediction.HomeWinProbability + prediction.DrawProbability + prediction.AwayWinProbability
	if math.Abs(probabilityTotal-100) > 0.2 {
		t.Fatalf("result probabilities sum to %.2f, want 100", probabilityTotal)
	}
	if prediction.ExpectedHomeGoals <= 0 || prediction.ExpectedAwayGoals <= 0 {
		t.Fatalf("expected positive goal rates, got %.2f - %.2f", prediction.ExpectedHomeGoals, prediction.ExpectedAwayGoals)
	}
	if prediction.PredictedScore == "" {
		t.Fatal("expected a most-likely score")
	}
}

func TestPredictMatchFallsBackWhenNoLeagueHistoryExists(t *testing.T) {
	prediction, err := PredictMatch(&leagueHistory{}, v1.ImportMatch{
		Date: "2026-11-15", League: "未知联赛", HomeTeam: "阿森纳", AwayTeam: "切尔西",
	})
	if err != nil {
		t.Fatalf("PredictMatch returned error: %v", err)
	}
	if prediction.HistoryMatches != 0 || !strings.Contains(prediction.ModelBasis, "通用进球先验") {
		t.Fatalf("expected fallback note and zero history, got %+v", prediction)
	}
	if prediction.HomeTeamHistory != 0 || prediction.AwayTeamHistory != 0 {
		t.Fatal("unmatched teams should report zero team-history samples")
	}
}

func TestParseCSVDateExpandsTwoDigitYears(t *testing.T) {
	cases := []struct {
		date string
		year int
	}{
		{date: "14/08/93", year: 1993},
		{date: "26/08/00", year: 2000},
		{date: "11/08/01", year: 2001},
		{date: "15/08/2025", year: 2025},
		{date: "2026-02-01", year: 2026},
	}
	for _, testCase := range cases {
		_, _, year := parseCsvDate(testCase.date)
		if year != testCase.year {
			t.Errorf("parseCsvDate(%q) year = %d, want %d", testCase.date, year, testCase.year)
		}
	}
}

func TestResolveLeagueCode(t *testing.T) {
	for input, expected := range map[string]string{
		"E0":                          "E0",
		"E0 - English Premier League": "E0",
		"英超":                          "E0",
		"西班牙甲级联赛":                     "SP1",
	} {
		if actual := resolveLeagueCode(input); actual != expected {
			t.Errorf("resolveLeagueCode(%q) = %q, want %q", input, actual, expected)
		}
	}
}
