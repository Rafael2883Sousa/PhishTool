-- +goose Up

-- Run in sqlite3 against the DB file you fixed (absolute path)
CREATE TABLE IF NOT EXISTS sms_templates (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  name        TEXT    NOT NULL UNIQUE,
  body        TEXT    NOT NULL,
  created_by  INTEGER NOT NULL,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_sms_templates_name ON sms_templates(name);

-- +goose Down
DROP TABLE IF EXISTS sms_templates;
