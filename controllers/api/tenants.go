package api

import (
	"encoding/json"
	"net/http"

	"github.com/gophish/gophish/models"
	"github.com/gophish/gophish/util/m365"
)

// GET /api/tenants
func (as *Server) GetTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := models.GetAllTenants()
	if err != nil {
		http.Error(w, "Error fetching tenants", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tenants)
}

// POST /api/tenants
func (as *Server) AddTenant(w http.ResponseWriter, r *http.Request) {
	var t models.M365Tenant
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	if err := models.SaveTenant(&t); err != nil {
		http.Error(w, "Failed to save tenant", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// POST /api/m365/import
func (as *Server) ImportFromTenant(w http.ResponseWriter, r *http.Request) {
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

	token, err := m365.GetAccessToken(tenant.TenantID, tenant.ClientID, tenant.ClientSecret)
	if err != nil {
		http.Error(w, "Failed to get token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	groups, err := m365.FetchGroupsFromGraph(token)
	if err != nil {
		http.Error(w, "Failed to fetch groups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := models.ImportGroupsToGoPhish(groups); err != nil {
		http.Error(w, "Failed to import groups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Import successful"))
}
