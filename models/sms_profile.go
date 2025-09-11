package models

import "time"

type SMSProfile struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Name            string    `gorm:"uniqueIndex;size:64;not null" json:"name"`
	Provider        string    `gorm:"size:16;not null;default:twilio" json:"provider"`
	AccountSID      string    `gorm:"size:64;not null" json:"account_sid"`
	AuthTokenEnc    string    `gorm:"size:256;not null" json:"auth_token_enc"` // store encrypted
	FromNumber      string    `gorm:"size:32;not null" json:"from_number"`     // E.164
	RateLimitPerMin int       `gorm:"not null;default:60" json:"rate_limit_per_min"`
	CreatedBy       uint      `gorm:"not null" json:"created_by"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// func AutoMigrateSMS() error {
// 	return DB().AutoMigrate(&SMSProfile{}).Error
// }