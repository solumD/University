package model

type Project struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Task struct {
	ID          int64  `json:"id"`
	ProjectID   int64  `json:"project_id"`
	Title       string `json:"title"`
	Completed   bool   `json:"completed"`
	Description string `json:"description"`
}

type TaskInput struct {
	ProjectID   int64
	Title       string
	Completed   bool
	Description string
}

type TaskPatch struct {
	ProjectID   *int64
	Title       *string
	Completed   *bool
	Description *string
}

type Filter struct {
	ProjectID *int64
	Completed *bool
	Limit     int
	Offset    int
}

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

var (
	ErrProjectNotFound = &Error{Code: "project_not_found", Message: "project not found"}
	ErrTaskNotFound    = &Error{Code: "task_not_found", Message: "task not found"}
	ErrProjectNotEmpty = &Error{Code: "project_not_empty", Message: "project contains tasks"}
	ErrKeyConflict     = &Error{Code: "idempotency_conflict", Message: "idempotency key was used with different data"}
)

func Invalid(message string) error {
	err := &Error{Code: "invalid_request", Message: message}

	return err
}
