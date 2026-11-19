package usecase

import (
	"context"
	"strings"
	"unicode/utf8"

	"gitlab.ozon.ru/university/api-2/internal/model"
)

type TaskList struct {
	Items []model.Task
	Total int
}

func (s *Service) Task(ctx context.Context, id int64) (model.Task, error) {
	task, err := s.repo.Task(ctx, id)

	return task, err
}

func (s *Service) Tasks(ctx context.Context, filter model.Filter) (TaskList, error) {
	items, total, err := s.repo.Tasks(ctx, filter)
	if err != nil {
		return TaskList{}, err
	}

	list := TaskList{Items: items, Total: total}

	return list, nil
}

func (s *Service) CreateTask(ctx context.Context, input model.TaskInput, key model.CommandKey) (model.Task, error) {
	input.Title = strings.TrimSpace(input.Title)

	if err := validateTask(input); err != nil {
		return model.Task{}, err
	}

	var task model.Task

	fingerprint, found, err := s.loadCached(ctx, key, input, &task)
	if err != nil {
		return model.Task{}, err
	}

	if found {
		return task, nil
	}

	_, err = s.repo.Project(ctx, input.ProjectID)
	if err != nil {
		return model.Task{}, err
	}

	task, err = s.repo.CreateTask(ctx, input)
	if err != nil {
		return model.Task{}, err
	}

	err = s.saveCached(ctx, key, fingerprint, task)

	return task, err
}

func (s *Service) UpdateTask(ctx context.Context, id int64, input model.TaskInput) (model.Task, error) {
	input.Title = strings.TrimSpace(input.Title)

	if err := validateTask(input); err != nil {
		return model.Task{}, err
	}

	_, err := s.repo.Task(ctx, id)
	if err != nil {
		return model.Task{}, err
	}

	_, err = s.repo.Project(ctx, input.ProjectID)
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

	err = s.repo.UpdateTask(ctx, task)

	return task, err
}

func (s *Service) PatchTask(ctx context.Context, id int64, patch model.TaskPatch) (model.Task, error) {
	task, err := s.repo.Task(ctx, id)
	if err != nil {
		return model.Task{}, err
	}

	if patch.ProjectID != nil {
		if *patch.ProjectID < 1 {
			err = model.Invalid("project_id must be positive")

			return model.Task{}, err
		}

		_, err = s.repo.Project(ctx, *patch.ProjectID)
		if err != nil {
			return model.Task{}, err
		}

		task.ProjectID = *patch.ProjectID
	}

	if patch.Title != nil {
		title := strings.TrimSpace(*patch.Title)
		err = validateName(title)
		if err != nil {
			return model.Task{}, err
		}

		task.Title = title
	}

	if patch.Completed != nil {
		task.Completed = *patch.Completed
	}

	if patch.Description != nil {
		if utf8.RuneCountInString(*patch.Description) > 2000 {
			err = model.Invalid("description must not exceed 2000 characters")

			return model.Task{}, err
		}

		task.Description = *patch.Description
	}

	err = s.repo.UpdateTask(ctx, task)

	return task, err
}

func (s *Service) DeleteTask(ctx context.Context, id int64) error {
	err := s.repo.DeleteTask(ctx, id)

	return err
}

func validateTask(input model.TaskInput) error {
	if input.ProjectID < 1 {
		err := model.Invalid("project_id must be positive")

		return err
	}

	if err := validateName(input.Title); err != nil {
		return err
	}

	if utf8.RuneCountInString(input.Description) > 2000 {
		err := model.Invalid("description must not exceed 2000 characters")

		return err
	}

	return nil
}
