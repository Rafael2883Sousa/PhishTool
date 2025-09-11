-- +goose Up
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
  smtp_max_per_hour  INTEGER
  -- start_time / end_time removidos pois a UI não usa
);

CREATE TABLE IF NOT EXISTS holidays (
  calendar TEXT NOT NULL,
  date     DATE NOT NULL,
  PRIMARY KEY (calendar, date)
);

-- +goose Down
DROP TABLE IF EXISTS holidays;
DROP TABLE IF EXISTS campaign_random_config;
