package sqlite

import (
	"context"
	"errors"
	"strings"

	"gitlab.ozon.ru/university/broker-4/internal/model"
)

func (r *Repository) Task(ctx context.Context, id int64) (model.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var task model.Task

	err := r.db.QueryRowContext(ctx, `SELECT id, project_id, title, completed, description FROM tasks WHERE id = ?`, id).
		Scan(&task.ID, &task.ProjectID, &task.Title, &task.Completed, &task.Description)
	err = notFound(err, model.ErrTaskNotFound)

	return task, err
}

func (r *Repository) Tasks(ctx context.Context, filter model.Filter) (items []model.Task, total int, err error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	where, args := taskFilter(filter)

	if err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tasks`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, filter.Limit, filter.Offset)
	rows, err := r.db.QueryContext(ctx, `SELECT id, project_id, title, completed, description FROM tasks`+where+` ORDER BY id LIMIT ? OFFSET ?`, args...)
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

	items = make([]model.Task, 0, capacity)

	for rows.Next() {
		var task model.Task

		if err = rows.Scan(&task.ID, &task.ProjectID, &task.Title, &task.Completed, &task.Description); err != nil {
			return nil, 0, err
		}

		items = append(items, task)
	}

	err = rows.Err()

	return items, total, err
}

func taskFilter(filter model.Filter) (string, []any) {
	conditions := make([]string, 0, 2)
	args := make([]any, 0, 4)
	if filter.ProjectID != nil {
		conditions = append(conditions, "project_id = ?")
		args = append(args, *filter.ProjectID)
	}

	if filter.Completed != nil {
		conditions = append(conditions, "completed = ?")
		args = append(args, *filter.Completed)
	}

	if len(conditions) == 0 {
		return "", args
	}

	where := " WHERE " + strings.Join(conditions, " AND ")

	return where, args
}

func (r *Repository) CreateTask(ctx context.Context, input model.TaskInput) (model.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	result, err := r.db.ExecContext(ctx, `INSERT INTO tasks (project_id, title, completed, description) VALUES (?, ?, ?, ?)`,
		input.ProjectID, input.Title, input.Completed, input.Description)
	if err != nil {
		return model.Task{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Task{}, err
	}

	task := model.Task{
		ID:          id,
		ProjectID:   input.ProjectID,
		Title:       input.Title,
		Completed:   input.Completed,
		Description: input.Description,
	}

	return task, nil
}

func (r *Repository) UpdateTask(ctx context.Context, task model.Task) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	result, err := r.db.ExecContext(ctx, `UPDATE tasks SET project_id = ?, title = ?, completed = ?, description = ? WHERE id = ?`,
		task.ProjectID, task.Title, task.Completed, task.Description, task.ID)
	err = affected(result, err, model.ErrTaskNotFound)

	return err
}

func (r *Repository) DeleteTask(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	result, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, id)
	err = affected(result, err, model.ErrTaskNotFound)

	return err
}
