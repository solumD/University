package usecase

import (
	"context"
	"time"

	"gitlab.ozon.ru/university/api-2/internal/model"
)

type Repository interface {
	Stats(context.Context) (model.Stats, error)
	Project(context.Context, int64) (model.Project, error)
	Projects(context.Context, model.Filter) ([]model.Project, int, error)
	CreateProject(context.Context, string) (model.Project, error)
	UpdateProject(context.Context, model.Project) error
	DeleteProject(context.Context, int64) error
	ProjectHasTasks(context.Context, int64) (bool, error)
	Task(context.Context, int64) (model.Task, error)
	Tasks(context.Context, model.Filter) ([]model.Task, int, error)
	CreateTask(context.Context, model.TaskInput) (model.Task, error)
	UpdateTask(context.Context, model.Task) error
	DeleteTask(context.Context, int64) error
	Cached(context.Context, model.CommandKey, time.Time) (string, []byte, error)
	SaveCached(context.Context, model.CommandKey, string, []byte, time.Time) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	service := &Service{repo: repo}

	return service
}
