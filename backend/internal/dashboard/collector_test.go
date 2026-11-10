package dashboard

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCollector(t *testing.T) {
	ctx, db, s := newDB(t)
	seedTotals(t, ctx, db)
	reg := prometheus.NewRegistry()
	reg.MustRegister(NewCollector(s, slog.New(slog.NewTextHandler(io.Discard, nil))))
	want := `
# HELP campaign_send_jobs Send jobs by channel and status.
# TYPE campaign_send_jobs gauge
campaign_send_jobs{channel="email",status="sent"} 2
campaign_send_jobs{channel="sms",status="failed"} 1
campaign_send_jobs{channel="sms",status="queued"} 1
campaign_send_jobs{channel="sms",status="retrying"} 1
campaign_send_jobs{channel="sms",status="sending"} 1
campaign_send_jobs{channel="sms",status="sent"} 3
campaign_send_jobs{channel="sms",status="skipped"} 1
campaign_send_jobs{channel="sms",status="unknown"} 2
# HELP campaign_campaigns Campaigns by status.
# TYPE campaign_campaigns gauge
campaign_campaigns{status="sending"} 1
`
	if err := testutil.GatherAndCompare(reg, strings.NewReader(want), "campaign_send_jobs", "campaign_campaigns"); err != nil {
		t.Error(err)
	}
}

func TestCollector_DBDown(t *testing.T) {
	_, db, s := newDB(t)
	db.Close()
	reg := prometheus.NewRegistry()
	reg.MustRegister(NewCollector(s, slog.New(slog.NewTextHandler(io.Discard, nil))))
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	if len(mfs) != 0 {
		t.Errorf("families = %d, want 0", len(mfs))
	}
}
