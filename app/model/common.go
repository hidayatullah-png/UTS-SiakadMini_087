package model

type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

type Meta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	LastPage    int `json:"last_page"`
}

type ListQuery struct {
	Page    int
	PerPage int
	Search  string
	Prodi   string
	Angkatan int
	Sort    string // "nama" atau "-ipk_terakhir"
}

func (q ListQuery) Offset() int {
	return (q.Page - 1) * q.PerPage
}