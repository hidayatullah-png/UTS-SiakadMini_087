package model

import "time"

type Student struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	NIM         string     `json:"nim"`
	Nama        string     `json:"nama"`
	Prodi       string     `json:"prodi"`
	Angkatan    int        `json:"angkatan"`
	IPKTerakhir *float64   `json:"ipk_terakhir,omitempty"`
	DeletedAt   *time.Time `json:"-"`
	TotalSKS    int        `json:"total_sks,omitempty"` // diisi khusus endpoint detail (no.5)
	BatasSKS    int        `json:"batas_sks,omitempty"` // idem
}

type CreateStudentRequest struct {
	NIM         string   `json:"nim" validate:"required,len=12,numeric"`
	Nama        string   `json:"nama" validate:"required"`
	Email       string   `json:"email" validate:"required,email"`
	Prodi       string   `json:"prodi" validate:"required"`
	Angkatan    int      `json:"angkatan" validate:"required,angkatan_valid"`
	IPKTerakhir *float64 `json:"ipk_terakhir" validate:"omitnil,min=0,max=4"`
}

type ReplaceStudentRequest struct {
	Nama        string   `json:"nama" validate:"required"`
	Prodi       string   `json:"prodi" validate:"required"`
	Angkatan    int      `json:"angkatan" validate:"required"`
	IPKTerakhir *float64 `json:"ipk_terakhir" validate:"omitnil,min=0,max=4"`
}
