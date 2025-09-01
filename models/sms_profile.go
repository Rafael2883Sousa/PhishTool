package models

import "time"

// GORM v1 uses `primary_key` tag (not `primaryKey`).
type SMSProfile struct {
	ID              uint      `gorm:"primary_key" json:"id"`
	Name            string    `gorm:"unique_index;size:64;not null" json:"name"`
	Provider        string    `gorm:"size:16;not null;default:'twilio'" json:"provider"`
	AccountSID      string    `gorm:"size:64;not null" json:"account_sid"`
	AuthTokenEnc    string    `gorm:"size:512;not null" json:"-"`
	FromNumber      string    `gorm:"size:32;not null" json:"from_number"`
	RateLimitPerMin int       `gorm:"not null;default:60" json:"rate_limit_per_min"`
	CreatedBy       uint      `gorm:"not null" json:"created_by"`
	CreatedAt       time.Time `gorm:"" json:"created_at"`
}
