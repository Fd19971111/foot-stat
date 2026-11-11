package v1

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// IFootballV1 reads match data from CSV files and stores user-uploaded import batches.
type IFootballV1 interface {
	GetSeasons(ctx context.Context, req *GetSeasonsReq) (res *GetSeasonsRes, err error)
	GetLeagues(ctx context.Context, req *GetLeaguesReq) (res *GetLeaguesRes, err error)
	GetMatches(ctx context.Context, req *GetMatchesReq) (res *GetMatchesRes, err error)
	CreateImport(ctx context.Context, req *CreateImportReq) (res *CreateImportRes, err error)
	GetImports(ctx context.Context, req *GetImportsReq) (res *GetImportsRes, err error)
	GetImportDetails(ctx context.Context, req *GetImportDetailsReq) (res *GetImportDetailsRes, err error)
	GetAnalysis(ctx context.Context, req *GetAnalysisReq) (res *GetAnalysisRes, err error)
	GetRemoteMatches(ctx context.Context, req *GetRemoteMatchesReq) (res *GetRemoteMatchesRes, err error)
}

type GetSeasonsReq struct {
	g.Meta `path:"/football/seasons" method:"get" tags:"Football" summary:"获取可用的赛季列表"`
}

type SeasonItem struct {
	Value   string `json:"value"`   // 文件夹名，如 2526
	Label   string `json:"label"`   // 如 2025-2026
	Current bool   `json:"current"` // 是否为当前赛季
}

type GetSeasonsRes struct {
	Seasons []SeasonItem `json:"seasons"`
}

type GetLeaguesReq struct {
	g.Meta `path:"/football/leagues" method:"get" tags:"Football" summary:"获取某赛季下的联赛列表"`
	Season string `json:"season" v:"required#请选择赛季"`
}

type LeagueItem struct {
	Value string `json:"value"` // 文件 code，如 E0
	Label string `json:"label"` // 如 E0 - 英格兰超级联赛
}

type GetLeaguesRes struct {
	Leagues []LeagueItem `json:"leagues"`
}

type GetMatchesReq struct {
	g.Meta `path:"/football/matches" method:"get" tags:"Football" summary:"获取某赛季某联赛某月的比赛数据"`
	Season string `json:"season" v:"required#请选择赛季"`
	League string `json:"league" v:"required#请选择联赛"`
	Year   int    `json:"year"`  // 筛选年份，0 表示不过滤
	Month  int    `json:"month"` // 筛选月份 1-12，0 表示不过滤
}

type GetMatchesRes struct {
	Columns []string            `json:"columns"` // CSV 表头顺序
	Rows    []map[string]string `json:"rows"`    // 每一场比赛的全部字段
}

type ImportMatch struct {
	Date     string `json:"date" v:"required#比赛日期不能为空"`
	Time     string `json:"time"`
	League   string `json:"league" v:"required#联赛不能为空"`
	HomeTeam string `json:"homeTeam" v:"required#主队不能为空"`
	AwayTeam string `json:"awayTeam" v:"required#客队不能为空"`
}

type GetRemoteMatchesReq struct {
	g.Meta `path:"/football/remote-matches" method:"get" tags:"Football" summary:"拉取中国竞彩网在售比赛"`
}

type RemoteMatch struct {
	MatchId  string `json:"matchId"`  // 竞彩网比赛 ID，可用于去重
	MatchNum string `json:"matchNum"` // 竞彩编号，如 周三001
	Date     string `json:"date"`     // 比赛日期
	Time     string `json:"time"`     // 开赛时间
	League   string `json:"league"`
	HomeTeam string `json:"homeTeam"`
	AwayTeam string `json:"awayTeam"`
	Status   string `json:"status"`
}

type GetRemoteMatchesRes struct {
	Source    string        `json:"source"`
	FetchedAt string        `json:"fetchedAt"`
	Matches   []RemoteMatch `json:"matches"`
}

type CreateImportReq struct {
	g.Meta   `path:"/football/imports" method:"post" tags:"Football" summary:"保存导入的比赛"`
	FileName string        `json:"fileName" v:"required#文件名不能为空"`
	Matches  []ImportMatch `json:"matches" v:"required#比赛数据不能为空"`
}

type ImportBatch struct {
	Id         int64  `json:"id"`
	FileName   string `json:"fileName"`
	MatchCount int    `json:"matchCount"`
	ImportedAt string `json:"importedAt"`
}

type MatchPrediction struct {
	MatchDate          string  `json:"matchDate"`
	KickoffTime        string  `json:"kickoffTime"`
	League             string  `json:"league"`
	HomeTeam           string  `json:"homeTeam"`
	AwayTeam           string  `json:"awayTeam"`
	PredictedScore     string  `json:"predictedScore"`
	ExpectedHomeGoals  float64 `json:"expectedHomeGoals"`
	ExpectedAwayGoals  float64 `json:"expectedAwayGoals"`
	HomeWinProbability float64 `json:"homeWinProbability"`
	DrawProbability    float64 `json:"drawProbability"`
	AwayWinProbability float64 `json:"awayWinProbability"`
	Over25Probability  float64 `json:"over25Probability"`
	HistoryMatches     int     `json:"historyMatches"`
	HomeTeamHistory    int     `json:"homeTeamHistory"`
	AwayTeamHistory    int     `json:"awayTeamHistory"`
	ModelBasis         string  `json:"modelBasis"`
}

type CreateImportRes struct {
	Batch       ImportBatch       `json:"batch"`
	Predictions []MatchPrediction `json:"predictions"`
}

type GetImportsReq struct {
	g.Meta `path:"/football/imports" method:"get" tags:"Football" summary:"获取导入记录"`
}

type GetImportsRes struct {
	Batches []ImportBatch `json:"batches"`
}

type GetImportDetailsReq struct {
	g.Meta `path:"/football/imports/{id}" method:"get" tags:"Football" summary:"获取导入批次及预测结果"`
	Id     int64 `json:"id" v:"required|min:1#导入批次无效"`
}

type GetImportDetailsRes struct {
	Batch       ImportBatch       `json:"batch"`
	Predictions []MatchPrediction `json:"predictions"`
}

type GetAnalysisReq struct {
	g.Meta `path:"/football/analysis/{id}" method:"get" tags:"Football" summary:"用常用分析算法计算导入批次的比赛结果"`
	Id     int64 `json:"id" v:"required|min:1#导入批次无效"`
}

type AnalysisProbability struct {
	HomeWin float64 `json:"homeWin"` // 主胜概率，百分比
	Draw    float64 `json:"draw"`    // 平局概率，百分比
	AwayWin float64 `json:"awayWin"` // 客胜概率，百分比
}

type AnalysisTeamForm struct {
	Matches      int     `json:"matches"`      // 统计场次（最近 5 场）
	Wins         int     `json:"wins"`         // 胜场
	Draws        int     `json:"draws"`        // 平局
	Losses       int     `json:"losses"`       // 负场
	Points       int     `json:"points"`       // 积分
	GoalsFor     int     `json:"goalsFor"`     // 进球
	GoalsAgainst int     `json:"goalsAgainst"` // 失球
	Sequence     string  `json:"sequence"`     // 由旧到新的赛果，如 WWDLW
	Elo          float64 `json:"elo"`          // Elo 等级分
}

type AnalysisScore struct {
	Score       string  `json:"score"`       // 比分，如 2 - 1
	Probability float64 `json:"probability"` // 出现概率，百分比
}

type AnalysisOdds struct {
	Home float64 `json:"home"` // 主胜公平赔率
	Draw float64 `json:"draw"` // 平局公平赔率
	Away float64 `json:"away"` // 客胜公平赔率
}

type MatchAnalysis struct {
	MatchDate          string              `json:"matchDate"`
	KickoffTime        string              `json:"kickoffTime"`
	League             string              `json:"league"`
	HomeTeam           string              `json:"homeTeam"`
	AwayTeam           string              `json:"awayTeam"`
	ExpectedHomeGoals  float64             `json:"expectedHomeGoals"`
	ExpectedAwayGoals  float64             `json:"expectedAwayGoals"`
	Poisson            AnalysisProbability `json:"poisson"`
	Elo                AnalysisProbability `json:"elo"`
	Final              AnalysisProbability `json:"final"`
	Over25Probability  float64             `json:"over25Probability"`
	BttsProbability    float64             `json:"bttsProbability"`
	FairOdds           AnalysisOdds        `json:"fairOdds"`
	PredictedScore     string              `json:"predictedScore"`
	Recommendation     string              `json:"recommendation"`
	RecommendationType string              `json:"recommendationType"` // home / draw / away
	GoalsAdvice        string              `json:"goalsAdvice"`
	GoalsType          string              `json:"goalsType"` // over / under
	BttsAdvice         string              `json:"bttsAdvice"`
	Confidence         float64             `json:"confidence"`
	ConfidenceLevel    string              `json:"confidenceLevel"` // 高 / 中 / 低
	HistoryMatches     int                 `json:"historyMatches"`
	HomeTeamHistory    int                 `json:"homeTeamHistory"`
	AwayTeamHistory    int                 `json:"awayTeamHistory"`
	HomeForm           AnalysisTeamForm    `json:"homeForm"`
	AwayForm           AnalysisTeamForm    `json:"awayForm"`
	ScoreLines         []AnalysisScore     `json:"scoreLines"`
	ModelBasis         string              `json:"modelBasis"`
}

type AnalysisSummary struct {
	MatchCount        int     `json:"matchCount"`
	HighConfidence    int     `json:"highConfidence"`
	HomeLean          int     `json:"homeLean"`
	OverLean          int     `json:"overLean"`
	AverageConfidence float64 `json:"averageConfidence"`
	Algorithm         string  `json:"algorithm"`
}

type GetAnalysisRes struct {
	Batch   ImportBatch     `json:"batch"`
	Summary AnalysisSummary `json:"summary"`
	Matches []MatchAnalysis `json:"matches"`
}
