package models

type SingleTodoModel struct {
	TaskName   string `json:"task_name"`
	TaskStatus string `json:"task_status"`
}

type AllTodoModel struct {
	Todos []SingleTodoModel `json:"todo_items"`
}
