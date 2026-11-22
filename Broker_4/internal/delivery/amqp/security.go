package delivery

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.ozon.ru/university/broker-4/internal/model"
)

type security struct {
	digest [32]byte
	scope  string
	limit  int
	used   int
	reset  time.Time
}

func newSecurity(key string, limit int) *security {
	digest := sha256.Sum256([]byte(key))
	s := &security{digest: digest, scope: hex.EncodeToString(digest[:]), limit: limit}
	return s
}

func (s *security) check(key string) (amqp.Table, error) {
	digest := sha256.Sum256([]byte(key))
	if subtle.ConstantTimeCompare(digest[:], s.digest[:]) != 1 {
		err := model.Protocol("unauthorized", "a valid API key is required")
		return nil, err
	}

	now := time.Now()
	if !now.Before(s.reset) {
		s.used = 0
		s.reset = now.Add(time.Minute)
	}
	headers := amqp.Table{"X-Limit-Remaining": int32(max(0, s.limit-s.used))}
	if s.used >= s.limit {
		headers["Retry-After"] = int32(max(1, int(time.Until(s.reset).Seconds())+1))
		err := model.Protocol("rate_limit_exceeded", "request limit exceeded")
		return headers, err
	}

	s.used++
	headers["X-Limit-Remaining"] = int32(s.limit - s.used)
	return headers, nil
}
