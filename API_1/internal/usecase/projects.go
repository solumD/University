package usecase

import (
	"context"
	"strings"
	"unicode/utf8"

	"gitlab.ozon.ru/university/api-1/internal/model"
)

type ProjectList struct {
	Items []model.Project
	Total int
}

func (s *Service) Project(ctx context.Context, id int64) (model.Project, error) {
	project, err := s.repo.Project(ctx, id)

	return project, err
}

func (s *Service) Projects(ctx context.Context, filter model.Filter) (ProjectList, error) {
	items, total, err := s.repo.Projects(ctx, filter)
	if err != nil {
		return ProjectList{}, err
	}

	list := ProjectList{Items: items, Total: total}

	return list, nil
}

func (s *Service) CreateProject(ctx context.Context, name string, key model.CommandKey) (model.Project, error) {
	name = strings.TrimSpace(name)

	if err := validateName(name); err != nil {
		return model.Project{}, err
	}

	var project model.Project

	fingerprint, found, err := s.loadCached(ctx, key, name, &project)
	if err != nil {
		return model.Project{}, err
	}

	if found {
		return project, nil
	}

	project, err = s.repo.CreateProject(ctx, name)
	if err != nil {
		return model.Project{}, err
	}

	err = s.saveCached(ctx, key, fingerprint, project)

	return project, err
}

func (s *Service) UpdateProject(ctx context.Context, id int64, name string) (model.Project, error) {
	name = strings.TrimSpace(name)

	if err := validateName(name); err != nil {
		return model.Project{}, err
	}

	project := model.Project{ID: id, Name: name}
	err := s.repo.UpdateProject(ctx, project)

	return project, err
}

func (s *Service) DeleteProject(ctx context.Context, id int64) error {
	exists, err := s.repo.ProjectHasTasks(ctx, id)
	if err != nil {
		return err
	}

	if exists {
		return model.ErrProjectNotEmpty
	}

	err = s.repo.DeleteProject(ctx, id)

	return err
}

func validateName(name string) error {
	if name == "" || utf8.RuneCountInString(name) > 200 {
		err := model.Invalid("name or title must contain between 1 and 200 characters")

		return err
	}

	return nil
}
