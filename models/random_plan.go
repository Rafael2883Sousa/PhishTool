package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type CampaignSendPlan struct {
	ID              int64      `gorm:"primary_key"`
	CampaignID      int64      `gorm:"index"`
	TargetID        int64      `gorm:"index"`
	ScheduledSendAt time.Time  `gorm:"index"`
	Attempt         int
	Status          string
	LastError       *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func InsertPlanBulk(db *gorm.DB, rows []CampaignSendPlan) error {
	tx := db.Begin()
	for _, r := range rows {
		if err := tx.Create(&r).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func NextDuePlan(db *gorm.DB, campaignID int64, now time.Time) (*CampaignSendPlan, error) {
	var row CampaignSendPlan
	err := db.
		Where("campaign_id = ? AND status = ? AND scheduled_send_at <= ?", campaignID, "scheduled", now).
		Order("scheduled_send_at ASC, id ASC").
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func MarkSent(db *gorm.DB, id int64) error {
	return db.Model(&CampaignSendPlan{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": "sent", "updated_at": time.Now()}).Error
}

func MarkFailed(db *gorm.DB, id int64, msg string) error {
	return db.Model(&CampaignSendPlan{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": "failed", "last_error": msg, "updated_at": time.Now()}).Error
}

func EnqueueRetry(db *gorm.DB, prev *CampaignSendPlan, nextAt time.Time) error {
	n := CampaignSendPlan{
		CampaignID:      prev.CampaignID,
		TargetID:        prev.TargetID,
		ScheduledSendAt: nextAt,
		Attempt:         prev.Attempt + 1,
		Status:          "scheduled",
	}
	return db.Create(&n).Error
}

type PlanList struct {
	Total int                `json:"total"`
	Items []CampaignSendPlan `json:"items"`
}

func ListPlan(db *gorm.DB, campaignID int64, status *string, limit, offset int) (*PlanList, error) {
	q := db.Model(&CampaignSendPlan{}).Where("campaign_id = ?", campaignID)
	if status != nil && *status != "" {
		q = q.Where("status = ?", *status)
	}
	var total int
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	var items []CampaignSendPlan
	if err := q.Order("scheduled_send_at asc, id asc").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}
	return &PlanList{Total: total, Items: items}, nil
}
