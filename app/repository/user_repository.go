package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"siakad-mini/app/model"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

type UserForLogin struct {
	model.User
	StudentDeleted bool
}

func (r *UserRepository) FindByEmailForLogin(ctx context.Context, email string) (UserForLogin, error) {
	var u UserForLogin
	var deletedAt *time.Time

	err := r.pool.QueryRow(ctx,
		`SELECT u.id, u.email, u.password, u.role, u.created_at, s.deleted_at
		 FROM users u
		 LEFT JOIN students s ON s.user_id = u.id
		 WHERE LOWER(u.email) = LOWER($1)`,
		email,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &deletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserForLogin{}, ErrNotFound
		}
		return UserForLogin{}, fmt.Errorf("mengambil user untuk login: %w", err)
	}

	u.StudentDeleted = deletedAt != nil
	return u, nil
}

// FindByID dipakai GET /auth/me.
func (r *UserRepository) FindByID(ctx context.Context, id int) (model.User, error) {
	var u model.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password, role, created_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, fmt.Errorf("mengambil user: %w", err)
	}
	return u, nil
}