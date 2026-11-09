package mockprovider

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

// mode is what the provider does with one valid request.
type mode int

const (
	modeNone mode = iota
	modeError500
	modeError503
	modeSlow
	modeLost
	modeForced429 // only from X-Mock-Fail
)

func (m mode) String() string {
	switch m {
	case modeNone:
		return "none"
	case modeError500:
		return "500"
	case modeError503:
		return "503"
	case modeSlow:
		return "slow"
	case modeLost:
		return "lost"
	case modeForced429:
		return "429"
	}
	return fmt.Sprintf("mode(%d)", int(m))
}

// parseForced reads the X-Mock-Fail header. ok is false when the header is
// empty, so the random picker decides.
func parseForced(header string) (m mode, ok bool, err error) {
	switch header {
	case "":
		return modeNone, false, nil
	case "none":
		return modeNone, true, nil
	case "429":
		return modeForced429, true, nil
	case "500":
		return modeError500, true, nil
	case "503":
		return modeError503, true, nil
	case "slow":
		return modeSlow, true, nil
	case "lost":
		return modeLost, true, nil
	}
	return modeNone, false, fmt.Errorf("X-Mock-Fail must be one of none, 429, 500, 503, slow, lost, got %q", header)
}

// picker selects random failure modes and latencies from one seeded source.
type picker struct {
	mu                            sync.Mutex
	rnd                           *rand.Rand
	errorRate, slowRate, lostRate float64
}

func newPicker(o Options) *picker {
	return &picker{
		rnd:       rand.New(rand.NewPCG(o.Random, 0)),
		errorRate: o.ErrorRate,
		slowRate:  o.SlowRate,
		lostRate:  o.LostRate,
	}
}

// pick selects a mode by cumulative rate. 500 and 503 share the error rate.
func (p *picker) pick() mode {
	p.mu.Lock()
	defer p.mu.Unlock()
	x := p.rnd.Float64()
	switch {
	case x < p.errorRate:
		if p.rnd.IntN(2) == 0 {
			return modeError500
		}
		return modeError503
	case x < p.errorRate+p.slowRate:
		return modeSlow
	case x < p.errorRate+p.slowRate+p.lostRate:
		return modeLost
	}
	return modeNone
}

// latency returns a random duration from 0 to max.
func (p *picker) latency(max time.Duration) time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	return time.Duration(p.rnd.Int64N(int64(max) + 1))
}
