package handler

import (
	"net/http"

	"gitlab.ozon.ru/university/api-1/internal/delivery/http/dto"
	"gitlab.ozon.ru/university/api-1/internal/delivery/http/middleware"
	"gitlab.ozon.ru/university/api-1/internal/delivery/http/response"
	v2dto "gitlab.ozon.ru/university/api-1/internal/delivery/http/v2/dto"
	"gitlab.ozon.ru/university/api-1/internal/model"
)

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	filter, err := h.filter(r, true)
	if err != nil {
		h.fail(w, err)

		return
	}

	list, err := h.service.Tasks(r.Context(), filter)
	if err != nil {
		h.fail(w, err)

		return
	}

	if h.paginated {
		page := v2dto.Page[model.Task]{
			Items:  list.Items,
			Total:  list.Total,
			Limit:  filter.Limit,
			Offset: filter.Offset,
		}

		response.JSON(w, http.StatusOK, page, h.log)

		return
	}

	response.JSON(w, http.StatusOK, list.Items, h.log)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	id, err := identifier(r)
	if err != nil {
		h.fail(w, err)

		return
	}

	task, err := h.service.Task(r.Context(), id)
	if err != nil {
		h.fail(w, err)

		return
	}

	response.JSON(w, http.StatusOK, task, h.log)
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	input, err := taskInput(w, r, false)
	if err != nil {
		h.fail(w, err)

		return
	}

	task, err := h.service.CreateTask(r.Context(), input, middleware.CommandKey(r))
	if err != nil {
		h.fail(w, err)

		return
	}

	h.created(w, r, task.ID, task)
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id, err := identifier(r)
	if err != nil {
		h.fail(w, err)

		return
	}

	input, err := taskInput(w, r, true)
	if err != nil {
		h.fail(w, err)

		return
	}

	task, err := h.service.UpdateTask(r.Context(), id, input)
	if err != nil {
		h.fail(w, err)

		return
	}

	response.JSON(w, http.StatusOK, task, h.log)
}

func (h *Handler) PatchTask(w http.ResponseWriter, r *http.Request) {
	id, err := identifier(r)
	if err != nil {
		h.fail(w, err)

		return
	}

	var input dto.TaskRequest

	err = decodePatch(w, r, &input)
	if err != nil {
		h.fail(w, err)

		return
	}

	patch := model.TaskPatch{
		ProjectID:   input.ProjectID,
		Title:       input.Title,
		Completed:   input.Completed,
		Description: input.Description,
	}

	task, err := h.service.PatchTask(r.Context(), id, patch)
	if err != nil {
		h.fail(w, err)

		return
	}

	response.JSON(w, http.StatusOK, task, h.log)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := identifier(r)
	if err != nil {
		h.fail(w, err)

		return
	}

	if err := h.service.DeleteTask(r.Context(), id); err != nil {
		h.fail(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func taskInput(w http.ResponseWriter, r *http.Request, update bool) (model.TaskInput, error) {
	var input dto.TaskRequest

	if err := decode(w, r, &input); err != nil {
		return model.TaskInput{}, err
	}

	if input.ProjectID == nil || input.Title == nil {
		validationErr := model.Invalid("project_id and title are required")

		return model.TaskInput{}, validationErr
	}

	if update && input.Completed == nil {
		validationErr := model.Invalid("completed is required for PUT")

		return model.TaskInput{}, validationErr
	}

	result := model.TaskInput{ProjectID: *input.ProjectID, Title: *input.Title}
	if input.Completed != nil {
		result.Completed = *input.Completed
	}

	if input.Description != nil {
		result.Description = *input.Description
	}

	return result, nil
}
