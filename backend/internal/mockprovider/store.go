package mockprovider

import (
	"fmt"
	"sync"
	"time"
)

// Channels.
const (
	channelSMS   = "sms"
	channelEmail = "email"
)

// Message is one delivered message.
type Message struct {
	ID         string    `json:"id"`
	Channel    string    `json:"channel"`
	To         string    `json:"to"`
	Subject    string    `json:"subject"`
	Body       string    `json:"body"`
	Reference  string    `json:"reference"`
	ReceivedAt time.Time `json:"received_at"`
}

// ChannelStats holds the counters for one channel.
type ChannelStats struct {
	Requests            int64 `json:"requests"`
	Delivered           int64 `json:"delivered"`
	Rejected            int64 `json:"rejected"`
	RateLimited         int64 `json:"rate_limited"`
	Errors              int64 `json:"errors"`
	Slow                int64 `json:"slow"`
	Lost                int64 `json:"lost"`
	DuplicateDeliveries int64 `json:"duplicate_deliveries"`
}

type filter struct {
	channel, to, reference string
	limit, offset          int
}

// store keeps all messages, references and counters in memory.
type store struct {
	mu       sync.Mutex
	nextID   int64
	messages []Message // oldest first
	refs     map[string]map[string]struct{}
	stats    map[string]*ChannelStats
}

func newStore() *store {
	s := &store{}
	s.resetLocked()
	return s
}

// resetLocked keeps nextID, so an id is never used twice, also across a reset.
func (s *store) resetLocked() {
	s.messages = nil
	s.refs = map[string]map[string]struct{}{channelSMS: {}, channelEmail: {}}
	s.stats = map[string]*ChannelStats{channelSMS: {}, channelEmail: {}}
}

func (s *store) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resetLocked()
}

// add stores m with a new ID and updates Delivered and DuplicateDeliveries.
func (s *store) add(m Message) Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	m.ID = fmt.Sprintf("%s_%08d", m.Channel, s.nextID)
	s.messages = append(s.messages, m)
	st := s.stats[m.Channel]
	st.Delivered++
	if m.Reference != "" {
		if _, seen := s.refs[m.Channel][m.Reference]; seen {
			st.DuplicateDeliveries++
		} else {
			s.refs[m.Channel][m.Reference] = struct{}{}
		}
	}
	return m
}

// count runs fn on the counters of channel under the lock.
func (s *store) count(channel string, fn func(*ChannelStats)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(s.stats[channel])
}

func (s *store) snapshot() map[string]ChannelStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]ChannelStats, len(s.stats))
	for ch, st := range s.stats {
		out[ch] = *st
	}
	return out
}

// list returns the matches newest first, after offset and limit, and the
// count of all matches.
func (s *store) list(f filter) ([]Message, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []Message{}
	total := 0
	for i := len(s.messages) - 1; i >= 0; i-- {
		m := s.messages[i]
		if (f.channel != "" && m.Channel != f.channel) ||
			(f.to != "" && m.To != f.to) ||
			(f.reference != "" && m.Reference != f.reference) {
			continue
		}
		if total >= f.offset && len(out) < f.limit {
			out = append(out, m)
		}
		total++
	}
	return out, total
}
