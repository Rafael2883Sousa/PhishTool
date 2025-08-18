package models

import "gorm.io/gorm"

func GetCampaignTargetIDs(db *gorm.DB, campaignID int64) ([]int64, error) {
    // Ajusta consoante o teu schema de Campaign → Groups → Targets
    var ids []int64
    // Exemplo: SELECT DISTINCT t.id FROM targets t
    // JOIN group_targets gt ON gt.target_id=t.id
    // JOIN groups g ON g.id=gt.group_id
    // JOIN campaign_groups cg ON cg.group_id=g.id
    // WHERE cg.campaign_id=?
    // Implementa conforme as tuas tabelas.
    return ids, nil
}
