package api

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

type randomPreviewReq struct {
	RandomConfig models.CampaignRandomConfig `json:"random_config"`
	Groups       []struct{ Name string `json:"name"` } `json:"groups"`
}

type randomPreviewResp struct {
	CampaignID               int64                     `json:"campaign_id"`
	Timezone                 string                    `json:"timezone"`
	TotalTargets             int                       `json:"total_targets"`
	EstimatedDurationMinutes int                       `json:"estimated_duration_minutes"`
	StartsAt                 string                    `json:"starts_at"`
	EndsAt                   string                    `json:"ends_at"`
	Rows                     []map[string]interface{}  `json:"rows"`
}

// POST /api/campaigns/preview/random
func (as *Server) PreviewRandomPlan(w http.ResponseWriter, r *http.Request) {
	var req randomPreviewReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONResponse(w, models.Response{Success:false, Message:"Invalid JSON"}, http.StatusBadRequest); return
	}
	cfg := req.RandomConfig
	cfg.CampaignID = 0 // preview
	if err := models.ValidateRandomConfig(&cfg); err != nil {
		JSONResponse(w, models.Response{Success:false, Message:err.Error()}, http.StatusBadRequest); return
	}
	loc, _ := time.LoadLocation(cfg.Timezone)
	now := time.Now().In(loc)
	start := now
	if cfg.StartTime != nil { start = cfg.StartTime.In(loc) }
	targetIDs, _ := models.GetTargetIDsForGroups(as.db, ctx.Get(r, "user_id").(int64), req.Groups)
	// gerar N amostras
	rows, etaEnd := generatePreviewRows(as.db, targetIDs, &cfg, loc, start, 200)
	resp := randomPreviewResp{
		CampaignID: 0,
		Timezone: cfg.Timezone,
		TotalTargets: len(targetIDs),
		EstimatedDurationMinutes: int(etaEnd.Sub(start).Minutes()),
		StartsAt: start.Format(time.RFC3339),
		EndsAt: etaEnd.Format(time.RFC3339),
		Rows: rows,
	}
	JSONResponse(w, resp, http.StatusOK)
}

// GET /api/campaigns/{id}/plan
func (as *Server) GetRandomPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)
	status := r.URL.Query().Get("status")
	var stPtr *string
	if status != "" { stPtr = &status }
	pl, err := models.ListPlan(as.db, id, stPtr, 500, 0)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			JSONResponse(w, models.Response{Success:false, Message:"Not found"}, http.StatusNotFound); return
		}
		JSONResponse(w, models.Response{Success:false, Message:err.Error()}, http.StatusInternalServerError); return
	}
	JSONResponse(w, pl, http.StatusOK)
}

// GET /api/campaigns/{id}/plan.csv
func (as *Server) GetRandomPlanCSV(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 10, 64)
	pl, err := models.ListPlan(as.db, id, nil, 100000, 0)
	if err != nil {
		JSONResponse(w, models.Response{Success:false, Message:err.Error()}, http.StatusInternalServerError); return
	}
	w.Header().Set("Content-Type","text/csv")
	w.Header().Set("Content-Disposition","attachment; filename=plan.csv")
	cw := csv.NewWriter(w)
	_ = cw.Write([]string{"target_id","scheduled_send_at","attempt","status"})
	for _, it := range pl.Items {
		_ = cw.Write([]string{
			strconv.FormatInt(it.TargetID,10),
			it.ScheduledSendAt.Format(time.RFC3339),
			strconv.Itoa(it.Attempt),
			it.Status,
		})
	}
	cw.Flush()
}

func generatePreviewRows(db *gorm.DB, ids []int64, cfg *models.CampaignRandomConfig, loc *time.Location, start time.Time, limit int) ([]map[string]interface{}, time.Time) {
	seq := shuffleIDs(ids, cfg.RandomSeed)
	min := cfg.DelayMinMinutes
	max := cfg.DelayMaxMinutes
	t := roundToMinute(start)
	rows := []map[string]interface{}{}
	count := 0
	for _, id := range seq {
		if count >= limit { break }
		t = nextAllowed(t, db, cfg, loc)
		rows = append(rows, map[string]interface{}{
			"target_id": id,
			"scheduled_send_at": t.Format(time.RFC3339),
			"attempt": 0,
		})
		d := randDelay(min, max, cfg.RandomSeed, count)
		t = t.Add(time.Duration(d) * time.Minute)
		count++
	}
	return rows, t
}

// --- helpers locais para preview ---
func roundToMinute(t time.Time) time.Time { return t.Truncate(time.Minute) }
func shuffleIDs(ids []int64, seed *int64) []int64 {
	out := make([]int64, len(ids)); copy(out, ids)
	s := time.Now().UnixNano()
	if seed != nil { s = *seed }
	r := randNew(s)
	for i := range out {
		j := r.Intn(i+1)
		out[i], out[j] = out[j], out[i]
	}
	return out
}
type rnd struct{ x uint64 }
func randNew(seed int64) *rnd { return &rnd{x: uint64(seed)} }
func (r *rnd) Intn(n int) int { r.x ^= r.x<<7; r.x ^= r.x>>9; r.x ^= r.x<<8; return int(r.x % uint64(n)) }
func randDelay(min,max int, seed *int64, i int) int {
	if max < min { max = min }
	s := time.Now().UnixNano() + int64(i)
	if seed != nil { s += *seed }
	r := randNew(s)
	return min + r.Intn(max-min+1)
}
func nextAllowed(t time.Time, db *gorm.DB, cfg *models.CampaignRandomConfig, loc *time.Location) time.Time {
	for {
		if cfg.ExcludeWeekends && (t.Weekday()==time.Saturday || t.Weekday()==time.Sunday) {
			t = time.Date(t.Year(), t.Month(), t.Day(), 0,0,0,0, loc).Add(24*time.Hour); continue
		}
		if cfg.ExcludeHolidays {
			if ok,_ := models.IsHoliday(db, cfg.HolidayCalendar, t); ok {
				t = time.Date(t.Year(), t.Month(), t.Day(), 0,0,0,0, loc).Add(24*time.Hour); continue
			}
		}
		return t
	}
}
	