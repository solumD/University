package dto

type ProjectRequest struct {
	Name *string `json:"name"`
}

type TaskRequest struct {
	ProjectID   *int64  `json:"project_id"`
	Title       *string `json:"title"`
	Completed   *bool   `json:"completed"`
	Description *string `json:"description"`
}
