package worker

import (
    "context"
    "encoding/json"
    "math"
    "math/rand"
    "time"

    "github.com/gophish/gophish/models"
    "gorm.io/gorm"
)

type RandomRunner struct {
    db  *gorm.DB
}

func NewRandomRunner(db *gorm.DB) *RandomRunner { return &RandomRunner{db: db} }

func (r *RandomRunner) RunCampaign(ctx context.Context, c *models.Campaign) error {
    if !c.RandomizeEnabled {
        return nil
    }
    loc, err := time.LoadLocation(z(c.TZ, "Europe/Lisbon"))
    if err != nil { loc = time.FixedZone("Europe/Lisbon", 0) }

    // Espera por start_time se existir
    if !c.LaunchDate.IsZero() {
        start := c.LaunchDate.In(loc)
        now := time.Now().In(loc)
        if now.Before(start) {
            time.Sleep(start.Sub(now))
        }
    }

    // Gera fila embaralhada se vazio
    var queue []int64
    if c.SendQueueJSON == nil || *c.SendQueueJSON == "" {
        ids, err := models.GetCampaignTargetIDs(r.db, c.Id)
        if err != nil { return err }
        seed := deriveSeed(c)
        queue = shuffle(ids, seed)
        b, _ := json.Marshal(queue)
        s := string(b)
        c.SendQueueJSON = &s
        c.SendNextIndex = 0
        if err := r.db.Model(c).Select("send_queue_json", "send_next_index").Updates(c).Error; err != nil {
            return err
        }
    } else {
        _ = json.Unmarshal([]byte(*c.SendQueueJSON), &queue)
    }

    // Loop sequencial
    idx := c.SendNextIndex
    for idx < len(queue) {
        if ctx.Err() != nil { return ctx.Err() }

        now := time.Now().In(loc)
        if invalidDay(now, loc, c.ExcludeWeekends, c.ExcludeHolidays) {
            sleepUntil(nextValidDayStart(now, loc, c.ExcludeWeekends, c.ExcludeHolidays))
            continue
        }

        targetID := queue[idx]
        ok := r.sendWithRetries(c, targetID)
        // avança sempre; keeps it simple
        idx++
        if err := r.db.Model(c).UpdateColumn("send_next_index", idx).Error; err != nil { return err }

        // Delay
        delayMin := clamp(c.RandomDelayMin, 1, 60)
        delayMax := clamp(c.RandomDelayMax, delayMin, 60)
        eff := delayMin + rand.Intn(delayMax-delayMin+1)

        if c.SMTPMaxPerHour != nil && *c.SMTPMaxPerHour > 0 {
            rateMin := int(math.Ceil(60.0 / float64(*c.SMTPMaxPerHour)))
            if eff < rateMin { eff = rateMin }
        }
        time.Sleep(time.Duration(eff) * time.Minute)
    }
    return nil
}

// ----- helpers -----

func (r *RandomRunner) sendWithRetries(c *models.Campaign, targetID int64) bool {
    const maxAttempts = 3
    for attempt := 1; attempt <= maxAttempts; attempt++ {
        if err := SendCampaignEmail(r.db, c.Id, targetID); err == nil {
            return true
        }
        // backoff leve aleatório em memória
        time.Sleep(time.Duration(1+rand.Intn(3)) * time.Minute)
    }
    return false
}

func SendCampaignEmail(db *gorm.DB, campaignID, targetID int64) error {
    // Chama o caminho normal de envio já existente no Gophish
    // Implementa a integração aqui.
    return nil
}

func shuffle(ids []int64, seed int64) []int64 {
    out := make([]int64, len(ids)); copy(out, ids)
    rng := rand.New(rand.NewSource(seed))
    rng.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
    return out
}

func deriveSeed(c *models.Campaign) int64 {
    if c.RandomSeed != nil { return *c.RandomSeed }
    // seed determinística por campanha
    return int64(146959810) ^ int64(c.Id<<13)
}

func z(s, def string) string { if s == "" { return def }; return s }
func clamp(v, lo, hi int) int { if v < lo { return lo }; if v > hi { return hi }; return v }

func invalidDay(t time.Time, loc *time.Location, exclW, exclH bool) bool {
    if exclW && (t.Weekday() == time.Saturday || t.Weekday() == time.Sunday) { return true }
    if exclH && isHolidayPT(t.In(loc)) { return true }
    return false
}

func nextValidDayStart(t time.Time, loc *time.Location, exclW, exclH bool) time.Time {
    d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc).Add(24 * time.Hour)
    for {
        if !( (exclW && (d.Weekday()==time.Saturday || d.Weekday()==time.Sunday)) || (exclH && isHolidayPT(d)) ) {
            return d
        }
        d = d.Add(24 * time.Hour)
    }
}

func sleepUntil(ts time.Time) { if d := time.Until(ts); d > 0 { time.Sleep(d) } }
