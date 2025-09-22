package api

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

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

// GET /api/sms/profiles/
func (as *Server) GetSMSProfiles(w http.ResponseWriter, r *http.Request) {
	var items []models.SMSProfile
	if err := models.DB().Order("id desc").Find(&items).Error; err != nil {
		log.Printf("GetSMSProfiles: db list failed: %v", err)
		http.Error(w, "list failed", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(items)
}

// DEV fallback: guarda token em base64 com prefixo quando não há APP_ENCRYPTION_KEY válida.
func devPlainToken(plain string) string {
	return "plain:v0:" + base64.StdEncoding.EncodeToString([]byte(plain))
}

// POST /api/sms/profiles/
func (as *Server) CreateSMSProfile(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var req smsProfileReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("CreateSMSProfile: bad json: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.AccountSID = strings.TrimSpace(req.AccountSID)
	req.FromNumber = strings.TrimSpace(req.FromNumber)
	req.AuthToken = strings.TrimSpace(req.AuthToken)

	if req.Name == "" || req.AccountSID == "" || !e164.MatchString(req.FromNumber) {
		log.Printf("CreateSMSProfile: invalid fields name=%q sid_set=%t from=%q", req.Name, req.AccountSID != "", req.FromNumber)
		http.Error(w, "invalid fields", http.StatusBadRequest)
		return
	}

	// cifra (ou fallback em DEV)
	enc := ""
	if req.AuthToken != "" {
		if v, err := models.EncryptString(req.AuthToken); err != nil {
			log.Printf("CreateSMSProfile: EncryptString failed, using DEV fallback: %v", err)
			enc = devPlainToken(req.AuthToken)
		} else {
			enc = v
		}
	}

	var uid uint
	if v, ok := ctx.Get(r, "user_id").(int64); ok && v >= 0 {
		uid = uint(v)
	}

	item := models.SMSProfile{
		Name:            req.Name,
		Provider:        "twilio",
		AccountSID:      req.AccountSID,
		AuthTokenEnc:    enc,
		FromNumber:      req.FromNumber,
		RateLimitPerMin: ifnz(req.RateLimitPerMin, 60),
		CreatedBy:       uid,
	}

	// DEBUG: confirmar ficheiro da BD e colunas vistas pelo processo
	type dbfile struct{ Seq int; Name, File string }
	var dbl []dbfile
	models.DB().Raw("PRAGMA database_list;").Scan(&dbl)
	if len(dbl) > 0 {
		log.Printf("SQLite file in use: %s", dbl[0].File)
	}
	type col struct{ Name string `gorm:"column:name"` }
	var cols []col
	models.DB().Raw("PRAGMA table_info(sms_profiles);").Scan(&cols)
	names := make([]string, 0, len(cols))
	for _, c := range cols { names = append(names, c.Name) }
	log.Printf("sms_profiles columns: %v", names)


	if err := models.DB().Create(&item).Error; err != nil {
		log.Printf("CreateSMSProfile: db create failed: %v (name=%q sid_set=%t from=%q uid=%d) in %s",
			err, item.Name, item.AccountSID != "", item.FromNumber, item.CreatedBy, time.Since(start))
		http.Error(w, "db create failed", http.StatusInternalServerError)
		return
	}

	log.Printf("CreateSMSProfile: created id=%d name=%q in %s", item.ID, item.Name, time.Since(start))
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": item.ID})
}

// PUT /api/sms/profiles/{id}
func (as *Server) UpdateSMSProfile(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var req smsProfileReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("UpdateSMSProfile: bad json: %v", err)
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
	if s := strings.TrimSpace(req.AccountSID); s != "" {
		item.AccountSID = s
	}
	if s := strings.TrimSpace(req.AuthToken); s != "" {
		if v, err := models.EncryptString(s); err != nil {
			log.Printf("UpdateSMSProfile: EncryptString failed, using DEV fallback: %v", err)
			item.AuthTokenEnc = devPlainToken(s)
		} else {
			item.AuthTokenEnc = v
		}
	}
	if s := strings.TrimSpace(req.FromNumber); s != "" {
		if !e164.MatchString(s) {
			http.Error(w, "invalid from_number", http.StatusBadRequest)
			return
		}
		item.FromNumber = s
	}
	if req.RateLimitPerMin > 0 {
		item.RateLimitPerMin = req.RateLimitPerMin
	}
	if err := models.DB().Save(&item).Error; err != nil {
		log.Printf("UpdateSMSProfile: db save failed: %v", err)
		http.Error(w, "update failed", http.StatusBadRequest)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// DELETE /api/sms/profiles/{id}
func (as *Server) DeleteSMSProfile(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if err := models.DB().Delete(&models.SMSProfile{}, id).Error; err != nil {
		log.Printf("DeleteSMSProfile: db delete failed: %v", err)
		http.Error(w, "delete failed", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func ifnz(v, d int) int { if v > 0 { return v }; return d }
