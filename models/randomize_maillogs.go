package models

import (
	"math"
	"time"

	"github.com/jinzhu/gorm"
	
)

// Reagenda os MailLogs de uma campanha para uma sequência aleatória, em minutos,
// respeitando fins de semana/feriados e throttle opcional.
func RandomizeMailLogsSchedule(db *gorm.DB, campaignID int64, cfg *CampaignRandomConfig) error {
	// 1) Base temporal = max(agora, launch_date)
	loc, _ := time.LoadLocation(zeroOr(cfg.Timezone, "Europe/Lisbon"))
	now := time.Now().In(loc)
	start := now
	var camp Campaign
	if err := db.Where("id = ?", campaignID).First(&camp).Error; err == nil {
		if !camp.LaunchDate.IsZero() {
			ld := camp.LaunchDate.In(loc)
			if ld.After(start) {
				start = ld
			}
		}
	}
	t := start.Truncate(time.Minute)

	// 2) Buscar todos os MailLogs desta campanha
	var logs []MailLog
	if err := db.Where("campaign_id = ?", campaignID).Order("id ASC").Find(&logs).Error; err != nil {
		return err
	}
	if len(logs) == 0 {
		return nil
	}

	// 3) Embaralhar índices com seed estável
	idx := make([]int, len(logs))
	for i := range idx {
		idx[i] = i
	}
	shuffleIdx(idx, cfg.RandomSeed)

	// 4) Atualizar send_date de cada MailLog
	for i, k := range idx {
		t = nextAllowedDay(db, cfg, loc, t)
		// guardar em UTC (Gophish usa UTC na DB)
		if err := db.Model(&logs[k]).Update("send_date", t.UTC()).Error; err != nil {
			return err
		}
		// incremento aleatório uniforme em minutos, com throttle opcional
		d := effDelayMinutes(cfg, i)
		t = t.Add(time.Duration(d) * time.Minute)
	}
	return nil
}

func RandomEnabledForCampaign(db *gorm.DB, campaignID int64) bool {
    var c CampaignRandomConfig
    if err := db.Where("campaign_id=? AND randomize_enabled=1", campaignID).First(&c).Error; err != nil {
        return false
    }
    return true
}

// --------- helpers locais ---------

func zeroOr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func shuffleIdx(a []int, seed *int64) {
	var s int64
	if seed != nil {
		s = *seed
	} else {
		s = time.Now().UnixNano()
	}
	x := uint64(s)
	rand := func(n int) int {
		x ^= x << 7
		x ^= x >> 9
		x ^= x << 8
		return int(x % uint64(n))
	}
	for i := range a {
		j := rand(i + 1)
		a[i], a[j] = a[j], a[i]
	}
}

func effDelayMinutes(cfg *CampaignRandomConfig, i int) int {
	min, max := cfg.DelayMinMinutes, cfg.DelayMaxMinutes
	if min < 1 {
		min = 1
	}
	if max < min {
		max = min
	}
	// PRNG simples por índice + seed
	var s int64 = int64(i)
	if cfg.RandomSeed != nil {
		s += *cfg.RandomSeed
	}
	x := uint64(s)
	x ^= x << 7
	x ^= x >> 9
	x ^= x << 8
	d := min + int(x%uint64(max-min+1))

	// throttle por hora
	if cfg.SMTPMaxPerHour != nil && *cfg.SMTPMaxPerHour > 0 {
		th := int(math.Ceil(60.0 / float64(*cfg.SMTPMaxPerHour)))
		if d < th {
			d = th
		}
	}
	return d
}

func nextAllowedDay(db *gorm.DB, cfg *CampaignRandomConfig, loc *time.Location, t time.Time) time.Time {
	for {
		if cfg.ExcludeWeekends && (t.Weekday() == time.Saturday || t.Weekday() == time.Sunday) {
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc).Add(24 * time.Hour)
			continue
		}
		if cfg.ExcludeHolidays {
			if ok, _ := IsHoliday(db, zeroOr(cfg.HolidayCalendar, "PT"), t); ok {
				t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc).Add(24 * time.Hour)
				continue
			}
		}
		return t
	}
}
