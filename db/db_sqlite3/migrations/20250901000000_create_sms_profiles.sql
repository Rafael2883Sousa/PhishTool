-- +goose Up
CREATE TABLE IF NOT EXISTS sms_profiles (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  provider TEXT NOT NULL DEFAULT 'twilio',
  account_sid TEXT NOT NULL,
  auth_token_enc TEXT NOT NULL,
  from_number TEXT NOT NULL,
  rate_limit_per_min INTEGER NOT NULL DEFAULT 60,
  created_by INTEGER NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_sms_profiles_name ON sms_profiles(name);

-- +goose Down
DROP TABLE IF EXISTS sms_profiles;
