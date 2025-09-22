package models

import "time"

// força nomes de coluna e tabela para evitar "account_s_id"
type SMSProfile struct {
	ID              uint      `gorm:"column:id;primary_key"                 json:"id"`
	Name            string    `gorm:"column:name;unique_index;not null"     json:"name"`
	Provider        string    `gorm:"column:provider;not null"              json:"provider"`
	AccountSID      string    `gorm:"column:account_sid;not null"           json:"account_sid"`
	AuthTokenEnc    string    `gorm:"column:auth_token_enc;not null"        json:"-"`
	FromNumber      string    `gorm:"column:from_number;not null"           json:"from_number"`
	RateLimitPerMin int       `gorm:"column:rate_limit_per_min;not null"    json:"rate_limit_per_min"`
	CreatedBy       uint      `gorm:"column:created_by;not null"            json:"created_by"`
	CreatedAt       time.Time `gorm:"column:created_at"                     json:"created_at"`
}

func (SMSProfile) TableName() string { return "sms_profiles" }
