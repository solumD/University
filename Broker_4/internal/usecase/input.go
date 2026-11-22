package usecase

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"

	"gitlab.ozon.ru/university/broker-4/internal/model"
)

type input struct {
	ID          *int64  `json:"id,omitempty"`
	Name        *string `json:"name,omitempty"`
	ProjectID   *int64  `json:"project_id,omitempty"`
	Title       *string `json:"title,omitempty"`
	Completed   *bool   `json:"completed,omitempty"`
	Description *string `json:"description,omitempty"`
	Limit       *int    `json:"limit,omitempty"`
	Offset      *int    `json:"offset,omitempty"`
}

func prepare(cmd model.Command) (input, error) {
	var value input
	if cmd.Version != "v1" && cmd.Version != "v2" {
		err := model.Protocol("unsupported_version", "version must be v1 or v2")
		return value, err
	}

	allowed, err := fields(cmd.Action, cmd.Version)
	if err != nil {
		return value, err
	}

	var raw map[string]json.RawMessage
	if err = json.Unmarshal(cmd.Data, &raw); err != nil || raw == nil {
		err = model.Invalid("data must be a JSON object")
		return value, err
	}

	for name, field := range raw {
		optionalDefault := cmd.Action == "create_task" && (name == "completed" || name == "description")
		if !slices.Contains(allowed, name) || (bytes.Equal(bytes.TrimSpace(field), []byte("null")) && !optionalDefault) {
			err = model.Invalid("data contains an unsupported or null field")
			return value, err
		}
	}

	if err = json.Unmarshal(cmd.Data, &value); err != nil {
		err = model.Invalid("data contains an invalid field type")
		return value, err
	}

	if value.Name != nil {
		value.Name = new(strings.TrimSpace(*value.Name))
	}
	if value.Title != nil {
		value.Title = new(strings.TrimSpace(*value.Title))
	}
	err = value.validate(cmd)
	return value, err
}

func fields(action, version string) ([]string, error) {
	var names []string
	switch action {
	case "create_project":
		names = []string{"name"}
	case "update_project", "patch_project":
		names = []string{"id", "name"}
	case "get_project", "delete_project", "get_task", "delete_task":
		names = []string{"id"}
	case "create_task":
		names = []string{"project_id", "title", "completed", "description"}
	case "update_task", "patch_task":
		names = []string{"id", "project_id", "title", "completed", "description"}
	case "list_projects":
		names = []string{}
	case "list_tasks":
		names = []string{"project_id", "completed"}
	default:
		err := model.Protocol("unsupported_action", "action is not supported")
		return nil, err
	}

	if strings.HasPrefix(action, "list_") && version == "v2" {
		names = append(names, "limit", "offset")
	}
	return names, nil
}

func (in *input) validate(cmd model.Command) error {
	action := cmd.Action
	if !strings.HasPrefix(action, "create_") && !strings.HasPrefix(action, "list_") && (in.ID == nil || *in.ID < 1) {
		err := model.Invalid("id must be a positive integer")
		return err
	}

	if slices.Contains([]string{"create_project", "update_project", "patch_project"}, action) && in.Name == nil {
		err := model.Invalid("name is required")
		return err
	}

	if action == "create_task" || action == "update_task" {
		if in.ProjectID == nil || in.Title == nil || (action == "update_task" && in.Completed == nil) {
			err := model.Invalid("project_id and title are required; update_task also requires completed")
			return err
		}

		if in.Completed == nil {
			in.Completed = new(false)
		}
		if in.Description == nil {
			in.Description = new("")
		}
	}

	if action == "patch_task" && in.ProjectID == nil && in.Title == nil && in.Completed == nil && in.Description == nil {
		err := model.Invalid("patch_task requires at least one field besides id")
		return err
	}

	if in.ProjectID != nil && *in.ProjectID < 1 {
		err := model.Invalid("project_id must be positive")
		return err
	}

	if in.Limit != nil && (*in.Limit < 1 || *in.Limit > 100) || in.Offset != nil && *in.Offset < 0 {
		err := model.Invalid("limit must be 1 to 100 and offset must be nonnegative")
		return err
	}

	if strings.HasPrefix(action, "list_") && cmd.Version == "v2" {
		if in.Limit == nil {
			in.Limit = new(20)
		}
		if in.Offset == nil {
			in.Offset = new(0)
		}
	}
	return nil
}

func (in input) filter() model.Filter {
	filter := model.Filter{ProjectID: in.ProjectID, Completed: in.Completed, Limit: -1}
	if in.Limit != nil {
		filter.Limit = *in.Limit
	}
	if in.Offset != nil {
		filter.Offset = *in.Offset
	}
	return filter
}
