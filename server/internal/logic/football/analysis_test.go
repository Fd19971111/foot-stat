package football

import (
	"math"
	"testing"

	v1 "server/api/football/v1"
)

func TestTeamKeyResolvesCommonAliases(t *testing.T) {
	cases := map[string]string{
		"阿森纳":             "arsenal",
		"Arsenal":         "arsenal",
		"曼城":              "mancity",
		"Manchester City": "mancity",
		"Man City":        "mancity",
		"皇家马德里":           "realmadrid",
		"Real Madrid":     "realmadrid",
		"诺丁汉森林":           "nottmforest",
		"Nott'm Forest":   "nottmforest",
	}
	for input, want := range cases {
		if got := teamKey(input); got != want {
			t.Fatalf("teamKey(%q) = %q, want %q", input, got, want)
		}
	}
	if got := teamKey("某支未知球队"); got != "raw:某支未知球队" {
		t.Fatalf("unknown Chinese team key = %q, want raw prefix", got)
	}
}

func TestPoissonOutcomeProducesUsableDistribution(t *testing.T) {
	result := poissonOutcome(1.65, 1.05)
	total := result.home + result.draw + result.away
	if math.Abs(total-1) > 1e-9 {
		t.Fatalf("result probabilities sum to %.6f, want 1", total)
	}
	if result.over25 < 0 || result.over25 > 1 {
		t.Fatalf("over 2.5 probability %.4f out of range", result.over25)
	}
	if result.btts < 0 || result.btts > 1 {
		t.Fatalf("both-teams-to-score probability %.4f out of range", result.btts)
	}
	if len(result.scoreLines) != 5 {
		t.Fatalf("expected top 5 scorelines, got %d", len(result.scoreLines))
	}
	if result.scoreLines[0].Score != result.topScore {
		t.Fatalf("top score %q does not match first scoreline %q", result.topScore, result.scoreLines[0].Score)
	}
	for _, line := range result.scoreLines {
		if line.Probability <= 0 || line.Probability > 100 {
			t.Fatalf("scoreline %s has invalid probability %.2f", line.Score, line.Probability)
		}
	}
}

func TestDixonColesTauKeepsLowScoresPositive(t *testing.T) {
	for _, home := range []int{0, 1} {
		for _, away := range []int{0, 1} {
			if tau := dixonColesTau(home, away, 4.5, 4.0); tau <= 0 {
				t.Fatalf("tau(%d,%d) = %.4f, want positive", home, away, tau)
			}
		}
	}
}

func TestBuildLeagueSnapshotRanksStrongerTeamHigher(t *testing.T) {
	history := &leagueHistory{code: "E0"}
	for index := 0; index < 12; index++ {
		if index%2 == 0 {
			history.matches = append(history.matches, historicalMatch{
				date: 20250101 + index*7, homeTeam: "strong", awayTeam: "weak", homeGoals: 3, awayGoals: 0,
			})
		} else {
			history.matches = append(history.matches, historicalMatch{
				date: 20250101 + index*7, homeTeam: "weak", awayTeam: "strong", homeGoals: 0, awayGoals: 3,
			})
		}
	}

	snapshot := buildLeagueSnapshot(history, 20260101)
	if snapshot.matches != 12 {
		t.Fatalf("snapshot matches = %d, want 12", snapshot.matches)
	}
	if snapshot.ratings["strong"] <= snapshot.ratings["weak"] {
		t.Fatalf("strong team rating %.1f should exceed weak team rating %.1f", snapshot.ratings["strong"], snapshot.ratings["weak"])
	}
	if snapshot.homeMean <= 0 || snapshot.awayMean <= 0 {
		t.Fatalf("league goal means must be positive, got %.2f / %.2f", snapshot.homeMean, snapshot.awayMean)
	}
}

func TestAnalyzeMatchReturnsBalancedResult(t *testing.T) {
	history := &leagueHistory{code: "E0"}
	for index := 0; index < 16; index++ {
		history.matches = append(history.matches, historicalMatch{
			date: 20250101 + index*7, homeTeam: "arsenal", awayTeam: "chelsea",
			homeGoals: 2, awayGoals: 1,
		})
	}
	snapshot := buildLeagueSnapshot(history, 20260101)

	analysis := analyzeMatch(snapshot, v1.ImportMatch{
		Date: "2026-02-01", Time: "20:30", League: "英超", HomeTeam: "Arsenal", AwayTeam: "Chelsea",
	})

	total := analysis.Final.HomeWin + analysis.Final.Draw + analysis.Final.AwayWin
	if math.Abs(total-100) > 0.3 {
		t.Fatalf("final probabilities sum to %.2f, want 100", total)
	}
	if analysis.RecommendationType != "home" && analysis.RecommendationType != "draw" && analysis.RecommendationType != "away" {
		t.Fatalf("unexpected recommendation type %q", analysis.RecommendationType)
	}
	if analysis.Confidence < 0 || analysis.Confidence > 100 {
		t.Fatalf("confidence %.1f out of range", analysis.Confidence)
	}
	if analysis.ConfidenceLevel != "高" && analysis.ConfidenceLevel != "中" && analysis.ConfidenceLevel != "低" {
		t.Fatalf("unexpected confidence level %q", analysis.ConfidenceLevel)
	}
	if analysis.FairOdds.Home < 1 || analysis.FairOdds.Draw < 1 || analysis.FairOdds.Away < 1 {
		t.Fatalf("fair odds should be >= 1, got %+v", analysis.FairOdds)
	}
	if analysis.PredictedScore == "" || len(analysis.ScoreLines) != 5 {
		t.Fatalf("missing scoreline output: %+v", analysis)
	}
	if analysis.HomeForm.Matches != 5 || analysis.AwayForm.Matches != 5 {
		t.Fatalf("expected last-five form, got %d / %d", analysis.HomeForm.Matches, analysis.AwayForm.Matches)
	}
}

func TestAnalyzeMatchFallsBackWithoutHistory(t *testing.T) {
	snapshot := buildLeagueSnapshot(&leagueHistory{}, 20260101)
	analysis := analyzeMatch(snapshot, v1.ImportMatch{
		Date: "2026-02-01", League: "未知联赛", HomeTeam: "阿森纳", AwayTeam: "切尔西",
	})

	if analysis.HistoryMatches != 0 {
		t.Fatalf("expected zero history matches, got %d", analysis.HistoryMatches)
	}
	if analysis.Confidence > 45 {
		t.Fatalf("confidence %.1f should stay conservative without samples", analysis.Confidence)
	}
	if analysis.HomeForm.Matches != 0 || analysis.AwayForm.Matches != 0 {
		t.Fatalf("expected empty form without history")
	}
	if analysis.PredictedScore == "" {
		t.Fatal("expected a fallback scoreline")
	}
}
