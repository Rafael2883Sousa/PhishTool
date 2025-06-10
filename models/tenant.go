package models

import (
	"time"
	"github.com/gophish/gophish/util"
)

type M365Tenant struct {
	ID           string    `gorm:"primary_key" json:"id"`
	TenantID     string    `json:"tenant_id"`
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	CreatedAt    time.Time `json:"created_at"`
}

func GetAllTenants() ([]M365Tenant, error) {
	var tenants []M365Tenant
	err := db.Find(&tenants).Error
	return tenants, err
}

func SaveTenant(t *M365Tenant) error {
	if t.ID == "" {
		t.ID = util.GenerateSecureRandomString(12)
	}
	return db.Create(t).Error
}

func GetTenantByID(id string) (*M365Tenant, error) {
	var tenant M365Tenant
	err := db.Where("id = ?", id).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}


