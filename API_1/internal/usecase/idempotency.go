package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"gitlab.ozon.ru/university/api-1/internal/model"
)

func (s *Service) loadCached(ctx context.Context, key model.CommandKey, input, result any) (string, bool, error) {
	data, err := json.Marshal(input)
	if err != nil {
		return "", false, err
	}

	digest := sha256.Sum256(data)
	fingerprint := hex.EncodeToString(digest[:])

	previous, response, err := s.repo.Cached(ctx, key, time.Now().UTC())
	if err != nil {
		return "", false, err
	}

	if response == nil {
		return fingerprint, false, nil
	}

	if previous != fingerprint {
		return "", false, model.ErrKeyConflict
	}

	err = json.Unmarshal(response, result)
	if err != nil {
		return "", false, err
	}

	return fingerprint, true, nil
}

func (s *Service) saveCached(ctx context.Context, key model.CommandKey, fingerprint string, value any) error {
	response, err := json.Marshal(value)
	if err != nil {
		return err
	}

	expires := time.Now().UTC().AddDate(0, 0, 1)
	err = s.repo.SaveCached(ctx, key, fingerprint, response, expires)

	return err
}
