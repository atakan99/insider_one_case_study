package delivery

import (
	"context"
	"slices"
	"testing"
	"time"
)

func TestLimiter_Rate(t *testing.T) {
	clk := newFakeClock(testNow)
	l := NewLimiter(100, clk)
	var times []time.Time
	for i := 0; i < 1000; i++ {
		if err := l.Wait(context.Background()); err != nil {
			t.Fatal(err)
		}
		times = append(times, clk.Now())
	}
	// For every call, count the calls in the 1-second window that starts there.
	j := 0
	for i := range times {
		for j < len(times) && times[j].Before(times[i].Add(time.Second)) {
			j++
		}
		if n := j - i; n > 100 {
			t.Fatalf("%d calls in the second from call %d", n, i)
		}
	}
	if span := times[len(times)-1].Sub(times[0]); span < 9*time.Second {
		t.Fatalf("1000 calls took %v of fake time, want about 10 s", span)
	}
}

func TestLimiter_Pause(t *testing.T) {
	clk := newFakeClock(testNow)
	l := NewLimiter(100, clk)
	if err := l.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	l.Pause(2 * time.Second)
	start := clk.Now()
	if err := l.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := clk.Now(); got.Before(start.Add(2 * time.Second)) {
		t.Fatalf("Wait returned at +%v, want at least +2s", got.Sub(start))
	}
}

func TestLimiter_WaitCancelled(t *testing.T) {
	clk := newFakeClock(testNow)
	l := NewLimiter(100, clk)
	l.Pause(time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	clk.onSleep = cancel
	if err := l.Wait(ctx); err != context.Canceled {
		t.Fatalf("Wait = %v, want context.Canceled", err)
	}
}

// Tokens are granted on time, but the first sends are prepared slowly and the
// later ones fast. Admit must still keep every second at perSec or fewer.
func TestLimiter_AdmitSpacesSlowPreparation(t *testing.T) {
	clk := newFakeClock(testNow)
	const perSec = 100
	l := NewLimiter(perSec, clk)
	type send struct{ token, at time.Time }
	var sends []send
	for i := 0; i < 300; i++ {
		if err := l.Wait(context.Background()); err != nil {
			t.Fatal(err)
		}
		sends = append(sends, send{token: clk.Now()})
	}
	// Prep of 90 ms for the first half, 0 for the rest: without Admit, the
	// sends of both halves overlap and pass perSec in one second.
	var at []time.Time
	for i, s := range sends {
		t0 := s.token
		if i < 150 {
			t0 = t0.Add(90 * time.Millisecond)
		}
		if clk.Now().Before(t0) {
			clk.Set(t0)
		}
		if err := l.Admit(context.Background()); err != nil {
			t.Fatal(err)
		}
		at = append(at, clk.Now())
	}
	slices.SortFunc(at, func(a, b time.Time) int { return a.Compare(b) })
	j := 0
	for i := range at {
		for j < len(at) && at[j].Before(at[i].Add(time.Second)) {
			j++
		}
		if n := j - i; n > perSec {
			t.Fatalf("%d sends in the second from send %d", n, i)
		}
	}
}

func TestLimiter_AdmitCancelled(t *testing.T) {
	clk := newFakeClock(testNow)
	l := NewLimiter(1, clk)
	if err := l.Admit(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := l.Admit(ctx); err != context.Canceled {
		t.Fatalf("Admit = %v, want context.Canceled", err)
	}
}
