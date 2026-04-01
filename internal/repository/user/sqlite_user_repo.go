package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/azharf99/wa-gateway/internal/domain"
)

type sqliteUserRepo struct {
	db *sql.DB
}

func NewSqliteUserRepository(db *sql.DB) domain.UserRepository {
	// Buat tabel users jika belum ada
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at DATETIME
	);`
	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}

	return &sqliteUserRepo{db: db}
}

func (r *sqliteUserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `SELECT id, username, password, created_at FROM users WHERE username = ?`
	row := r.db.QueryRowContext(ctx, query, username)

	var u domain.User
	err := row.Scan(&u.ID, &u.Username, &u.Password, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *sqliteUserRepo) Create(ctx context.Context, u *domain.User) error {
	query := `INSERT INTO users (username, password, created_at) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, u.Username, u.Password, time.Now())
	return err
}

func (r *sqliteUserRepo) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}
