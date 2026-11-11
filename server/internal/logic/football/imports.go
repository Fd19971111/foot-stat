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
	return db, nil
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
	}

	histories := make(map[string]*leagueHistory)
	predictions := make([]v1.MatchPrediction, len(req.Matches))
	for index, match := range req.Matches {
		code := resolveLeagueCode(match.League)
		cacheKey := code
		if cacheKey == "" {
			cacheKey = strings.ToLower(match.League)
		}
		history, ok := histories[cacheKey]
		if !ok {
			loadedHistory, loadErr := loadLeagueHistory(ctx, match.League)
			if loadErr != nil {
				return nil, fmt.Errorf("加载第 %d 场比赛的历史数据失败: %w", index+1, loadErr)
			}
			history = loadedHistory
			histories[cacheKey] = history
		}
		prediction, predictionErr := PredictMatch(history, match)
		if predictionErr != nil {
			return nil, fmt.Errorf("分析第 %d 场比赛失败: %w", index+1, predictionErr)
		}
		predictions[index] = prediction
	}

	db, err := openImportDB(ctx)
	if err != nil {
		return nil, err
	}
	defer db.Close()

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
		(batch_id, match_date, kickoff_time, league, home_team, away_team) VALUES (?, ?, ?, ?, ?, ?)`)
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
		result, execErr := statement.ExecContext(ctx, batchId, match.Date, match.Time, match.League, match.HomeTeam, match.AwayTeam)
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

	rows, err := db.QueryContext(ctx, `SELECT m.match_date, m.kickoff_time, m.league, m.home_team, m.away_team,
		p.predicted_score, p.expected_home_goals, p.expected_away_goals,
		p.home_win_probability, p.draw_probability, p.away_win_probability, p.over25_probability,
		p.history_matches, p.home_team_history, p.away_team_history, p.model_basis
		FROM footstat_import_matches m
		JOIN footstat_import_predictions p ON p.import_match_id = m.id
		WHERE m.batch_id = ? ORDER BY m.id`, id)
	if err != nil {
		return nil, fmt.Errorf("读取比赛预测失败: %w", err)
	}
	defer rows.Close()

	predictions := make([]v1.MatchPrediction, 0, batch.MatchCount)
	for rows.Next() {
		var prediction v1.MatchPrediction
		if err = rows.Scan(&prediction.MatchDate, &prediction.KickoffTime, &prediction.League,
			&prediction.HomeTeam, &prediction.AwayTeam, &prediction.PredictedScore,
			&prediction.ExpectedHomeGoals, &prediction.ExpectedAwayGoals, &prediction.HomeWinProbability,
			&prediction.DrawProbability, &prediction.AwayWinProbability, &prediction.Over25Probability,
			&prediction.HistoryMatches, &prediction.HomeTeamHistory, &prediction.AwayTeamHistory, &prediction.ModelBasis); err != nil {
			return nil, fmt.Errorf("解析比赛预测失败: %w", err)
		}
		predictions = append(predictions, prediction)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("遍历比赛预测失败: %w", err)
	}
	return &v1.GetImportDetailsRes{Batch: batch, Predictions: predictions}, nil
}
