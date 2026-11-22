package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"gitlab.ozon.ru/university/broker-4/config"
	"gitlab.ozon.ru/university/broker-4/internal/model"
)

type Executor struct{ storage Storage }

func NewExecutor(storage Storage) *Executor {
	executor := &Executor{storage: storage}
	return executor
}

func (e *Executor) Execute(ctx context.Context, scope string, cmd model.Command) (result any, err error) {
	in, err := prepare(cmd)
	if err != nil {
		return nil, err
	}

	normalized, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256(append([]byte(cmd.Version+" "+cmd.Action+" "), normalized...))
	fingerprint := hex.EncodeToString(digest[:])

	ctx, cancel := context.WithTimeout(ctx, config.OperationTimeout)
	defer cancel()
	tx, err := e.storage.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, tx.Rollback()) }()

	previous, response, err := tx.Cached(ctx, scope, cmd.ID)
	if err != nil {
		return nil, err
	}
	if response != nil {
		if previous != fingerprint {
			return nil, model.ErrKeyConflict
		}
		cachedResult := json.RawMessage(response)

		return cachedResult, nil
	}

	service := New(tx)
	result, err = service.dispatch(ctx, cmd, in)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(cmd.Action, "get_") && !strings.HasPrefix(cmd.Action, "list_") {
		response, err = json.Marshal(result)
		if err != nil {
			return nil, err
		}
		if err = tx.SaveCached(ctx, scope, cmd.ID, fingerprint, response); err != nil {
			return nil, err
		}
	}

	err = tx.Commit()
	return result, err
}

func (s *Service) dispatch(ctx context.Context, cmd model.Command, in input) (any, error) {
	var result any
	var err error
	switch cmd.Action {
	case "create_project":
		result, err = s.CreateProject(ctx, *in.Name)
	case "get_project":
		result, err = s.Project(ctx, *in.ID)
	case "update_project", "patch_project":
		result, err = s.UpdateProject(ctx, *in.ID, *in.Name)
	case "delete_project":
		err = s.DeleteProject(ctx, *in.ID)
	case "get_task":
		result, err = s.Task(ctx, *in.ID)
	case "delete_task":
		err = s.DeleteTask(ctx, *in.ID)
	case "create_task", "update_task":
		data := model.TaskInput{ProjectID: *in.ProjectID, Title: *in.Title, Completed: *in.Completed, Description: *in.Description}
		if cmd.Action == "create_task" {
			result, err = s.CreateTask(ctx, data)
			break
		}
		result, err = s.UpdateTask(ctx, *in.ID, data)
	case "patch_task":
		patch := model.TaskPatch{ProjectID: in.ProjectID, Title: in.Title, Completed: in.Completed, Description: in.Description}
		result, err = s.PatchTask(ctx, *in.ID, patch)
	case "list_projects":
		result, err = s.projectList(ctx, cmd.Version, in.filter())
	case "list_tasks":
		result, err = s.taskList(ctx, cmd.Version, in.filter())
	}
	return result, err
}

func (s *Service) projectList(ctx context.Context, version string, filter model.Filter) (any, error) {
	list, err := s.Projects(ctx, filter)
	if err != nil {
		return nil, err
	}
	if version == "v1" {
		return list.Items, nil
	}
	page := model.Page[model.Project]{Items: list.Items, Total: list.Total, Limit: filter.Limit, Offset: filter.Offset}
	return page, nil
}

func (s *Service) taskList(ctx context.Context, version string, filter model.Filter) (any, error) {
	list, err := s.Tasks(ctx, filter)
	if err != nil {
		return nil, err
	}
	if version == "v1" {
		return list.Items, nil
	}
	page := model.Page[model.Task]{Items: list.Items, Total: list.Total, Limit: filter.Limit, Offset: filter.Offset}
	return page, nil
}
