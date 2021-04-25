package models

//ResponseSuccess ...
type ResponseSuccess struct {
	Metadata interface{}
	Data     interface{}
}

//ResponseError ...
type ResponseError struct {
	Error interface{}
}

//InternalServerError ...
type InternalServerError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

//ValidationError ...
type ValidationError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	UserMessage string `json:"unread_message"`
}

type ResponseOK struct {
	Message interface{}
}

type Response struct {
	ID interface{} `json:"id"`
}

// Find query ...
type FindQueryModel struct {
	Page     int64  `json:"page,string"`
	Search   string `json:"search"`
	Active   bool   `json:"active"`
	Inactive bool   `josn:"inactive"`
	Limit    int64  `json:"limit,string"`
	Sort     string `json:"sort" example:"name|asc"`
	Lang     string `json:"lang"`
}

type AuthorizationModel struct {
	Token string `header:"Authorization"`
}

type UserInfo struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}
