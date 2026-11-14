package football

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	v1 "server/api/football/v1"
)

const sportteryMatchURL = "https://webapi.sporttery.cn/gateway/jc/football/getMatchCalculatorV1.qry?poolCode=hhad,had&channel=c"

type sportteryResponse struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"errorMessage"`
	Value        struct {
		MatchInfoList []sportteryMatchGroup `json:"matchInfoList"`
	} `json:"value"`
}

type sportteryMatchGroup struct {
	BusinessDate string           `json:"businessDate"`
	SubMatchList []sportteryMatch `json:"subMatchList"`
}

type sportteryOdds struct {
	H string `json:"h"`
	D string `json:"d"`
	A string `json:"a"`
}

type sportteryMatch struct {
	BusinessDate    string         `json:"businessDate"`
	MatchDate       string         `json:"matchDate"`
	MatchTime       string         `json:"matchTime"`
	MatchNumStr     string         `json:"matchNumStr"`
	LeagueAbbName   string         `json:"leagueAbbName"`
	LeagueAllName   string         `json:"leagueAllName"`
	HomeTeamAbbName string         `json:"homeTeamAbbName"`
	HomeTeamAllName string         `json:"homeTeamAllName"`
	AwayTeamAbbName string         `json:"awayTeamAbbName"`
	AwayTeamAllName string         `json:"awayTeamAllName"`
	MatchStatus     string         `json:"matchStatus"`
	MatchId         int64          `json:"matchId"`
	Had             sportteryOdds  `json:"had"`
	OddsList        []sportteryOdd `json:"oddsList"`
}

type sportteryOdd struct {
	H        string `json:"h"`
	D        string `json:"d"`
	A        string `json:"a"`
	PoolCode string `json:"poolCode"`
}

// parseSportteryOdds 读取竞彩网「胜平负」(HAD) 赔率；优先取 had 字段，缺失时回退到 oddsList。
func parseSportteryOdds(item sportteryMatch) (home, draw, away float64) {
	candidate := item.Had
	if parseSportteryOdd(candidate.H) <= 1 {
		for _, odd := range item.OddsList {
			if strings.EqualFold(odd.PoolCode, "HAD") && parseSportteryOdd(odd.H) > 1 {
				candidate = sportteryOdds{H: odd.H, D: odd.D, A: odd.A}
				break
			}
		}
	}
	home, draw, away = parseSportteryOdd(candidate.H), parseSportteryOdd(candidate.D), parseSportteryOdd(candidate.A)
	if home <= 1 || draw <= 1 || away <= 1 {
		return 0, 0, 0
	}
	return home, draw, away
}

func parseSportteryOdd(value string) float64 {
	odds, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || odds <= 1 || odds > 1000 {
		return 0
	}
	return odds
}

// GetRemoteMatches 拉取中国竞彩网当前在售的足球赛程。
func GetRemoteMatches(ctx context.Context) (*v1.GetRemoteMatchesRes, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, sportteryMatchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建竞彩网请求失败: %w", err)
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0 Safari/537.36")
	request.Header.Set("Referer", "https://www.sporttery.cn/")
	request.Header.Set("Accept", "application/json, text/plain, */*")

	client := &http.Client{Timeout: 20 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("连接中国竞彩网失败，请检查网络: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 5<<20))
	if err != nil {
		return nil, fmt.Errorf("读取竞彩网数据失败: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("竞彩网接口返回 HTTP %d", response.StatusCode)
	}

	matches, err := parseSportteryMatches(body)
	if err != nil {
		return nil, err
	}
	return &v1.GetRemoteMatchesRes{
		Source:    "中国竞彩网",
		FetchedAt: time.Now().Format("2006-01-02 15:04:05"),
		Matches:   matches,
	}, nil
}

func parseSportteryMatches(body []byte) ([]v1.RemoteMatch, error) {
	var payload sportteryResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("解析竞彩网数据失败: %w", err)
	}
	if !payload.Success {
		message := strings.TrimSpace(payload.ErrorMessage)
		if message == "" {
			message = "未知错误"
		}
		return nil, fmt.Errorf("竞彩网接口返回错误：%s", message)
	}

	seen := make(map[string]bool)
	matches := make([]v1.RemoteMatch, 0)
	for _, group := range payload.Value.MatchInfoList {
		for _, item := range group.SubMatchList {
			date := firstNonEmpty(item.MatchDate, item.BusinessDate, group.BusinessDate)
			league := firstNonEmpty(item.LeagueAbbName, item.LeagueAllName)
			homeTeam := firstNonEmpty(item.HomeTeamAbbName, item.HomeTeamAllName)
			awayTeam := firstNonEmpty(item.AwayTeamAbbName, item.AwayTeamAllName)
			if date == "" || league == "" || homeTeam == "" || awayTeam == "" || homeTeam == awayTeam {
				continue
			}
			if isFinishedMatch(item.MatchStatus) {
				continue
			}

			matchId := strconv.FormatInt(item.MatchId, 10)
			key := matchId
			if matchId == "0" {
				key = date + "|" + item.MatchTime + "|" + homeTeam + "|" + awayTeam
			}
			if seen[key] {
				continue
			}
			seen[key] = true

			homeOdds, drawOdds, awayOdds := parseSportteryOdds(item)
			matches = append(matches, v1.RemoteMatch{
				MatchId:  matchId,
				MatchNum: strings.TrimSpace(item.MatchNumStr),
				Date:     date,
				Time:     normalizeKickoff(item.MatchTime),
				League:   league,
				HomeTeam: homeTeam,
				AwayTeam: awayTeam,
				Status:   item.MatchStatus,
				HomeOdds: homeOdds,
				DrawOdds: drawOdds,
				AwayOdds: awayOdds,
			})
		}
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Date == matches[j].Date {
			return matches[i].Time < matches[j].Time
		}
		return matches[i].Date < matches[j].Date
	})
	return matches, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func normalizeKickoff(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 5 {
		return value[:5]
	}
	return value
}

func isFinishedMatch(status string) bool {
	status = strings.ToLower(status)
	return strings.Contains(status, "end") || strings.Contains(status, "cancel") || strings.Contains(status, "result")
}
