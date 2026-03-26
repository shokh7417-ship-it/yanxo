package libsqlrepo

import (
	"context"
	"database/sql"
	"time"

	"yanxo/internal/models"
	"yanxo/internal/repository"
)

type UsersRepo struct {
	db *sql.DB
}

func NewUsersRepo(db *sql.DB) *UsersRepo { return &UsersRepo{db: db} }

func (r *UsersRepo) GetByTelegramID(ctx context.Context, telegramID int64) (models.User, error) {
	var u models.User
	var role *string
	var created, updated string

	row := r.db.QueryRowContext(ctx, `
SELECT telegram_id, username, first_name, last_name, role, created_at, updated_at
FROM users
WHERE telegram_id = ?`, telegramID)

	if err := row.Scan(&u.TelegramID, &u.Username, &u.FirstName, &u.LastName, &role, &created, &updated); err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, repository.ErrNotFound
		}
		return models.User{}, err
	}
	if role != nil && *role != "" {
		rv := models.UserRole(*role)
		u.Role = &rv
	}
	if t, err := time.Parse(time.RFC3339, created); err == nil {
		u.CreatedAt = t
	}
	if t, err := time.Parse(time.RFC3339, updated); err == nil {
		u.UpdatedAt = t
	}
	return u, nil
}

func (r *UsersRepo) UpsertTelegramUser(ctx context.Context, telegramID int64, username, firstName, lastName *string, updatedAtRFC3339 string) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO users (telegram_id, username, first_name, last_name, role, created_at, updated_at)
VALUES (?, ?, ?, ?, NULL, ?, ?)
ON CONFLICT(telegram_id) DO UPDATE SET
  username=excluded.username,
  first_name=excluded.first_name,
  last_name=excluded.last_name,
  updated_at=excluded.updated_at`,
		telegramID, username, firstName, lastName, updatedAtRFC3339, updatedAtRFC3339)
	return err
}

func (r *UsersRepo) SetRole(ctx context.Context, telegramID int64, role models.UserRole, updatedAtRFC3339 string) (models.User, error) {
	res, err := r.db.ExecContext(ctx, `UPDATE users SET role = ?, updated_at = ? WHERE telegram_id = ?`,
		string(role), updatedAtRFC3339, telegramID)
	if err != nil {
		return models.User{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return models.User{}, repository.ErrNotFound
	}
	return r.GetByTelegramID(ctx, telegramID)
}

func (r *UsersRepo) ClearRole(ctx context.Context, telegramID int64, updatedAtRFC3339 string) (models.User, error) {
	res, err := r.db.ExecContext(ctx, `UPDATE users SET role = NULL, updated_at = ? WHERE telegram_id = ?`,
		updatedAtRFC3339, telegramID)
	if err != nil {
		return models.User{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return models.User{}, repository.ErrNotFound
	}
	return r.GetByTelegramID(ctx, telegramID)
}

