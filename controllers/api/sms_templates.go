package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
)

type smsTplReq struct {
	Name string `json:"name"`
	Body string `json:"body"`
}

// GET /api/sms/templates/
func (as *Server) GetSMSTemplates(w http.ResponseWriter, r *http.Request) {
	var items []models.SMSTemplate
	if err := models.DB().Order("id desc").Find(&items).Error; err != nil {
		log.Printf("GetSMSTemplates: db list failed: %v", err)
		http.Error(w, "list failed", http.StatusInternalServerError); return
	}
	_ = json.NewEncoder(w).Encode(items)
}

// POST /api/sms/templates/
func (as *Server) CreateSMSTemplate(w http.ResponseWriter, r *http.Request) {
	var req smsTplReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest); return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Body = strings.TrimSpace(req.Body)
	if req.Name == "" || req.Body == "" {
		http.Error(w, "invalid fields", http.StatusBadRequest); return
	}
	var uid uint
	if v, ok := ctx.Get(r, "user_id").(int64); ok && v >= 0 { uid = uint(v) }
	item := models.SMSTemplate{Name: req.Name, Body: req.Body, CreatedBy: uid}
	if err := models.DB().Create(&item).Error; err != nil {
		log.Printf("CreateSMSTemplate: db create failed: %v", err)
		http.Error(w, "create failed", http.StatusBadRequest); return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": item.ID})
}

// PUT /api/sms/templates/{id}
func (as *Server) UpdateSMSTemplate(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var req smsTplReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest); return
	}
	var item models.SMSTemplate
	if err := models.DB().First(&item, id).Error; err != nil {
		http.Error(w, "not found", http.StatusNotFound); return
	}
	if s := strings.TrimSpace(req.Name); s != "" { item.Name = s }
	if s := strings.TrimSpace(req.Body); s != "" { item.Body = s }
	if err := models.DB().Save(&item).Error; err != nil {
		http.Error(w, "update failed", http.StatusBadRequest); return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"status":"ok"})
}

// DELETE /api/sms/templates/{id}
func (as *Server) DeleteSMSTemplate(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if err := models.DB().Delete(&models.SMSTemplate{}, id).Error; err != nil {
		http.Error(w, "delete failed", http.StatusBadRequest); return
	}
	w.WriteHeader(http.StatusNoContent)
}
