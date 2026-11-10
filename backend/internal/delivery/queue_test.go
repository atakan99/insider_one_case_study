package delivery

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"campaigncomposer/internal/delivery/deliverytest"
)

func addRaw(t *testing.T, rdb *redis.Client, values map[string]any) {
	t.Helper()
	if err := rdb.XAdd(context.Background(), &redis.XAddArgs{Stream: "send:sms", Values: values}).Err(); err != nil {
		t.Fatal(err)
	}
}

func assertNoPending(t *testing.T, rdb *redis.Client) {
	t.Helper()
	p, err := rdb.XPending(context.Background(), "send:sms", "workers").Result()
	if err != nil || p.Count != 0 {
		t.Fatalf("XPENDING = %v, %v; want 0", p, err)
	}
}

func TestQueue_PushBeforeEnsure(t *testing.T) {
	ctx := context.Background()
	q := NewQueue(deliverytest.NewRedis(t), "c1")
	if err := q.Push(ctx, ChannelSMS, []int64{5}); err != nil {
		t.Fatal(err)
	}
	if err := q.Ensure(ctx); err != nil {
		t.Fatal(err)
	}
	got, err := q.Read(ctx, ChannelSMS, 10, 100*time.Millisecond)
	if err != nil || len(got) != 1 || got[0].JobID != 5 {
		t.Fatalf("Read = %v, %v; want job 5", got, err)
	}
}

func TestQueue_ReadDropsPoisonEntries(t *testing.T) {
	ctx := context.Background()
	rdb := deliverytest.NewRedis(t)
	q := NewQueue(rdb, "c1")
	if err := q.Ensure(ctx); err != nil {
		t.Fatal(err)
	}
	addRaw(t, rdb, map[string]any{"job": "1"})
	addRaw(t, rdb, map[string]any{"other": "x"})
	addRaw(t, rdb, map[string]any{"job": "abc"})
	addRaw(t, rdb, map[string]any{"job": "2"})
	got, err := q.Read(ctx, ChannelSMS, 10, 100*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].JobID != 1 || got[1].JobID != 2 {
		t.Fatalf("Read = %v, want jobs 1 and 2", got)
	}
	for _, e := range got {
		if err := q.Ack(ctx, ChannelSMS, e.StreamID); err != nil {
			t.Fatal(err)
		}
	}
	assertNoPending(t, rdb)
}

func TestQueue_AutoClaimDropsPoisonEntries(t *testing.T) {
	ctx := context.Background()
	rdb := deliverytest.NewRedis(t)
	a, b := NewQueue(rdb, "a"), NewQueue(rdb, "b")
	if err := a.Ensure(ctx); err != nil {
		t.Fatal(err)
	}
	addRaw(t, rdb, map[string]any{"job": "1"})
	addRaw(t, rdb, map[string]any{"other": "x"})
	addRaw(t, rdb, map[string]any{"job": "abc"})
	addRaw(t, rdb, map[string]any{"job": "2"})
	// Consumer A takes all four without the queue, so the poison entries stay pending.
	if err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group: "workers", Consumer: "a", Streams: []string{"send:sms", ">"}, Count: 10, Block: -1,
	}).Err(); err != nil {
		t.Fatal(err)
	}
	got, err := b.AutoClaim(ctx, ChannelSMS, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].JobID != 1 || got[1].JobID != 2 {
		t.Fatalf("AutoClaim = %v, want jobs 1 and 2", got)
	}
	for _, e := range got {
		if err := b.Ack(ctx, ChannelSMS, e.StreamID); err != nil {
			t.Fatal(err)
		}
	}
	assertNoPending(t, rdb)
}

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
