package api

import (
	"encoding/json"
	"net/http"

	"github.com/gophish/gophish/models"
	"github.com/gophish/gophish/util/m365"
	"github.com/gophish/gophish/util"
)

// func GetTenants(w http.ResponseWriter, r *http.Request) {
// 	tenants, err := models.GetAllTenants()
// 	if err != nil {
// 		http.Error(w, "Error fetching tenants", 500)
// 		return
// 	}
// 	json.NewEncoder(w).Encode(tenants)
// }

func AddTenant(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	t := models.M365Tenant{
		TenantID:     r.FormValue("tenant_id"),
		ClientID:     r.FormValue("client_id"),
		ClientSecret: r.FormValue("client_secret"),
	}
	if t.ID == "" {
		t.ID = util.GenerateSecureRandomString(12)
	}

	if err := models.SaveTenant(&t); err != nil {
		http.Error(w, "Failed to save tenant", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}


func ImportFromTenant(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		TenantID string `json:"tenant_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	tenant, err := models.GetTenantByID(payload.TenantID)
	if err != nil {
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}

	// 1. Obter token
	token, err := m365.GetAccessToken(tenant.TenantID, tenant.ClientID, tenant.ClientSecret)
	if err != nil {
		http.Error(w, "Failed to get token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 2. Buscar grupos (ou usuários) no Microsoft Graph
	groups, err := m365.FetchGroupsFromGraph(token)
	if err != nil {
		http.Error(w, "Failed to fetch groups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Inserir os grupos no GoPhish
	if err := models.ImportGroupsToGoPhish(groups); err != nil {
		http.Error(w, "Failed to import groups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Import successful"))
}

