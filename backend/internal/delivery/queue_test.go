package delivery

import (
	"context"
	"testing"
	"time"

	"campaigncomposer/internal/delivery/deliverytest"
)

func TestQueue_PushReadAck(t *testing.T) {
	ctx := context.Background()
	rdb := deliverytest.NewRedis(t)
	q := NewQueue(rdb, "c1")
	if err := q.Ensure(ctx); err != nil {
		t.Fatal(err)
	}
	if err := q.Push(ctx, ChannelSMS, []int64{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	got, err := q.Read(ctx, ChannelSMS, 10, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d entries, want 3", len(got))
	}
	for i, e := range got {
		if e.JobID != int64(i+1) {
			t.Errorf("entry %d job = %d, want %d", i, e.JobID, i+1)
		}
	}
	msgs, err := rdb.XRange(ctx, "send:sms", "-", "+").Result()
	if err != nil || len(msgs) != 3 || msgs[0].Values["job"] == nil {
		t.Fatalf("XRANGE = %v, %v; want 3 entries with field job", msgs, err)
	}
	for _, e := range got {
		if err := q.Ack(ctx, ChannelSMS, e.StreamID); err != nil {
			t.Fatal(err)
		}
	}
	p, err := rdb.XPending(ctx, "send:sms", "workers").Result()
	if err != nil || p.Count != 0 {
		t.Fatalf("XPENDING = %v, %v; want 0", p, err)
	}
	if n, _ := rdb.XLen(ctx, "send:email").Result(); n != 0 {
		t.Errorf("send:email length = %d, want 0", n)
	}
}

func TestQueue_EnsureTwice(t *testing.T) {
	ctx := context.Background()
	q := NewQueue(deliverytest.NewRedis(t), "c1")
	for i := 0; i < 2; i++ {
		if err := q.Ensure(ctx); err != nil {
			t.Fatalf("Ensure call %d: %v", i+1, err)
		}
	}
}

func TestQueue_AutoClaim(t *testing.T) {
	ctx := context.Background()
	rdb := deliverytest.NewRedis(t)
	a, b := NewQueue(rdb, "a"), NewQueue(rdb, "b")
	if err := a.Ensure(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.Push(ctx, ChannelEmail, []int64{7}); err != nil {
		t.Fatal(err)
	}
	got, err := a.Read(ctx, ChannelEmail, 10, 100*time.Millisecond)
	if err != nil || len(got) != 1 {
		t.Fatalf("Read = %v, %v", got, err)
	}
	claimed, err := b.AutoClaim(ctx, ChannelEmail, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(claimed) != 1 || claimed[0].JobID != 7 || claimed[0].StreamID != got[0].StreamID {
		t.Fatalf("claimed = %v, want job 7", claimed)
	}
}

func TestQueue_NoGroup(t *testing.T) {
	ctx := context.Background()
	rdb := deliverytest.NewRedis(t)
	q := NewQueue(rdb, "c1")
	if err := q.Ensure(ctx); err != nil {
		t.Fatal(err)
	}
	if err := rdb.Del(ctx, "send:sms").Err(); err != nil {
		t.Fatal(err)
	}
	_, err := q.Read(ctx, ChannelSMS, 1, 10*time.Millisecond)
	if err == nil || !IsNoGroup(err) {
		t.Fatalf("err = %v, want NOGROUP", err)
	}
}
