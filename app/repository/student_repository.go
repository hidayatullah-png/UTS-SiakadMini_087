package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
	"siakad-mini/helper"
)

var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data sudah ada")
	ErrDuplicateNIM   = errors.New("nim sudah terdaftar")
	ErrDuplicateEmail = errors.New("email sudah terdaftar")
)

// whitelist kolom sort buat pertahanan terhadap SQL injection
var kolomUrutStudent = map[string]string{
	"nama":         "s.nama",
	"ipk_terakhir": "s.ipk_terakhir",
}

type StudentRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) *StudentRepository {
	return &StudentRepository{pool: pool}
}

// CreateWithUser membuat baris users (role mahasiswa, password = NIM yang
// di-hash) dan students jg dlm satu transaction, sesuai spek
// endpoint 4. kalo salah satu INSERT gagal, keduanya dibatalkan -> tidak boleh ada users tanpa students atau sebaliknya.
func (r *StudentRepository) CreateWithUser(
	ctx context.Context, req model.CreateStudentRequest,
) (model.Student, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Student{}, fmt.Errorf("memulai transaksi: %w", err)
	}
	defer tx.Rollback(ctx) // no-op bila sudah di-Commit

	passwordHash, err := helper.HashPassword(req.NIM) // password awal = NIM
	if err != nil {
		return model.Student{}, fmt.Errorf("hash password: %w", err)
	}

	var userID int
	err = tx.QueryRow(ctx,
		`INSERT INTO users (email, password, role) VALUES ($1, $2, 'mahasiswa')
		 RETURNING id`,
		req.Email, passwordHash,
	).Scan(&userID)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicateEmail // email duplikat
		}
		return model.Student{}, fmt.Errorf("membuat user: %w", err)
	}

	var s model.Student
	err = tx.QueryRow(ctx,
		`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir`,
		userID, req.NIM, req.Nama, req.Prodi, req.Angkatan, req.IPKTerakhir,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Student{}, ErrDuplicateNIM // nim duplikat
		}
		return model.Student{}, fmt.Errorf("membuat student: %w", err)
	}

	return s, tx.Commit(ctx)
}

// FindAll - seluruh query baca WAJIB menyertakan deleted_at IS NULL,
// supaya mahasiswa yang sudah soft-delete tidak pernah muncul di listing.
func (r *StudentRepository) FindAll(
	ctx context.Context, q model.ListQuery,
) ([]model.Student, int, error) {
	where := " WHERE s.deleted_at IS NULL"
	args := []any{}

	if q.Search != "" {
		where += fmt.Sprintf(" AND (s.nim ILIKE $%d OR s.nama ILIKE $%d)", len(args)+1, len(args)+1)
		args = append(args, "%"+q.Search+"%")
	}
	if q.Prodi != "" {
		where += fmt.Sprintf(" AND s.prodi = $%d", len(args)+1)
		args = append(args, q.Prodi)
	}
	if q.Angkatan != 0 {
		where += fmt.Sprintf(" AND s.angkatan = $%d", len(args)+1)
		args = append(args, q.Angkatan)
	}

	var total int
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM students s"+where, args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("menghitung student: %w", err)
	}

	sortCol := "s.nama"
	arah := "ASC"
	sortKey := q.Sort
	if len(sortKey) > 0 && sortKey[0] == '-' {
		arah = "DESC"
		sortKey = sortKey[1:]
	}
	if col, ok := kolomUrutStudent[sortKey]; ok {
		sortCol = col
	}

	sqlText := fmt.Sprintf(
		`SELECT s.id, s.user_id, s.nim, s.nama, s.prodi, s.angkatan, s.ipk_terakhir
		 FROM students s%s
		 ORDER BY %s %s
		 LIMIT $%d OFFSET $%d`,
		where, sortCol, arah, len(args)+1, len(args)+2,
	)
	args = append(args, q.PerPage, q.Offset())

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("mengambil daftar student: %w", err)
	}
	defer rows.Close()

	hasil := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir); err != nil {
			return nil, 0, fmt.Errorf("membaca baris student: %w", err)
		}
		hasil = append(hasil, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("membaca hasil query: %w", err)
	}

	return hasil, total, nil
}

func (r *StudentRepository) FindByID(ctx context.Context, id int) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir
		 FROM students WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil student: %w", err)
	}
	return s, nil
}

func (r *StudentRepository) Replace(
	ctx context.Context, id int, req model.ReplaceStudentRequest,
) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx,
		`UPDATE students SET nama=$1, prodi=$2, angkatan=$3, ipk_terakhir=$4
		 WHERE id=$5 AND deleted_at IS NULL
		 RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir`,
		req.Nama, req.Prodi, req.Angkatan, req.IPKTerakhir, id,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("memperbarui student: %w", err)
	}
	return s, nil
}

// SoftDelete - UPDATE, bukan DELETE. Syarat "deleted_at IS NULL" di WHERE
// mencegah soft-delete dua kali pada baris yang sudah terhapus (kalau
// sudah terhapus, RowsAffected jadi 0, dilaporkan sebagai 404 - bukan
// berhasil menghapus "lagi").
func (r *StudentRepository) SoftDelete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE students SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("soft delete student: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

func (r *StudentRepository) FindByUserID(ctx context.Context, userID int) (model.Student, error) {
	var s model.Student
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir
		 FROM students WHERE user_id = $1 AND deleted_at IS NULL`, userID,
	).Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("mengambil student milik user: %w", err)
	}
	return s, nil
}

func (r *StudentRepository) FindByIDWithSKS(ctx context.Context, id int) (model.Student, error) {
	s, err := r.FindByID(ctx, id)
	if err != nil {
		return model.Student{}, err
	}

	var totalSKS int
	if err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(c.sks), 0) FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1`, id,
	).Scan(&totalSKS); err != nil {
		return model.Student{}, fmt.Errorf("menghitung total sks: %w", err)
	}
	s.TotalSKS = totalSKS
	return s, nil
}
