package workspace

import (
	"testing"
	"time"
)

func TestComputeNextRunAt_Manual(t *testing.T) {
	next, err := ComputeNextRunAt("manual", "", "Asia/Shanghai", time.Now())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if next != nil {
		t.Fatalf("expected nil next run for manual trigger")
	}
}

func TestComputeNextRunAt_IntervalHours(t *testing.T) {
	now := time.Date(2026, 3, 1, 10, 0, 0, 0, time.FixedZone("CST", 8*3600))
	next, err := ComputeNextRunAt("interval_hours", "2", "Asia/Shanghai", now)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if next == nil {
		t.Fatalf("expected non-nil next run")
	}
	if got := next.Sub(now); got < 2*time.Hour || got > 2*time.Hour+time.Second {
		t.Fatalf("expected around 2h, got %v", got)
	}
}

func TestComputeNextRunAt_DailyValidation(t *testing.T) {
	_, err := ComputeNextRunAt("daily", "25:30", "Asia/Shanghai", time.Now())
	if err == nil {
		t.Fatalf("expected validation err for invalid daily trigger")
	}
}

func TestDefaultAsyncStatus(t *testing.T) {
	if got := DefaultAsyncStatus("manual"); got != "idle" {
		t.Fatalf("expected idle for manual, got %s", got)
	}
	if got := DefaultAsyncStatus("daily"); got != "scheduled" {
		t.Fatalf("expected scheduled for daily, got %s", got)
	}
}
