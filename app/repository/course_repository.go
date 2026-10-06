package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
)

type CourseRepository struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) *CourseRepository {
	return &CourseRepository{pool: pool}
}

type CourseListQuery struct {
	Semester  int
	Search    string
	Available bool
}

// FindAll menghitung terisi & sisa_kuota lewat LEFT JOIN + COUNT, bukan
// query terpisah per course - satu query untuk seluruh daftar, bukan N+1.
func (r *CourseRepository) FindAll(ctx context.Context, q CourseListQuery) ([]model.Course, error) {
	where := " WHERE 1=1"
	args := []any{}

	if q.Semester != 0 {
		where += fmt.Sprintf(" AND c.semester = $%d", len(args)+1)
		args = append(args, q.Semester)
	}
	if q.Search != "" {
		where += fmt.Sprintf(" AND (c.kode_mk ILIKE $%d OR c.nama_mk ILIKE $%d)", len(args)+1, len(args)+1)
		args = append(args, "%"+q.Search+"%")
	}

	having := ""
	if q.Available {
		// HAVING, bukan WHERE - karena "terisi" adalah hasil agregasi
		// COUNT(e.id), belum ada di baris mentah sebelum GROUP BY.
		having = " HAVING COUNT(e.id) < c.kuota"
	}

	sqlText := fmt.Sprintf(
		`SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota, COUNT(e.id) AS terisi
		 FROM courses c
		 LEFT JOIN enrollments e ON e.course_id = c.id
		 %s
		 GROUP BY c.id
		 %s
		 ORDER BY c.kode_mk ASC`,
		where, having,
	)

	rows, err := r.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("mengambil daftar course: %w", err)
	}
	defer rows.Close()

	hasil := []model.Course{}
	for rows.Next() {
		var c model.Course
		if err := rows.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.Terisi); err != nil {
			return nil, fmt.Errorf("membaca baris course: %w", err)
		}
		c.SisaKuota = c.Kuota - c.Terisi
		hasil = append(hasil, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("membaca hasil query: %w", err)
	}
	return hasil, nil
}

func (r *CourseRepository) FindByID(ctx context.Context, id int) (model.Course, error) {
	var c model.Course
	err := r.pool.QueryRow(ctx,
		`SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota, COUNT(e.id)
		 FROM courses c LEFT JOIN enrollments e ON e.course_id = c.id
		 WHERE c.id = $1 GROUP BY c.id`, id,
	).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.Terisi)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Course{}, ErrNotFound
		}
		return model.Course{}, fmt.Errorf("mengambil course: %w", err)
	}
	c.SisaKuota = c.Kuota - c.Terisi
	return c, nil
}