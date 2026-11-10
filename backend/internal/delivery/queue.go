package delivery

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	groupName    = "workers"
	jobField     = "job"
	streamMaxLen = 100000
)

// Entry is one queue entry: the stream id and the send job id.
type Entry struct {
	StreamID string
	JobID    int64
}

// Queue wraps the Redis streams send:sms and send:email. All workers share
// the group "workers". Consumer is the name of this worker in the group.
type Queue struct {
	rdb      *redis.Client
	consumer string
}

// NewQueue returns a queue for one consumer name.
func NewQueue(rdb *redis.Client, consumer string) *Queue {
	return &Queue{rdb: rdb, consumer: consumer}
}

func streamKey(ch Channel) string { return "send:" + string(ch) }

// Ensure creates the streams and the group. A group that exists is not an error.
func (q *Queue) Ensure(ctx context.Context) error {
	for _, ch := range Channels {
		err := q.rdb.XGroupCreateMkStream(ctx, streamKey(ch), groupName, "$").Err()
		if err != nil && !strings.HasPrefix(err.Error(), "BUSYGROUP") {
			return fmt.Errorf("create group on %s: %w", streamKey(ch), err)
		}
	}
	return nil
}

// Push adds one entry for each send job id, in order.
func (q *Queue) Push(ctx context.Context, ch Channel, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	pipe := q.rdb.Pipeline()
	for _, id := range ids {
		pipe.XAdd(ctx, &redis.XAddArgs{
			Stream: streamKey(ch),
			MaxLen: streamMaxLen,
			Approx: true,
			Values: map[string]any{jobField: id},
		})
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("push to %s: %w", streamKey(ch), err)
	}
	return nil
}

// Read returns up to n new entries for this consumer. It waits up to block
// for the first entry. It returns no entries and no error on timeout.
func (q *Queue) Read(ctx context.Context, ch Channel, n int, block time.Duration) ([]Entry, error) {
	if block <= 0 {
		block = -1 // go-redis omits BLOCK for a negative value
	}
	res, err := q.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: q.consumer,
		Streams:  []string{streamKey(ch), ">"},
		Count:    int64(n),
		Block:    block,
	}).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", streamKey(ch), err)
	}
	var out []Entry
	for _, s := range res {
		out = append(out, q.entries(ctx, ch, s.Messages)...)
	}
	return out, nil
}

// Ack removes an entry from the pending list of the group.
func (q *Queue) Ack(ctx context.Context, ch Channel, streamID string) error {
	if err := q.rdb.XAck(ctx, streamKey(ch), groupName, streamID).Err(); err != nil {
		return fmt.Errorf("ack %s: %w", streamKey(ch), err)
	}
	return nil
}

// AutoClaim takes up to n entries that another consumer has held for at least
// minIdle, and gives them to this consumer.
func (q *Queue) AutoClaim(ctx context.Context, ch Channel, minIdle time.Duration, n int) ([]Entry, error) {
	msgs, _, err := q.rdb.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream:   streamKey(ch),
		Group:    groupName,
		Consumer: q.consumer,
		MinIdle:  minIdle,
		Start:    "0-0",
		Count:    int64(n),
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("autoclaim %s: %w", streamKey(ch), err)
	}
	return q.entries(ctx, ch, msgs), nil
}

// entries converts stream messages. An entry without a valid job id can never
// work, so it is acknowledged and dropped.
func (q *Queue) entries(ctx context.Context, ch Channel, msgs []redis.XMessage) []Entry {
	out := make([]Entry, 0, len(msgs))
	for _, m := range msgs {
		id, err := strconv.ParseInt(fmt.Sprint(m.Values[jobField]), 10, 64)
		if err != nil {
			_ = q.rdb.XAck(ctx, streamKey(ch), groupName, m.ID).Err()
			continue
		}
		out = append(out, Entry{StreamID: m.ID, JobID: id})
	}
	return out
}

// IsNoGroup reports whether err means that the stream or the group is gone,
// for example after Redis lost its data.
func IsNoGroup(err error) bool {
	return err != nil && strings.Contains(err.Error(), "NOGROUP")
}
