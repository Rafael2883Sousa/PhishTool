-- +goose Up
ALTER TABLE campaigns ADD COLUMN randomize_enabled INTEGER NOT NULL DEFAULT 0;
ALTER TABLE campaigns ADD COLUMN random_delay_min INTEGER NOT NULL DEFAULT 1;
ALTER TABLE campaigns ADD COLUMN random_delay_max INTEGER NOT NULL DEFAULT 60;
ALTER TABLE campaigns ADD COLUMN exclude_weekends INTEGER NOT NULL DEFAULT 0;
ALTER TABLE campaigns ADD COLUMN exclude_holidays INTEGER NOT NULL DEFAULT 1;
ALTER TABLE campaigns ADD COLUMN tz TEXT NOT NULL DEFAULT 'Europe/Lisbon';
ALTER TABLE campaigns ADD COLUMN random_seed INTEGER NULL;
ALTER TABLE campaigns ADD COLUMN smtp_max_per_hour INTEGER NULL;
ALTER TABLE campaigns ADD COLUMN send_queue_json TEXT NULL;
ALTER TABLE campaigns ADD COLUMN send_next_index INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE campaigns DROP COLUMN send_next_index;
ALTER TABLE campaigns DROP COLUMN send_queue_json;
ALTER TABLE campaigns DROP COLUMN smtp_max_per_hour;
ALTER TABLE campaigns DROP COLUMN random_seed;
ALTER TABLE campaigns DROP COLUMN tz;
ALTER TABLE campaigns DROP COLUMN exclude_holidays;
ALTER TABLE campaigns DROP COLUMN exclude_weekends;
ALTER TABLE campaigns DROP COLUMN random_delay_max;
ALTER TABLE campaigns DROP COLUMN random_delay_min;
ALTER TABLE campaigns DROP COLUMN randomize_enabled;
