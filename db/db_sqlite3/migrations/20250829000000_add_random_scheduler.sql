-- +goose Up
BEGIN;

CREATE TABLE IF NOT EXISTS campaign_random_config (
  campaign_id        INTEGER PRIMARY KEY,
  randomize_enabled  BOOLEAN NOT NULL DEFAULT 0,
  delay_min_minutes  INTEGER,
  delay_max_minutes  INTEGER,
  exclude_weekends   BOOLEAN NOT NULL DEFAULT 0,
  exclude_holidays   BOOLEAN NOT NULL DEFAULT 1,
  holiday_calendar   TEXT NOT NULL DEFAULT 'PT',
  timezone           TEXT NOT NULL DEFAULT 'Europe/Lisbon',
  random_seed        BIGINT,
  smtp_max_per_hour  INTEGER,
  start_time         TIMESTAMP,
  end_time_soft      TIMESTAMP,
  FOREIGN KEY(campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS campaign_send_plan (
  id                 INTEGER PRIMARY KEY AUTOINCREMENT,
  campaign_id        INTEGER NOT NULL,
  target_id          INTEGER NOT NULL,
  scheduled_send_at  TIMESTAMP NOT NULL,
  attempt            INTEGER NOT NULL DEFAULT 0,
  status             TEXT NOT NULL DEFAULT 'scheduled',
  last_error         TEXT,
  created_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY(campaign_id) REFERENCES campaigns(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_plan_campaign_time ON campaign_send_plan(campaign_id, scheduled_send_at);
CREATE INDEX IF NOT EXISTS idx_plan_campaign_target ON campaign_send_plan(campaign_id, target_id);

CREATE TABLE IF NOT EXISTS holidays (
  calendar TEXT NOT NULL,
  date     DATE NOT NULL,
  PRIMARY KEY (calendar, date)
);

COMMIT;

-- +goose Down
BEGIN;
DROP TABLE IF EXISTS campaign_send_plan;
DROP TABLE IF EXISTS campaign_random_config;
DROP TABLE IF EXISTS holidays;
COMMIT;
