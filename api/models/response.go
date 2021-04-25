package models

type ResponseError struct {
	Code    int    `json:"code" default:"0"`
	Message string `json:"message"`
	Reason  string `json:"reason"`
}

type Response struct {
	ID      string `json:"id"`
	Message string `json:"message"`
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
