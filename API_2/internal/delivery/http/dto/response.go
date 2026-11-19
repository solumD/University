package dto

import "gitlab.ozon.ru/university/api-2/internal/model"

func ProjectView(project model.Project, fields []string) any {
	if len(fields) == 0 {
		return project
	}

	values := map[string]any{
		"id":   project.ID,
		"name": project.Name,
	}

	result := selectFields(values, fields)

	return result
}

func TaskView(task model.Task, fields []string) any {
	if len(fields) == 0 {
		return task
	}

	values := map[string]any{
		"id":          task.ID,
		"project_id":  task.ProjectID,
		"title":       task.Title,
		"completed":   task.Completed,
		"description": task.Description,
	}

	result := selectFields(values, fields)

	return result
}

func selectFields(values map[string]any, fields []string) map[string]any {
	result := make(map[string]any, len(fields))

	for _, field := range fields {
		result[field] = values[field]
	}

	return result
}
