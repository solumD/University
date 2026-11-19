package usecase

import (
	"context"

	"gitlab.ozon.ru/university/api-2/internal/model"
)

func (s *Service) Stats(ctx context.Context) (model.Stats, error) {
	stats, err := s.repo.Stats(ctx)

	return stats, err
}
