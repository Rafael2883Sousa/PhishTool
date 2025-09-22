package models

import "time"

type SMSTemplate struct {
	ID        uint      `gorm:"column:id;primary_key"              json:"id"`
	Name      string    `gorm:"column:name;unique_index;not null"  json:"name"`
	Body      string    `gorm:"column:body;not null"               json:"body"`
	CreatedBy uint      `gorm:"column:created_by;not null"         json:"created_by"`
	CreatedAt time.Time `gorm:"column:created_at"                  json:"created_at"`
}
func (SMSTemplate) TableName() string { return "sms_templates" }
