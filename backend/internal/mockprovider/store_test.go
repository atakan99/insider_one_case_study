package mockprovider

import (
	"testing"
	"time"
)

func TestStore_IDs(t *testing.T) {
	s := newStore()
	now := time.Now()
	a := s.add(Message{Channel: "sms", ReceivedAt: now})
	b := s.add(Message{Channel: "email", ReceivedAt: now})
	c := s.add(Message{Channel: "sms", ReceivedAt: now})
	if a.ID != "sms_00000001" || b.ID != "email_00000002" || c.ID != "sms_00000003" {
		t.Fatalf("ids = %q %q %q", a.ID, b.ID, c.ID)
	}
}
