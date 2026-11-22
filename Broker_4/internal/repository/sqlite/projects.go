package sqlite

import (
	"context"
	"errors"

	"gitlab.ozon.ru/university/broker-4/internal/model"
)

func (r *Repository) Project(ctx context.Context, id int64) (model.Project, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var project model.Project

	err := r.db.QueryRowContext(ctx, `SELECT id, name FROM projects WHERE id = ?`, id).Scan(&project.ID, &project.Name)
	err = notFound(err, model.ErrProjectNotFound)

	return project, err
}

func (r *Repository) Projects(ctx context.Context, filter model.Filter) (items []model.Project, total int, err error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	if err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM projects`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `SELECT id, name FROM projects ORDER BY id LIMIT ? OFFSET ?`, filter.Limit, filter.Offset)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		closeErr := rows.Close()
		err = errors.Join(err, closeErr)
	}()

	capacity := max(0, total-filter.Offset)
	if filter.Limit >= 0 {
		capacity = min(capacity, filter.Limit)
	}

	items = make([]model.Project, 0, capacity)

	for rows.Next() {
		var project model.Project

		if err = rows.Scan(&project.ID, &project.Name); err != nil {
			return nil, 0, err
		}

		items = append(items, project)
	}

	err = rows.Err()

	return items, total, err
}

func (r *Repository) CreateProject(ctx context.Context, name string) (model.Project, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	result, err := r.db.ExecContext(ctx, `INSERT INTO projects (name) VALUES (?)`, name)
	if err != nil {
		return model.Project{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Project{}, err
	}

	project := model.Project{ID: id, Name: name}

	return project, nil
}

func (r *Repository) UpdateProject(ctx context.Context, project model.Project) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	result, err := r.db.ExecContext(ctx, `UPDATE projects SET name = ? WHERE id = ?`, project.Name, project.ID)
	err = affected(result, err, model.ErrProjectNotFound)

	return err
}

func (r *Repository) DeleteProject(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	result, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	err = affected(result, err, model.ErrProjectNotFound)

	return err
}

func (r *Repository) ProjectHasTasks(ctx context.Context, id int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var exists bool

	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM tasks WHERE project_id = ?)`, id).Scan(&exists)

	return exists, err
}
