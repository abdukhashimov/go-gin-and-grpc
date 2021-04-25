package models

type TodoModel struct {
	TaskName   string `json:"task_name"`
	TaskStatus string `json:"task_status"`
}
