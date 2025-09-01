package models

import (
	"github.com/jinzhu/gorm"
)

func GetCampaignTargetIDs(db *gorm.DB, campaignID int64) ([]int64, error) {
	type row struct{ ID int64 }
	var rows []row
	q := `
SELECT DISTINCT t.id AS id
FROM targets t
JOIN group_targets gt   ON gt.target_id = t.id
JOIN groups g           ON g.id = gt.group_id
JOIN campaign_groups cg ON cg.group_id = g.id
WHERE cg.campaign_id = ?`
	if err := db.Raw(q, campaignID).Scan(&rows).Error; err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, r := range rows { ids = append(ids, r.ID) }
	return ids, nil
}

// usado na preview: grupos por nome do utilizador atual
type GroupName struct{ Name string `json:"name"` }

func GetTargetIDsForGroups(db *gorm.DB, userID int64, groups []GroupName) ([]int64, error) {
	if len(groups) == 0 {
		return []int64{}, nil
	}
	names := make([]string, 0, len(groups))
	for _, g := range groups { if g.Name != "" { names = append(names, g.Name) } }
	type row struct{ ID int64 }
	var rows []row
	q := `
SELECT DISTINCT t.id AS id
FROM targets t
JOIN group_targets gt ON gt.target_id = t.id
JOIN groups g         ON g.id = gt.group_id
WHERE g.user_id = ? AND g.name IN (?)`
	if err := db.Raw(q, userID, names).Scan(&rows).Error; err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, r := range rows { ids = append(ids, r.ID) }
	return ids, nil
}
