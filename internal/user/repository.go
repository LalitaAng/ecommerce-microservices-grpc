package user

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *User, passwordHash string) error {
	query := `
		INSERT INTO users (id, email, password, name, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query, user.ID, user.Email, passwordHash, user.Name, time.Now())
	return err
}

func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	return exists, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, string, error) {
	var user User
	var passwordHash string
	query := `
		SELECT id, email, name, created_at, password
		FROM users
		WHERE email = $1
	`

	if err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Name, &user.CreatedAt, &passwordHash,
	); err != nil {
		return nil, "", err
	}
	return &user, passwordHash, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	var user User
	query := `
		SELECT id, email, name, created_at
		FROM users
		WHERE id = $1
	`

	if err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Name, &user.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &user, nil
}
