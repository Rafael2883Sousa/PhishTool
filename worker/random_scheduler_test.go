package worker

import "testing"

func TestShuffleDeterministic(t *testing.T) {
	ids := []int64{1,2,3,4,5}
	seed := int64(42)
	a := shuffle(ids, &seed)
	b := shuffle(ids, &seed)
	for i := range a {
		if a[i] != b[i] { t.Fatalf("not deterministic") }
	}
}

func TestEffDelayBounds(t *testing.T) {
	cfg := &models.CampaignRandomConfig{DelayMinMinutes:3, DelayMaxMinutes:5}
	for i:=0;i<20;i++ {
		d := effDelay(cfg, i)
		if d < 3 || d > 5 { t.Fatalf("out of bounds: %d", d) }
	}
}
