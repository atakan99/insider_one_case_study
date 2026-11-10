package delivery

import (
	"context"
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
