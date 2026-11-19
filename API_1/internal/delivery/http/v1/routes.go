package v1

import (
	"github.com/go-chi/chi/v5"
	"gitlab.ozon.ru/university/api-1/internal/delivery/http/handler"
	"gitlab.ozon.ru/university/api-1/internal/usecase"
	"go.uber.org/zap"
)

func Register(r chi.Router, service *usecase.Service, log *zap.Logger) {
	h := handler.New(service, log, false)

	r.Get("/projects", h.ListProjects)
	r.Post("/projects", h.CreateProject)
	r.Get("/projects/{id}", h.GetProject)
	r.Put("/projects/{id}", h.UpdateProject)
	r.Patch("/projects/{id}", h.UpdateProject)
	r.Delete("/projects/{id}", h.DeleteProject)

	r.Get("/tasks", h.ListTasks)
	r.Post("/tasks", h.CreateTask)
	r.Get("/tasks/{id}", h.GetTask)
	r.Put("/tasks/{id}", h.UpdateTask)
	r.Patch("/tasks/{id}", h.PatchTask)
	r.Delete("/tasks/{id}", h.DeleteTask)
}
