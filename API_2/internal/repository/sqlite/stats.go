package sqlite

import (
	"context"

	"gitlab.ozon.ru/university/api-2/internal/model"
)

func (r *Repository) Stats(ctx context.Context) (model.Stats, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	const query = `SELECT
		(SELECT COUNT(*) FROM projects),
		COUNT(*),
		COALESCE(SUM(completed), 0)
		FROM tasks`

	var stats model.Stats

	err := r.db.QueryRowContext(ctx, query).Scan(&stats.Projects, &stats.Tasks, &stats.CompletedTasks)

	return stats, err
}
