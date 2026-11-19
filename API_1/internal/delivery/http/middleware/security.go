package middleware

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"gitlab.ozon.ru/university/api-1/internal/delivery/http/response"
	"go.uber.org/zap"
)

type clientKey struct{}

type bucket struct {
	used  int
	reset time.Time
}

type Security struct {
	digest   [32]byte
	clientID string
	limit    int
	log      *zap.Logger
	mu       sync.Mutex
	buckets  map[string]bucket
	cleanup  time.Time
}

func NewSecurity(apiKey string, limit int, log *zap.Logger) *Security {
	digest := sha256.Sum256([]byte(apiKey))
	clientID := hex.EncodeToString(digest[:])
	security := &Security{
		digest:   digest,
		clientID: clientID,
		limit:    limit,
		log:      log,
		buckets:  make(map[string]bucket),
	}

	return security
}

func (s *Security) valid(r *http.Request) bool {
	if len(r.Header.Values("X-API-Key")) != 1 {
		return false
	}

	digest := sha256.Sum256([]byte(r.Header.Get("X-API-Key")))
	valid := subtle.ConstantTimeCompare(digest[:], s.digest[:]) == 1

	return valid
}

func (s *Security) Authenticate(next http.Handler) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.valid(r) {
			response.Fail(w, http.StatusUnauthorized, "unauthorized", "a valid X-API-Key is required", s.log)

			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), clientKey{}, s.clientID)))
	})

	return handler
}

func (s *Security) RateLimit(next http.Handler) http.Handler {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity := "client:" + s.clientID
		if !s.valid(r) {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}

			identity = "ip:" + host
		}

		remaining, retry := s.take(identity, time.Now())
		w.Header().Set("X-Limit-Remaining", strconv.Itoa(remaining))
		if retry > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(retry))
			response.Fail(w, http.StatusTooManyRequests, "rate_limit_exceeded", "request limit exceeded", s.log)

			return
		}

		next.ServeHTTP(w, r)
	})

	return handler
}

func (s *Security) take(identity string, now time.Time) (int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !now.Before(s.cleanup) {
		for key, entry := range s.buckets {
			if !now.Before(entry.reset) {
				delete(s.buckets, key)
			}
		}

		s.cleanup = now.Add(time.Minute)
	}

	entry := s.buckets[identity]
	if !now.Before(entry.reset) {
		entry = bucket{reset: now.Add(time.Minute)}
	}

	if entry.used >= s.limit {
		retry := max(1, int((entry.reset.Sub(now)+time.Second-1)/time.Second))

		return 0, retry
	}

	entry.used++
	s.buckets[identity] = entry

	return s.limit - entry.used, 0
}
