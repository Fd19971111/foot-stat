package football

import (
	"context"

	v1 "server/api/football/v1"
	"server/internal/logic/football"
)

type ControllerV1 struct{}

func NewV1() v1.IFootballV1 {
	return &ControllerV1{}
}

func (c *ControllerV1) GetSeasons(ctx context.Context, _ *v1.GetSeasonsReq) (*v1.GetSeasonsRes, error) {
	return &v1.GetSeasonsRes{Seasons: football.GetSeasons(ctx)}, nil
}

func (c *ControllerV1) GetLeagues(ctx context.Context, req *v1.GetLeaguesReq) (*v1.GetLeaguesRes, error) {
	return &v1.GetLeaguesRes{Leagues: football.GetLeagues(ctx, req.Season)}, nil
}

func (c *ControllerV1) GetMatches(ctx context.Context, req *v1.GetMatchesReq) (*v1.GetMatchesRes, error) {
	columns, rows := football.GetMatches(ctx, req.Season, req.League, req.Year, req.Month)
	return &v1.GetMatchesRes{Columns: columns, Rows: rows}, nil
}

func (c *ControllerV1) CreateImport(ctx context.Context, req *v1.CreateImportReq) (*v1.CreateImportRes, error) {
	return football.CreateImport(ctx, req)
}

func (c *ControllerV1) GetImports(ctx context.Context, _ *v1.GetImportsReq) (*v1.GetImportsRes, error) {
	batches, err := football.GetImports(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetImportsRes{Batches: batches}, nil
}

func (c *ControllerV1) GetImportDetails(ctx context.Context, req *v1.GetImportDetailsReq) (*v1.GetImportDetailsRes, error) {
	return football.GetImportDetails(ctx, req.Id)
}

func (c *ControllerV1) GetAnalysis(ctx context.Context, req *v1.GetAnalysisReq) (*v1.GetAnalysisRes, error) {
	return football.AnalyzeImport(ctx, req.Id)
}

func (c *ControllerV1) GetRemoteMatches(ctx context.Context, req *v1.GetRemoteMatchesReq) (*v1.GetRemoteMatchesRes, error) {
	return football.GetRemoteMatches(ctx)
}
