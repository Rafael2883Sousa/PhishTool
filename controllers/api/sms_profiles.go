package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"log"

	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
)

var e164 = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)

type smsProfileReq struct {
	Name            string `json:"name"`
	AccountSID      string `json:"account_sid"`
	AuthToken       string `json:"auth_token"`
	FromNumber      string `json:"from_number"`
	RateLimitPerMin int    `json:"rate_limit_per_min"`
}

// GET /api/sms/profiles
func (as *Server) GetSMSProfiles(w http.ResponseWriter, r *http.Request) {
	var items []models.SMSProfile
	if err := models.DB().Order("id desc").Find(&items).Error; err != nil {
		http.Error(w, "list failed", http.StatusInternalServerError)
		return
	}
	type out struct {
		ID              uint   `json:"id"`
		Name            string `json:"name"`
		Provider        string `json:"provider"`
		AccountSID      string `json:"account_sid"`
		FromNumber      string `json:"from_number"`
		RateLimitPerMin int    `json:"rate_limit_per_min"`
		CreatedBy       uint   `json:"created_by"`
		CreatedAt       string `json:"created_at"`
	}
	resp := make([]out, 0, len(items))
	for _, p := range items {
		resp = append(resp, out{
			ID: p.ID, Name: p.Name, Provider: p.Provider, AccountSID: p.AccountSID,
			FromNumber: p.FromNumber, RateLimitPerMin: p.RateLimitPerMin,
			CreatedBy: p.CreatedBy, CreatedAt: p.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// POST /api/sms/profiles
func (as *Server) CreateSMSProfile(w http.ResponseWriter, r *http.Request) {
	var req smsProfileReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {

		enc, err := models.EncryptString(req.AuthToken)
		if err != nil {
			log.Printf("SMS profile encrypt error: %v", err) // verá "APP_ENCRYPTION_KEY not set" ou "must decode to 32 bytes"
			http.Error(w, "encryption error", http.StatusInternalServerError)
			return
		}
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || req.AccountSID == "" || req.AuthToken == "" || !e164.MatchString(req.FromNumber) {
		http.Error(w, "invalid fields", http.StatusBadRequest)
		return
	}
	enc, err := models.EncryptString(req.AuthToken)
	if err != nil {
		http.Error(w, "encryption error", http.StatusInternalServerError)
		return
	}
	uidAny := ctx.Get(r, "user_id")
	var uid uint
	if v, ok := uidAny.(int64); ok && v >= 0 {
		uid = uint(v)
	}
	item := models.SMSProfile{
		Name:            req.Name,
		Provider:        "twilio",
		AccountSID:      req.AccountSID,
		AuthTokenEnc:    enc,
		FromNumber:      req.FromNumber,
		RateLimitPerMin: req.RateLimitPerMin,
		CreatedBy:       uid,
	}
	if item.RateLimitPerMin <= 0 {
		item.RateLimitPerMin = 60
	}
	if err := models.DB().Create(&item).Error; err != nil {
		http.Error(w, "failed to create", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": item.ID})
}

// PUT /api/sms/profiles/{id}
func (as *Server) UpdateSMSProfile(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	var req smsProfileReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var item models.SMSProfile
	if err := models.DB().First(&item, id).Error; err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if s := strings.TrimSpace(req.Name); s != "" {
		item.Name = s
	}
	if req.AccountSID != "" {
		item.AccountSID = req.AccountSID
	}
	if req.AuthToken != "" {
		enc, err := models.EncryptString(req.AuthToken)
		if err != nil {
			http.Error(w, "encryption error", http.StatusInternalServerError)
			return
		}
		item.AuthTokenEnc = enc
	}
	if req.FromNumber != "" {
		if !e164.MatchString(req.FromNumber) {
			http.Error(w, "invalid from_number", http.StatusBadRequest)
			return
		}
		item.FromNumber = req.FromNumber
	}
	if req.RateLimitPerMin > 0 {
		item.RateLimitPerMin = req.RateLimitPerMin
	}
	if err := models.DB().Save(&item).Error; err != nil {
		http.Error(w, "failed to update", http.StatusBadRequest)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// DELETE /api/sms/profiles/{id}
func (as *Server) DeleteSMSProfile(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)
	if err := models.DB().Delete(&models.SMSProfile{}, id).Error; err != nil {
		http.Error(w, "failed to delete", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
