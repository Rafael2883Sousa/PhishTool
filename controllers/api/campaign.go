package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"fmt"

	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// helpers locais (Go 1.13)
func clampInt(v, lo, hi int) int {
	if v < lo { return lo }
	if v > hi { return hi }
	return v
}

type randomDTO struct {
    RandomizeEnabled bool   `json:"randomize_enabled"`
    RandomDelayMin   int    `json:"random_delay_min"`
    RandomDelayMax   int    `json:"random_delay_max"`
    ExcludeWeekends  bool   `json:"exclude_weekends"`
    ExcludeHolidays  bool   `json:"exclude_holidays"`
    TZ               string `json:"tz"`
    RandomSeed       *int64 `json:"random_seed"`
    SMTPMaxPerHour   *int   `json:"smtp_max_per_hour"`
}

func normalizeRandomFields(c *models.Campaign) {
    if !c.RandomizeEnabled { return }
    if c.TZ == "" { c.TZ = "Europe/Lisbon" }
    c.RandomDelayMin = clampInt(c.RandomDelayMin, 1, 60)
    if c.RandomDelayMax == 0 { c.RandomDelayMax = 60 }
    if c.RandomDelayMax < c.RandomDelayMin { c.RandomDelayMax = c.RandomDelayMin }
    if c.SMTPMaxPerHour != nil {
        v := clampInt(*c.SMTPMaxPerHour, 1, 120)
        c.SMTPMaxPerHour = &v
    }
}

func validateRandomFields(c *models.Campaign) error {
    if !c.RandomizeEnabled { return nil }
    if c.RandomDelayMin < 1 || c.RandomDelayMin > 60 ||
       c.RandomDelayMax < c.RandomDelayMin || c.RandomDelayMax > 60 {
        return fmt.Errorf("invalid random delay range")
    }
    if c.SMTPMaxPerHour != nil && (*c.SMTPMaxPerHour < 1 || *c.SMTPMaxPerHour > 120) {
        return fmt.Errorf("invalid smtp_max_per_hour")
    }
    return nil
}

func applyRandomFields(c *models.Campaign, d randomDTO) {
	c.RandomizeEnabled = d.RandomizeEnabled
	c.RandomDelayMin = clampInt(d.RandomDelayMin, 1, 60)
	mx := d.RandomDelayMax
	if mx < c.RandomDelayMin {
		mx = c.RandomDelayMin
	}
	c.RandomDelayMax = clampInt(mx, 1, 60)
	c.ExcludeWeekends = d.ExcludeWeekends
	c.ExcludeHolidays = d.ExcludeHolidays
	if d.TZ != "" { c.TZ = d.TZ } else { c.TZ = "Europe/Lisbon" }
	c.RandomSeed = d.RandomSeed
	c.SMTPMaxPerHour = d.SMTPMaxPerHour
}


// Campaigns returns a list of campaigns if requested via GET.
// If requested via POST, APICampaigns creates a new campaign and returns a reference to it.
func (as *Server) Campaigns(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		cs, err := models.GetCampaigns(ctx.Get(r, "user_id").(int64))
		if err != nil {
			log.Error(err)
		}
		JSONResponse(w, cs, http.StatusOK)
	//POST: Create a new campaign and return it as JSON
	case r.Method == "POST":
		c := models.Campaign{}
		// Put the request into a campaign
		err := json.NewDecoder(r.Body).Decode(&c)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
			return
		}

		normalizeRandomFields(&c)
		if err := validateRandomFields(&c); err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
    	}
		
		err = models.PostCampaign(&c, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		// If the campaign is scheduled to launch immediately, send it to the worker.
		// Otherwise, the worker will pick it up at the scheduled time
		if c.Status == models.CampaignInProgress {
			go as.worker.LaunchCampaign(c)
		}
		JSONResponse(w, c, http.StatusCreated)
	}
}

// CampaignsSummary returns the summary for the current user's campaigns
func (as *Server) CampaignsSummary(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		cs, err := models.GetCampaignSummaries(ctx.Get(r, "user_id").(int64))
		if err != nil {
			log.Error(err)
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, cs, http.StatusOK)
	}
}

// Campaign returns details about the requested campaign. If the campaign is not
// valid, APICampaign returns null.
func (as *Server) Campaign(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	c, err := models.GetCampaign(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		log.Error(err)
		JSONResponse(w, models.Response{Success: false, Message: "Campaign not found"}, http.StatusNotFound)
		return
	}
	switch {
	case r.Method == "GET":
		JSONResponse(w, c, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteCampaign(id)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting campaign"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Campaign deleted successfully!"}, http.StatusOK)
	}
}

// CampaignResults returns just the results for a given campaign to
// significantly reduce the information returned.
func (as *Server) CampaignResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	cr, err := models.GetCampaignResults(id, ctx.Get(r, "user_id").(int64))
	if err != nil {
		log.Error(err)
		JSONResponse(w, models.Response{Success: false, Message: "Campaign not found"}, http.StatusNotFound)
		return
	}
	if r.Method == "GET" {
		JSONResponse(w, cr, http.StatusOK)
		return
	}
}

// CampaignSummary returns the summary for a given campaign.
func (as *Server) CampaignSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	switch {
	case r.Method == "GET":
		cs, err := models.GetCampaignSummary(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				JSONResponse(w, models.Response{Success: false, Message: "Campaign not found"}, http.StatusNotFound)
			} else {
				JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			}
			log.Error(err)
			return
		}
		JSONResponse(w, cs, http.StatusOK)
	}
}

// CampaignComplete effectively "ends" a campaign.
// Future phishing emails clicked will return a simple "404" page.
func (as *Server) CampaignComplete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	switch {
	case r.Method == "GET":
		err := models.CompleteCampaign(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error completing campaign"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Campaign completed successfully!"}, http.StatusOK)
	}
}
