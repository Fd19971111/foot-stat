CREATE TABLE IF NOT EXISTS footstat_import_batches (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  file_name VARCHAR(255) NOT NULL,
  match_count INT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_footstat_import_batches_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS footstat_import_matches (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS footstat_import_predictions (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;