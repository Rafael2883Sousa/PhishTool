package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
)

type CampaignRandomConfig struct {
	CampaignID       int64  `gorm:"primary_key"`
	RandomizeEnabled bool
	DelayMinMinutes  int
	DelayMaxMinutes  int
	ExcludeWeekends  bool
	ExcludeHolidays  bool
	HolidayCalendar  string
	Timezone         string
	RandomSeed       *int64
	SMTPMaxPerHour   *int
	StartTime        *time.Time
	EndTimeSoft      *time.Time
}

func ValidateRandomConfig(c *CampaignRandomConfig) error {
	if !c.RandomizeEnabled {
		return nil
	}
	if c.DelayMinMinutes < 1 || c.DelayMinMinutes > 60 {
		return errors.New("invalid delay_min_minutes")
	}
	if c.DelayMaxMinutes < c.DelayMinMinutes || c.DelayMaxMinutes > 60 {
		return errors.New("invalid delay_max_minutes")
	}
	if c.SMTPMaxPerHour != nil && (*c.SMTPMaxPerHour < 1 || *c.SMTPMaxPerHour > 120) {
		return errors.New("invalid smtp_max_per_hour")
	}
	if c.Timezone == "" {
		c.Timezone = "Europe/Lisbon"
	}
	if c.HolidayCalendar == "" {
		c.HolidayCalendar = "PT"
	}
	if c.StartTime != nil && c.EndTimeSoft != nil && c.StartTime.After(*c.EndTimeSoft) {
		return errors.New("start_time must be before end_time_soft")
	}
	return nil
}

func UpsertRandomConfig(db *gorm.DB, cfg *CampaignRandomConfig) error {
	if err := ValidateRandomConfig(cfg); err != nil {
		return err
	}
	var existing CampaignRandomConfig
	if err := db.Where("campaign_id = ?", cfg.CampaignID).First(&existing).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return db.Create(cfg).Error
		}
		return err
	}
	return db.Model(&existing).Updates(cfg).Error
}

func GetRandomConfig(db *gorm.DB, campaignID int64) (*CampaignRandomConfig, error) {
	var c CampaignRandomConfig
	if err := db.Where("campaign_id = ?", campaignID).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}
