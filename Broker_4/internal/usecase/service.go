package usecase

import (
	"context"

	"gitlab.ozon.ru/university/broker-4/internal/model"
)

type Repository interface {
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
}

type Transaction interface {
	Repository
	Cached(context.Context, string, string) (string, []byte, error)
	SaveCached(context.Context, string, string, string, []byte) error
	Commit() error
	Rollback() error
}

type Storage interface {
	Begin(context.Context) (Transaction, error)
}

type Service struct{ repo Repository }

func New(repo Repository) *Service {
	service := &Service{repo: repo}
	return service
}
