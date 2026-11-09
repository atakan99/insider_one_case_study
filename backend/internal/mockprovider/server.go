package mockprovider

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/time/rate"
)

const maxRequestBody = 1 << 20

type server struct {
	o        Options
	store    *store
	picker   *picker
	limiters map[string]*rate.Limiter
}

// New returns the provider handler with all routes from the spec,
// including GET /healthz.
func New(o Options) http.Handler {
	if o.Now == nil {
		o.Now = time.Now
	}
	if o.Sleep == nil {
		o.Sleep = time.Sleep
	}
	s := &server{
		o:      o,
		store:  newStore(),
		picker: newPicker(o),
		limiters: map[string]*rate.Limiter{
			channelSMS:   rate.NewLimiter(rate.Limit(o.RatePerSec), o.Burst),
			channelEmail: rate.NewLimiter(rate.Limit(o.RatePerSec), o.Burst),
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/sms", func(w http.ResponseWriter, r *http.Request) {
		var req smsRequest
		s.send(w, r, channelSMS, &req, func() (Message, error) {
			if err := validateSMS(req); err != nil {
				return Message{}, err
			}
			return Message{Channel: channelSMS, To: req.To, Body: req.Body, Reference: req.Reference}, nil
		})
	})
	mux.HandleFunc("POST /v1/email", func(w http.ResponseWriter, r *http.Request) {
		var req emailRequest
		s.send(w, r, channelEmail, &req, func() (Message, error) {
			if err := validateEmail(req); err != nil {
				return Message{}, err
			}
			return Message{Channel: channelEmail, To: req.To, Subject: req.Subject, Body: req.Body, Reference: req.Reference}, nil
		})
	})
	mux.HandleFunc("GET /v1/messages", s.listMessages)
	mux.HandleFunc("DELETE /v1/messages", func(w http.ResponseWriter, _ *http.Request) {
		s.store.reset()
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /v1/stats", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, s.store.snapshot())
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return mux
}

// send runs the order of checks for one send request. req is the target
// for the JSON body. build validates req and returns the message to store.
func (s *server) send(w http.ResponseWriter, r *http.Request, channel string, req any, build func() (Message, error)) {
	s.store.count(channel, func(st *ChannelStats) { st.Requests++ })

	reject := func(msg string) {
		s.store.count(channel, func(st *ChannelStats) { st.Rejected++ })
		writeError(w, http.StatusBadRequest, "invalid_request", msg)
	}
	rateLimited := func(retryAfter int) {
		s.store.count(channel, func(st *ChannelStats) { st.RateLimited++ })
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		writeError(w, http.StatusTooManyRequests, "rate_limited", "too many requests for channel "+channel)
	}

	// 1. X-Mock-Fail.
	forced, isForced, err := parseForced(r.Header.Get("X-Mock-Fail"))
	if err != nil {
		reject(err.Error())
		return
	}

	// 2. Rate limit. A 429 uses no token.
	if forced == modeForced429 {
		rateLimited(1)
		return
	}
	now := s.o.Now()
	lim := s.limiters[channel]
	if !lim.AllowN(now, 1) {
		rateLimited(retryAfter(lim, now, s.o.RatePerSec))
		return
	}

	// 3. Parse and validate.
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBody))
	if err := dec.Decode(req); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			reject("request body is larger than 1 MiB")
			return
		}
		reject("request body is not valid JSON: " + err.Error())
		return
	}
	m, err := build()
	if err != nil {
		reject(err.Error())
		return
	}

	// 4. Failure mode.
	md := forced
	if !isForced {
		md = s.picker.pick()
	}

	// 5. Errors store nothing.
	switch md {
	case modeError500:
		s.store.count(channel, func(st *ChannelStats) { st.Errors++ })
		writeError(w, http.StatusInternalServerError, "provider_error", "internal provider error")
		return
	case modeError503:
		s.store.count(channel, func(st *ChannelStats) { st.Errors++ })
		writeError(w, http.StatusServiceUnavailable, "unavailable", "provider is unavailable, try again later")
		return
	}

	// 6. Wait. The wait and the store ignore r.Context(): if the client
	// leaves, the provider still delivers the message, like a real one.
	if md == modeSlow {
		s.store.count(channel, func(st *ChannelStats) { st.Slow++ })
		s.o.Sleep(s.o.SlowDelay)
	} else if s.o.LatencyMax > 0 {
		s.o.Sleep(s.picker.latency(s.o.LatencyMax))
	}

	// 7. Store.
	m.ReceivedAt = s.o.Now().UTC()
	m = s.store.add(m)

	// 8. Lost: close the connection with no response.
	if md == modeLost {
		s.store.count(channel, func(st *ChannelStats) { st.Lost++ })
		conn, _, err := http.NewResponseController(w).Hijack()
		if err == nil {
			_ = conn.Close()
			return
		}
		// No hijack support (HTTP/2). Abort the response instead.
		panic(http.ErrAbortHandler)
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"id": m.ID, "status": "accepted"})
}

// retryAfter returns the whole seconds until the limiter has one token,
// rounded up, with a minimum of 1.
func retryAfter(lim *rate.Limiter, now time.Time, perSec float64) int {
	wait := (1 - lim.TokensAt(now)) / perSec
	return max(1, int(math.Ceil(wait)))
}

func (s *server) listMessages(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := filter{channel: q.Get("channel"), to: q.Get("to"), reference: q.Get("reference"), limit: 100}
	if f.channel != "" && f.channel != channelSMS && f.channel != channelEmail {
		writeError(w, http.StatusBadRequest, "invalid_request", "channel must be sms or email")
		return
	}
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 1000 {
			writeError(w, http.StatusBadRequest, "invalid_request", "limit must be an integer from 1 to 1000")
			return
		}
		f.limit = n
	}
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			writeError(w, http.StatusBadRequest, "invalid_request", "offset must be an integer of 0 or more")
			return
		}
		f.offset = n
	}
	msgs, total := s.store.list(f)
	writeJSON(w, http.StatusOK, map[string]any{"total": total, "messages": msgs})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": msg}})
}
