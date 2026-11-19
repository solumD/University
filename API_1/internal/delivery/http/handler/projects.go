package handler

import (
	"net/http"

	"gitlab.ozon.ru/university/api-1/internal/delivery/http/dto"
	"gitlab.ozon.ru/university/api-1/internal/delivery/http/middleware"
	"gitlab.ozon.ru/university/api-1/internal/delivery/http/response"
	v2dto "gitlab.ozon.ru/university/api-1/internal/delivery/http/v2/dto"
	"gitlab.ozon.ru/university/api-1/internal/model"
)

func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	filter, err := h.filter(r, false)
	if err != nil {
		h.fail(w, err)

		return
	}

	list, err := h.service.Projects(r.Context(), filter)
	if err != nil {
		h.fail(w, err)

		return
	}

	if h.paginated {
		page := v2dto.Page[model.Project]{
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

func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	id, err := identifier(r)
	if err != nil {
		h.fail(w, err)

		return
	}

	project, err := h.service.Project(r.Context(), id)
	if err != nil {
		h.fail(w, err)

		return
	}

	response.JSON(w, http.StatusOK, project, h.log)
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	name, err := projectInput(w, r)
	if err != nil {
		h.fail(w, err)

		return
	}

	project, err := h.service.CreateProject(r.Context(), name, middleware.CommandKey(r))
	if err != nil {
		h.fail(w, err)

		return
	}

	h.created(w, r, project.ID, project)
}

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id, err := identifier(r)
	if err != nil {
		h.fail(w, err)

		return
	}

	name, err := projectInput(w, r)
	if err != nil {
		h.fail(w, err)

		return
	}

	project, err := h.service.UpdateProject(r.Context(), id, name)
	if err != nil {
		h.fail(w, err)

		return
	}

	response.JSON(w, http.StatusOK, project, h.log)
}

func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id, err := identifier(r)
	if err != nil {
		h.fail(w, err)

		return
	}

	if err := h.service.DeleteProject(r.Context(), id); err != nil {
		h.fail(w, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func projectInput(w http.ResponseWriter, r *http.Request) (string, error) {
	var input dto.ProjectRequest

	if err := decode(w, r, &input); err != nil {
		return "", err
	}

	if input.Name == nil {
		validationErr := model.Invalid("name is required")

		return "", validationErr
	}

	return *input.Name, nil
}
