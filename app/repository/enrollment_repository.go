package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
)

var (
	ErrKuotaPenuh  = errors.New("kuota mata kuliah penuh")
	ErrMelebihiSKS = errors.New("melebihi batas SKS")
)

type EnrollmentRepository struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) *EnrollmentRepository {
	return &EnrollmentRepository{pool: pool}
}

// Create
func (r *EnrollmentRepository) Create(
	ctx context.Context, studentID, courseID int, tahunAkademik string, batasSKS int,
) (model.Enrollment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Enrollment{}, fmt.Errorf("memulai transaksi: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1) Kunci baris course. FOR UPDATE tidak boleh dipakai bersama
	// GROUP BY/agregat, jadi mengunci dan menghitung dipisah.
	var sks, kuota int
	err = tx.QueryRow(ctx,
		`SELECT sks, kuota FROM courses WHERE id = $1 FOR UPDATE`,
		courseID).Scan(&sks, &kuota)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrNotFound
		}
		return model.Enrollment{}, fmt.Errorf("mengunci data course: %w", err)
	}

	// Hitung terisi SETELAH kunci didapat, di transaksi yang sama.
	var terisi int
	if err := tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM enrollments WHERE course_id = $1`,
		courseID).Scan(&terisi); err != nil {
		return model.Enrollment{}, fmt.Errorf("menghitung kuota terisi: %w", err)
	}
	if terisi >= kuota {
		return model.Enrollment{}, ErrKuotaPenuh
	}

	// 2) Total SKS yang sudah diambil mahasiswa ini pada tahun akademik ini.
	var totalSKS int
	if err := tx.QueryRow(ctx,
		`SELECT COALESCE(SUM(c.sks), 0) FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1 AND e.tahun_akademik = $2`,
		studentID, tahunAkademik).Scan(&totalSKS); err != nil {
		return model.Enrollment{}, fmt.Errorf("menghitung total SKS: %w", err)
	}
	if totalSKS+sks > batasSKS {
		return model.Enrollment{}, ErrMelebihiSKS
	}

	// 3) Insert - UNIQUE di database jadi penjaga terakhir terhadap duplikasi.
	var e model.Enrollment
	err = tx.QueryRow(ctx,
		`INSERT INTO enrollments (student_id, course_id, tahun_akademik)
		 VALUES ($1, $2, $3)
		 RETURNING id, student_id, course_id, tahun_akademik, created_at`,
		studentID, courseID, tahunAkademik,
	).Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return model.Enrollment{}, ErrDuplicate
		}
		return model.Enrollment{}, fmt.Errorf("mencatat enrollment: %w", err)
	}

	return e, tx.Commit(ctx)
}

func (r *EnrollmentRepository) FindByID(ctx context.Context, id int) (model.Enrollment, error) {
	var e model.Enrollment
	err := r.pool.QueryRow(ctx,
		`SELECT id, student_id, course_id, tahun_akademik, created_at
		 FROM enrollments WHERE id = $1`, id,
	).Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrNotFound
		}
		return model.Enrollment{}, fmt.Errorf("mengambil enrollment: %w", err)
	}
	return e, nil
}

func (r *EnrollmentRepository) TotalSKS(ctx context.Context, studentID int, tahunAkademik string) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(c.sks), 0) FROM enrollments e
		 JOIN courses c ON c.id = e.course_id
		 WHERE e.student_id = $1 AND e.tahun_akademik = $2`,
		studentID, tahunAkademik).Scan(&total)
	return total, err
}

// Delete
func (r *EnrollmentRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM enrollments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("menghapus enrollment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
