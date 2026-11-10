// Package deliverytest gives tests a clean Redis.
package deliverytest

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedis returns a client on TEST_REDIS_URL and flushes its DB. It skips the
// test when TEST_REDIS_URL is empty. The optional argument is the DB number.
// Use 0 for delivery, 1 for httpapi and 2 for cmd/worker, so that packages
// that run in parallel do not share a DB.
func NewRedis(t testing.TB, db ...int) *redis.Client {
	t.Helper()
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		t.Skip("TEST_REDIS_URL not set")
	}
	opt, err := redis.ParseURL(url)
	if err != nil {
		t.Fatalf("parse TEST_REDIS_URL: %v", err)
	}
	if len(db) > 0 {
		opt.DB = db[0]
	}
	rdb := redis.NewClient(opt)
	t.Cleanup(func() { _ = rdb.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := rdb.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("flush test redis: %v", err)
	}
	return rdb
}
