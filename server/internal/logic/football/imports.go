package football

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gogf/gf/v2/frame/g"

	v1 "server/api/football/v1"
)

var importSchema = []string{
	`CREATE TABLE IF NOT EXISTS footstat_import_batches (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		file_name VARCHAR(255) NOT NULL,
		match_count INT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY idx_footstat_import_batches_created_at (created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS footstat_import_matches (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		batch_id BIGINT UNSIGNED NOT NULL,
		match_date VARCHAR(20) NOT NULL,
		kickoff_time VARCHAR(10) NOT NULL DEFAULT '',
		league VARCHAR(80) NOT NULL,
		home_team VARCHAR(120) NOT NULL,
		away_team VARCHAR(120) NOT NULL,
		home_odds DECIMAL(8,3) NULL,
		draw_odds DECIMAL(8,3) NULL,
		away_odds DECIMAL(8,3) NULL,
		home_goals TINYINT NULL,
		away_goals TINYINT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		KEY idx_footstat_import_matches_batch_id (batch_id),
		CONSTRAINT fk_footstat_import_matches_batch
			FOREIGN KEY (batch_id) REFERENCES footstat_import_batches (id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS footstat_import_predictions (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		import_match_id BIGINT UNSIGNED NOT NULL,
		predicted_score VARCHAR(20) NOT NULL,
		expected_home_goals DECIMAL(6,2) NOT NULL,
		expected_away_goals DECIMAL(6,2) NOT NULL,
		home_win_probability DECIMAL(5,1) NOT NULL,
		draw_probability DECIMAL(5,1) NOT NULL,
		away_win_probability DECIMAL(5,1) NOT NULL,
		over25_probability DECIMAL(5,1) NOT NULL,
		history_matches INT NOT NULL,
		home_team_history INT NOT NULL,
		away_team_history INT NOT NULL,
		model_basis VARCHAR(255) NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (id),
		UNIQUE KEY idx_footstat_import_predictions_match (import_match_id),
		CONSTRAINT fk_footstat_import_predictions_match
			FOREIGN KEY (import_match_id) REFERENCES footstat_import_matches (id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
}

func openImportDB(ctx context.Context) (*sql.DB, error) {
	link, err := g.Cfg().Get(ctx, "database.default.link")
	if err != nil {
		return nil, fmt.Errorf("read database configuration: %w", err)
	}
	dsn := strings.TrimPrefix(link.String(), "mysql:")
	if dsn == link.String() {
		return nil, fmt.Errorf("database.default.link must use the mysql driver")
	}
	if !strings.Contains(dsn, "charset=") {
		separator := "?"
		if strings.Contains(dsn, "?") {
			separator = "&"
		}
		dsn += separator + "charset=utf8mb4&parseTime=true"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	for _, statement := range importSchema {
		if _, err = db.ExecContext(ctx, statement); err != nil {
			db.Close()
			return nil, fmt.Errorf("ensure import tables: %w", err)
		}
	}
	if err = ensureImportColumns(ctx, db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// ensureImportColumns 为已经存在的旧库补齐赛果/赔率列（MySQL 5.7 不支持 ADD COLUMN IF NOT EXISTS）。
func ensureImportColumns(ctx context.Context, db *sql.DB) error {
	columns := map[string]string{
		"home_odds":  "DECIMAL(8,3) NULL",
		"draw_odds":  "DECIMAL(8,3) NULL",
		"away_odds":  "DECIMAL(8,3) NULL",
		"home_goals": "TINYINT NULL",
		"away_goals": "TINYINT NULL",
	}
	rows, err := db.QueryContext(ctx, `SELECT column_name FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = 'footstat_import_matches'`)
	if err != nil {
		return fmt.Errorf("inspect import columns: %w", err)
	}
	existing := make(map[string]bool)
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			rows.Close()
			return fmt.Errorf("read import columns: %w", err)
		}
		existing[strings.ToLower(name)] = true
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("iterate import columns: %w", err)
	}
	rows.Close()
	for name, definition := range columns {
		if existing[name] {
			continue
		}
		if _, err = db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE footstat_import_matches ADD COLUMN %s %s", name, definition)); err != nil {
			return fmt.Errorf("add import column %s: %w", name, err)
		}
	}
	return nil
}

// loadImportHistory 把用户导入过、带赛果的比赛汇总成历史样本池，跨批次持续累积。
func loadImportHistory(ctx context.Context, db *sql.DB) ([]historicalMatch, error) {
	rows, err := db.QueryContext(ctx, `SELECT match_date, league, home_team, away_team, home_goals, away_goals,
		COALESCE(home_odds, 0), COALESCE(draw_odds, 0), COALESCE(away_odds, 0)
		FROM footstat_import_matches WHERE home_goals IS NOT NULL AND away_goals IS NOT NULL`)
	if err != nil {
		return nil, fmt.Errorf("读取自定义历史赛果失败: %w", err)
	}
	defer rows.Close()

	history := make([]historicalMatch, 0)
	for rows.Next() {
		var date, league, homeTeam, awayTeam string
		var homeGoals, awayGoals int
		var homeOdds, drawOdds, awayOdds float64
		if err = rows.Scan(&date, &league, &homeTeam, &awayTeam, &homeGoals, &awayGoals, &homeOdds, &drawOdds, &awayOdds); err != nil {
			return nil, fmt.Errorf("解析自定义历史赛果失败: %w", err)
		}
		day, month, year := parseCsvDate(date)
		if year == 0 || month == 0 || day == 0 {
			continue
		}
		market, hasMarket := marketProbabilities(homeOdds, drawOdds, awayOdds)
		history = append(history, historicalMatch{
			date:      year*10000 + month*100 + day,
			league:    league,
			homeTeam:  teamKey(homeTeam),
			awayTeam:  teamKey(awayTeam),
			homeGoals: float64(homeGoals),
			awayGoals: float64(awayGoals),
			market:    market,
			hasMarket: hasMarket,
		})
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历自定义历史赛果失败: %w", err)
	}
	return history, nil
}

// importHistoryForLeague 选出与该联赛相关的自定义历史样本；
// 联赛无法解析时使用全部自定义样本，保证小众/国家队赛事也有可用历史。
func importHistoryForLeague(all []historicalMatch, league, code string) []historicalMatch {
	if len(all) == 0 {
		return nil
	}
	selected := make([]historicalMatch, 0, len(all))
	for _, item := range all {
		if code == "" || strings.EqualFold(item.league, league) || resolveLeagueCode(item.league) == code {
			selected = append(selected, item)
		}
	}
	return selected
}

// predictMatches 统一生成预测：CSV 历史 + 用户导入的赛果样本 + 竞彩赔率。
// includePending 表示本批次的赛果尚未入库，需要临时并入样本池。
func predictMatches(ctx context.Context, db *sql.DB, matches []v1.ImportMatch, includePending bool) ([]v1.MatchPrediction, error) {
	imported, err := loadImportHistory(ctx, db)
	if err != nil {
		return nil, err
	}
	// 当前批次中已完赛的比赛也立即进入样本池（PredictMatch 会按日期过滤，只使用赛前样本）。
	for _, match := range matches {
		if !includePending || match.HomeGoals == nil || match.AwayGoals == nil {
			continue
		}
		day, month, year := parseCsvDate(match.Date)
		if year == 0 || month == 0 || day == 0 {
			continue
		}
		market, hasMarket := marketProbabilitySet(match)
		imported = append(imported, historicalMatch{
			date:      year*10000 + month*100 + day,
			league:    match.League,
			homeTeam:  teamKey(match.HomeTeam),
			awayTeam:  teamKey(match.AwayTeam),
			homeGoals: float64(*match.HomeGoals),
			awayGoals: float64(*match.AwayGoals),
			market:    market,
			hasMarket: hasMarket,
		})
	}

	histories := make(map[string]*leagueHistory)
	predictions := make([]v1.MatchPrediction, len(matches))
	for index, match := range matches {
		code := resolveLeagueCode(match.League)
		cacheKey := code
		if cacheKey == "" {
			cacheKey = "raw:" + strings.ToLower(match.League)
		}
		history, ok := histories[cacheKey]
		if !ok {
			loaded, loadErr := loadLeagueHistory(ctx, match.League)
			if loadErr != nil {
				return nil, fmt.Errorf("加载第 %d 场比赛的历史数据失败: %w", index+1, loadErr)
			}
			loaded.matches = append(loaded.matches, importHistoryForLeague(imported, match.League, code)...)
			history = loaded
			histories[cacheKey] = history
		}
		prediction, predictionErr := PredictMatch(history, match)
		if predictionErr != nil {
			return nil, fmt.Errorf("分析第 %d 场比赛失败: %w", index+1, predictionErr)
		}
		predictions[index] = prediction
	}
	return predictions, nil
}

// persistEnrichedOdds 把补回的赔率写回历史批次，后续查看与分析无需再次联网。
func persistEnrichedOdds(ctx context.Context, db *sql.DB, batchId int64, matches []v1.ImportMatch, filled []int) {
	for _, index := range filled {
		match := matches[index]
		_, _ = db.ExecContext(ctx, `UPDATE footstat_import_matches
			SET home_odds = ?, draw_odds = ?, away_odds = ?
			WHERE batch_id = ? AND match_date = ? AND home_team = ? AND away_team = ? AND home_odds IS NULL`,
			match.HomeOdds, match.DrawOdds, match.AwayOdds,
			batchId, match.Date, match.HomeTeam, match.AwayTeam)
	}
}

func nullableOdds(value float64) any {
	if value <= 1 {
		return nil
	}
	return value
}

func nullableGoals(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func CreateImport(ctx context.Context, req *v1.CreateImportReq) (*v1.CreateImportRes, error) {
	fileName := strings.TrimSpace(req.FileName)
	if fileName == "" || len(fileName) > 255 {
		return nil, fmt.Errorf("文件名不能为空且不能超过 255 个字符")
	}
	if len(req.Matches) == 0 || len(req.Matches) > 500 {
		return nil, fmt.Errorf("单次导入比赛数量须为 1 至 500 场")
	}
	for index := range req.Matches {
		match := &req.Matches[index]
		match.Date = strings.TrimSpace(match.Date)
		match.Time = strings.TrimSpace(match.Time)
		match.League = strings.TrimSpace(match.League)
		match.HomeTeam = strings.TrimSpace(match.HomeTeam)
		match.AwayTeam = strings.TrimSpace(match.AwayTeam)
		if match.Date == "" || match.League == "" || match.HomeTeam == "" || match.AwayTeam == "" {
			return nil, fmt.Errorf("第 %d 场比赛缺少日期、联赛或球队信息", index+1)
		}
		if match.HomeTeam == match.AwayTeam {
			return nil, fmt.Errorf("第 %d 场比赛的主队和客队不能相同", index+1)
		}
		if (match.HomeGoals == nil) != (match.AwayGoals == nil) {
			return nil, fmt.Errorf("第 %d 场比赛的主队进球与客队进球必须同时提供", index+1)
		}
		if match.HomeGoals != nil && (*match.HomeGoals < 0 || *match.HomeGoals > 20 || *match.AwayGoals < 0 || *match.AwayGoals > 20) {
			return nil, fmt.Errorf("第 %d 场比赛的进球数须在 0 至 20 之间", index+1)
		}
		for _, odds := range []float64{match.HomeOdds, match.DrawOdds, match.AwayOdds} {
			if odds < 0 || odds > 1000 {
				return nil, fmt.Errorf("第 %d 场比赛的赔率无效", index+1)
			}
		}
		if (match.HomeOdds > 0 || match.DrawOdds > 0 || match.AwayOdds > 0) &&
			(match.HomeOdds <= 1 || match.DrawOdds <= 1 || match.AwayOdds <= 1) {
			return nil, fmt.Errorf("第 %d 场比赛的主胜、平局、客胜赔率须同时提供且大于 1", index+1)
		}
	}

	// 手动上传的 CSV 若不带赔率，尝试按赛程补齐，避免整批退化成同一个预测。
	enrichOddsFromRemote(ctx, req.Matches)

	db, err := openImportDB(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	predictions, err := predictMatches(ctx, db, req.Matches, true)
	if err != nil {
		return nil, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("start import transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx,
		`INSERT INTO footstat_import_batches (file_name, match_count) VALUES (?, ?)`,
		fileName, len(req.Matches),
	)
	if err != nil {
		return nil, fmt.Errorf("save import batch: %w", err)
	}
	batchId, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("read import batch id: %w", err)
	}

	statement, err := tx.PrepareContext(ctx, `INSERT INTO footstat_import_matches
		(batch_id, match_date, kickoff_time, league, home_team, away_team,
		home_odds, draw_odds, away_odds, home_goals, away_goals) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return nil, fmt.Errorf("prepare imported matches: %w", err)
	}
	defer statement.Close()
	predictionStatement, err := tx.PrepareContext(ctx, `INSERT INTO footstat_import_predictions
		(import_match_id, predicted_score, expected_home_goals, expected_away_goals,
		home_win_probability, draw_probability, away_win_probability, over25_probability,
		history_matches, home_team_history, away_team_history, model_basis)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return nil, fmt.Errorf("prepare imported predictions: %w", err)
	}
	defer predictionStatement.Close()
	for index, match := range req.Matches {
		result, execErr := statement.ExecContext(ctx, batchId, match.Date, match.Time, match.League, match.HomeTeam, match.AwayTeam,
			nullableOdds(match.HomeOdds), nullableOdds(match.DrawOdds), nullableOdds(match.AwayOdds),
			nullableGoals(match.HomeGoals), nullableGoals(match.AwayGoals))
		if execErr != nil {
			return nil, fmt.Errorf("save imported match: %w", execErr)
		}
		matchId, idErr := result.LastInsertId()
		if idErr != nil {
			return nil, fmt.Errorf("read imported match id: %w", idErr)
		}
		prediction := predictions[index]
		_, err = predictionStatement.ExecContext(ctx, matchId, prediction.PredictedScore,
			prediction.ExpectedHomeGoals, prediction.ExpectedAwayGoals, prediction.HomeWinProbability,
			prediction.DrawProbability, prediction.AwayWinProbability, prediction.Over25Probability,
			prediction.HistoryMatches, prediction.HomeTeamHistory, prediction.AwayTeamHistory, prediction.ModelBasis)
		if err != nil {
			return nil, fmt.Errorf("save prediction for match %d: %w", index+1, err)
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit import batch: %w", err)
	}

	return &v1.CreateImportRes{Batch: v1.ImportBatch{
		Id:         batchId,
		FileName:   fileName,
		MatchCount: len(req.Matches),
		ImportedAt: time.Now().Format("2006-01-02 15:04:05"),
	}, Predictions: predictions}, nil
}

func GetImports(ctx context.Context) ([]v1.ImportBatch, error) {
	db, err := openImportDB(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, `SELECT id, file_name, match_count,
		DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s')
		FROM footstat_import_batches ORDER BY id DESC LIMIT 100`)
	if err != nil {
		return nil, fmt.Errorf("query import batches: %w", err)
	}
	defer rows.Close()

	batches := make([]v1.ImportBatch, 0)
	for rows.Next() {
		var batch v1.ImportBatch
		if err = rows.Scan(&batch.Id, &batch.FileName, &batch.MatchCount, &batch.ImportedAt); err != nil {
			return nil, fmt.Errorf("read import batch: %w", err)
		}
		batches = append(batches, batch)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate import batches: %w", err)
	}
	return batches, nil
}

func GetImportDetails(ctx context.Context, id int64) (*v1.GetImportDetailsRes, error) {
	db, err := openImportDB(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var batch v1.ImportBatch
	err = db.QueryRowContext(ctx, `SELECT id, file_name, match_count,
		DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s')
		FROM footstat_import_batches WHERE id = ?`, id).Scan(&batch.Id, &batch.FileName, &batch.MatchCount, &batch.ImportedAt)
	if err != nil {
		return nil, fmt.Errorf("读取导入批次失败: %w", err)
	}

	rows, err := db.QueryContext(ctx, `SELECT match_date, kickoff_time, league, home_team, away_team,
		COALESCE(home_odds, 0), COALESCE(draw_odds, 0), COALESCE(away_odds, 0), home_goals, away_goals
		FROM footstat_import_matches WHERE batch_id = ? ORDER BY id`, id)
	if err != nil {
		return nil, fmt.Errorf("读取导入比赛失败: %w", err)
	}

	matches := make([]v1.ImportMatch, 0, batch.MatchCount)
	for rows.Next() {
		var match v1.ImportMatch
		var homeGoals, awayGoals sql.NullInt64
		if err = rows.Scan(&match.Date, &match.Time, &match.League, &match.HomeTeam, &match.AwayTeam,
			&match.HomeOdds, &match.DrawOdds, &match.AwayOdds, &homeGoals, &awayGoals); err != nil {
			rows.Close()
			return nil, fmt.Errorf("解析导入比赛失败: %w", err)
		}
		if homeGoals.Valid && awayGoals.Valid {
			home, away := int(homeGoals.Int64), int(awayGoals.Int64)
			match.HomeGoals, match.AwayGoals = &home, &away
		}
		matches = append(matches, match)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("遍历导入比赛失败: %w", err)
	}
	rows.Close()

	// 旧批次没有保存赔率时按赛程补回；随后用当前历史样本重新计算，避免展示过时的同一批预测。
	filled := enrichOddsFromRemote(ctx, matches)
	persistEnrichedOdds(ctx, db, id, matches, filled)
	predictions, err := predictMatches(ctx, db, matches, false)
	if err != nil {
		return nil, err
	}
	return &v1.GetImportDetailsRes{Batch: batch, Predictions: predictions}, nil
}
