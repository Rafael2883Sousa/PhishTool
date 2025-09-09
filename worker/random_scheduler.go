package worker

import (
	"time"
	"errors"
	"math"

	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/jinzhu/gorm"
)

type RandomScheduler struct {
	db   *gorm.DB
	quit chan struct{}
}

func NewRandomScheduler(db *gorm.DB) *RandomScheduler {
	return &RandomScheduler{db: db, quit: make(chan struct{})}
}

func (s *RandomScheduler) Start() {
	go s.run()
} 

func (s *RandomScheduler) Stop() { close(s.quit) }

func (s *RandomScheduler) run() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.quit:
			return
		case <-ticker.C:
			s.tick()
		}
	}
}

func (s *RandomScheduler) tick() {
	// encontra campanhas com config ativa
	var cfgs []models.CampaignRandomConfig
	if err := s.db.Where("randomize_enabled = 1").Find(&cfgs).Error; err != nil {
		return
	}
	for _, cfg := range cfgs {
		// garantir que existe plano; se vazio, gerar
		var count int
		_ = s.db.Model(&models.CampaignSendPlan{}).Where("campaign_id = ?", cfg.CampaignID).Count(&count).Error
		if count == 0 {
			if err := s.generateFullPlan(&cfg); err != nil {
				log.Error(err)
				continue
			}
		}
		// processar due
		_ = s.processDue(&cfg)
	}
}

func (s *RandomScheduler) generateFullPlan(cfg *models.CampaignRandomConfig) error {
	ids, err := models.GetCampaignTargetIDs(s.db, cfg.CampaignID)
	if err != nil { return err }
	if len(ids) == 0 { return errors.New("no targets to schedule") }

	loc, _ := time.LoadLocation(cfg.Timezone)

	now := time.Now().In(loc)
	start := now
	// preferir launch_date da campanha se for futuro
	var camp models.Campaign
	if err := s.db.Where("id = ?", cfg.CampaignID).First(&camp).Error; err == nil {
		if !camp.LaunchDate.IsZero() {
			ld := camp.LaunchDate.In(loc)
			if ld.After(now) { start = ld }
		}
	}

	seq := shuffle(ids, cfg.RandomSeed)
	rows := make([]models.CampaignSendPlan, 0, len(seq))
	t := roundMin(start)
	for i, id := range seq {
		t = nextAllowed(s.db, cfg, loc, t)
		rows = append(rows, models.CampaignSendPlan{
			CampaignID:      cfg.CampaignID,
			TargetID:        id,
			ScheduledSendAt: t,
			Attempt:         0,
			Status:          "scheduled",
		})
		d := effDelay(cfg, i)
		t = t.Add(time.Duration(d) * time.Minute)
	}
	return models.InsertPlanBulk(s.db, rows)
}

func (s *RandomScheduler) processDue(cfg *models.CampaignRandomConfig) error {
	loc, _ := time.LoadLocation(cfg.Timezone)
	now := time.Now().In(loc)
	for {
		row, err := models.NextDuePlan(s.db, cfg.CampaignID, now)
		if err != nil {
			var next models.CampaignSendPlan
			if e := s.db.
				Where("campaign_id = ? AND status = ? AND scheduled_send_at > ?", cfg.CampaignID, "scheduled", now).
				Order("scheduled_send_at ASC").First(&next).Error; e == nil {
				if d := next.ScheduledSendAt.Sub(now); d > 0 { time.Sleep(d) }
			}
			return nil
		}
		// se hoje é inválido, salta
		if isInvalidDay(s.db, cfg, row.ScheduledSendAt.In(loc)) {
			// reprograma para próximo dia útil 00:00 mantendo a ordem
			next := dayStartNextValid(s.db, cfg, row.ScheduledSendAt.In(loc))
			_ = s.db.Model(&models.CampaignSendPlan{}).Where("id = ?", row.ID).
				Updates(map[string]interface{}{"scheduled_send_at": next, "updated_at": time.Now()}).Error
			continue
		}
		// envia
		err = s.sendOne(cfg.CampaignID, row.TargetID)
		if err == nil {
			_ = models.MarkSent(s.db, row.ID)
		} else {
			if row.Attempt < 3 {
				// retry with new delay
				d := effDelay(cfg, row.Attempt+1)
				next := nextAllowed(s.db, cfg, loc, roundMin(time.Now().In(loc)).Add(time.Duration(d)*time.Minute))
				_ = models.EnqueueRetry(s.db, row, next)
				_ = models.MarkFailed(s.db, row.ID, err.Error())
			} else {
				_ = models.MarkFailed(s.db, row.ID, err.Error())
			}
		}
	}
}

func (s *RandomScheduler) sendOne(campaignID, targetID int64) error {
	// TODO: integrar com pipeline nativo de envio da tua build do Gophish:
	// return SendCampaignEmail(s.db, campaignID, targetID)
	return nil
}

// ----- helpers -----
func roundMin(t time.Time) time.Time { return t.Truncate(time.Minute) }
func shuffle(ids []int64, seed *int64) []int64 {
	out := make([]int64, len(ids)); copy(out, ids)
	var s int64
	if seed != nil { s = *seed } else { s = time.Now().UnixNano() }
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

func effDelay(cfg *models.CampaignRandomConfig, i int) int {
	min, max := cfg.DelayMinMinutes, cfg.DelayMaxMinutes
	if max < min { max = min }
	seed := int64(i)
	if cfg.RandomSeed != nil { seed += *cfg.RandomSeed }
	r := randNew(seed)
	d := min + r.Intn(max-min+1)
	if cfg.SMTPMaxPerHour != nil && *cfg.SMTPMaxPerHour > 0 {
		th := int(math.Ceil(60.0 / float64(*cfg.SMTPMaxPerHour)))
		if d < th { d = th }
	}
	return d
}

func isInvalidDay(db *gorm.DB, cfg *models.CampaignRandomConfig, t time.Time) bool {
	if cfg.ExcludeWeekends && (t.Weekday()==time.Saturday || t.Weekday()==time.Sunday) { return true }
	if cfg.ExcludeHolidays {
		ok, _ := models.IsHoliday(db, cfg.HolidayCalendar, t)
		if ok { return true }
	}
	return false
}
func nextAllowed(db *gorm.DB, cfg *models.CampaignRandomConfig, loc *time.Location, t time.Time) time.Time {
	for {
		if isInvalidDay(db, cfg, t) {
			t = time.Date(t.Year(), t.Month(), t.Day(), 0,0,0,0, loc).Add(24*time.Hour)
			continue
		}
		return t
	}
}
func dayStartNextValid(db *gorm.DB, cfg *models.CampaignRandomConfig, t time.Time) time.Time {
	loc, _ := time.LoadLocation(cfg.Timezone)
	d := time.Date(t.Year(), t.Month(), t.Day(), 0,0,0,0, loc).Add(24*time.Hour)
	for isInvalidDay(db, cfg, d) {
		d = d.Add(24 * time.Hour)
	}
	return d
}

// Placeholder: obter IDs de targets de uma campanha
// Implementa isto conforme o teu schema real:
func GetCampaignTargetIDs(db *gorm.DB, campaignID int64) ([]int64, error) {
	return models.GetCampaignTargetIDs(db, campaignID)
}

// Placeholder de envio síncrono:
// func SendCampaignEmail(db *gorm.DB, campaignID, targetID int64) error { ... }

type RandomKick struct{ db *gorm.DB }
func NewRandomKick(db *gorm.DB) *RandomKick { return &RandomKick{db: db} }
func (k *RandomKick) Kick(campaignID int64) {
    // gera plano se não existir e processa de imediato
    cfg, err := models.GetRandomConfig(k.db, campaignID); if err != nil || !cfg.RandomizeEnabled { return }
    var count int
    _ = k.db.Model(&models.CampaignSendPlan{}).Where("campaign_id = ?", campaignID).Count(&count).Error
    if count == 0 { _ = NewRandomScheduler(k.db).generateFullPlan(cfg) }
    _ = NewRandomScheduler(k.db).processDue(cfg)
}	