-- +goose Up

DROP TABLE IF EXISTS sms_profiles;
CREATE TABLE sms_profiles (
  id                INTEGER PRIMARY KEY AUTOINCREMENT,
  name              TEXT    NOT NULL UNIQUE,
  provider          TEXT    NOT NULL DEFAULT 'twilio',
  account_sid       TEXT    NOT NULL,
  auth_token_enc    TEXT    NOT NULL,
  from_number       TEXT    NOT NULL,
  rate_limit_per_min INTEGER NOT NULL DEFAULT 60,
  created_by        INTEGER NOT NULL,
  created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_sms_profiles_name ON sms_profiles(name);

-- +goose Down
DROP TABLE IF EXISTS sms_profiles;
