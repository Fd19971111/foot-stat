package football

import (
	"strings"
	"testing"
)

func TestParseSportteryMatches(t *testing.T) {
	body := []byte(`{
		"success": true,
		"value": {
			"matchInfoList": [{
				"businessDate": "2026-11-11",
				"subMatchList": [
					{"matchId": 1, "matchDate": "2026-11-11", "matchTime": "18:00:00", "matchNumStr": "周三001",
					 "leagueAbbName": "日联赛杯", "homeTeamAbbName": "横滨水手", "awayTeamAbbName": "鹿岛鹿角", "matchStatus": "Selling"},
					{"matchId": 1, "matchDate": "2026-11-11", "matchTime": "18:00:00", "matchNumStr": "周三001",
					 "leagueAbbName": "日联赛杯", "homeTeamAbbName": "横滨水手", "awayTeamAbbName": "鹿岛鹿角", "matchStatus": "Selling"},
					{"matchId": 2, "matchDate": "2026-11-11", "matchTime": "20:00:00",
					 "leagueAbbName": "英超", "homeTeamAbbName": "阿森纳", "awayTeamAbbName": "切尔西", "matchStatus": "Ended"},
					{"matchId": 3, "matchDate": "2026-11-12", "matchTime": "03:00:00",
					 "leagueAbbName": "英锦标赛", "homeTeamAbbName": "谢周三", "awayTeamAbbName": "阿克宁顿", "matchStatus": "Selling"}
				]
			}]
		}
	}`)

	matches, err := parseSportteryMatches(body)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 available matches, got %d", len(matches))
	}
	if matches[0].MatchId != "1" || matches[0].Time != "18:00" || matches[0].HomeTeam != "横滨水手" {
		t.Fatalf("unexpected first match: %+v", matches[0])
	}
	if matches[1].Date != "2026-11-12" {
		t.Fatalf("expected late-night match date 2026-11-12, got %s", matches[1].Date)
	}
}

func TestParseSportteryMatchesReportsApiError(t *testing.T) {
	_, err := parseSportteryMatches([]byte(`{"success":false,"errorMessage":"处理失败"}`))
	if err == nil || !strings.Contains(err.Error(), "处理失败") {
		t.Fatalf("expected api error, got %v", err)
	}
}

func TestNormalizeKickoff(t *testing.T) {
	cases := map[string]string{"18:00:00": "18:00", " 03:00 ": "03:00", "": ""}
	for input, want := range cases {
		if got := normalizeKickoff(input); got != want {
			t.Fatalf("normalizeKickoff(%q) = %q, want %q", input, got, want)
		}
	}
}
