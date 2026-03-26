package libsqlrepo

import (
	"context"
	"database/sql"

	"yanxo/internal/repository"
)

type LocationRepo struct {
	db *sql.DB
}

func NewLocationRepo(db *sql.DB) *LocationRepo { return &LocationRepo{db: db} }

func (r *LocationRepo) CanonicalByAlias(ctx context.Context, aliasNormalized string) (string, error) {
	var canonical string
	err := r.db.QueryRowContext(ctx, `
		SELECT l.name_canonical
		FROM location_aliases a
		JOIN locations l ON l.id = a.location_id
		WHERE a.alias_normalized = ?`, aliasNormalized).Scan(&canonical)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return canonical, nil
}

func (r *LocationRepo) AllCanonicals(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT name_canonical FROM locations ORDER BY name_canonical`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *LocationRepo) EnsureLocationWithAliases(ctx context.Context, canonical string, aliasNormalizedList []string) error {
	_, err := r.db.ExecContext(ctx, `INSERT OR IGNORE INTO locations (name_canonical) VALUES (?)`, canonical)
	if err != nil {
		return err
	}
	var id int64
	err = r.db.QueryRowContext(ctx, `SELECT id FROM locations WHERE name_canonical = ?`, canonical).Scan(&id)
	if err != nil {
		return err
	}
	for _, alias := range aliasNormalizedList {
		if alias == "" {
			continue
		}
		_, err = r.db.ExecContext(ctx, `INSERT OR IGNORE INTO location_aliases (location_id, alias_normalized) VALUES (?, ?)`, id, alias)
		if err != nil {
			return err
		}
	}
	return nil
}

var _ repository.LocationRepository = (*LocationRepo)(nil)
