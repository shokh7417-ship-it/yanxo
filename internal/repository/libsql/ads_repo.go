package libsqlrepo

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"yanxo/internal/models"
	"yanxo/internal/repository"
)

type AdsRepo struct {
	db *sql.DB
}

func NewAdsRepo(db *sql.DB) *AdsRepo { return &AdsRepo{db: db} }

func (r *AdsRepo) Create(ctx context.Context, ad models.Ad) error {
	_, err := r.db.ExecContext(ctx, `
INSERT INTO ads(
  id,user_id,category,status,created_at,updated_at,
  from_city,to_city,ride_date,departure_time,car_type,total_seats,occupied_seats,
  service_type,area,note,
  contact,channel_message_id
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		ad.ID, ad.UserID, string(ad.Category), string(ad.Status), ad.CreatedAt.Format(time.RFC3339), ad.UpdatedAt.Format(time.RFC3339),
		ad.FromCity, ad.ToCity, ad.RideDate, ad.DepartureTime, ad.CarType, ad.TotalSeats, ad.OccupiedSeats,
		ad.ServiceType, ad.Area, ad.Note,
		ad.Contact, ad.ChannelMessageID,
	)
	return err
}

func (r *AdsRepo) GetByID(ctx context.Context, id string) (models.Ad, error) {
	var a models.Ad
	row := r.db.QueryRowContext(ctx, `
SELECT
  id,user_id,category,status,created_at,updated_at,
  from_city,to_city,ride_date,departure_time,car_type,total_seats,occupied_seats,
  service_type,area,note,
  contact,channel_message_id
FROM ads WHERE id = ?`, id)
	if err := scanAd(row, &a); err != nil {
		if err == sql.ErrNoRows {
			return a, repository.ErrNotFound
		}
		return a, err
	}
	return a, nil
}

func (r *AdsRepo) ListByUser(ctx context.Context, userID int64, category *models.AdCategory, statuses []models.AdStatus, limit int) ([]models.Ad, error) {
	if limit <= 0 {
		limit = 50
	}

	q := `
SELECT
  id,user_id,category,status,created_at,updated_at,
  from_city,to_city,ride_date,departure_time,car_type,total_seats,occupied_seats,
  service_type,area,note,
  contact,channel_message_id
FROM ads
WHERE user_id = ?`
	args := []any{userID}

	if category != nil {
		q += " AND category = ?"
		args = append(args, string(*category))
	}
	if len(statuses) > 0 {
		q += " AND status IN (" + strings.Repeat("?,", len(statuses)-1) + "?)"
		for _, s := range statuses {
			args = append(args, string(s))
		}
	}
	q += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Ad
	for rows.Next() {
		var a models.Ad
		if err := scanAd(rows, &a); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AdsRepo) SearchTaxiActive(ctx context.Context, fromCity, toCity string, nowLocalSQLite string, limit int) ([]models.Ad, error) {
	if limit <= 0 {
		limit = 20
	}

	// Rules:
	// - category=road
	// - status=active
	// - seats available
	// - departure datetime > now
	// - fallback daily expiry: treat as expired after 23:59 of ride_date (handled by date comparison below)
	rows, err := r.db.QueryContext(ctx, `
SELECT
  id,user_id,category,status,created_at,updated_at,
  from_city,to_city,ride_date,departure_time,car_type,total_seats,occupied_seats,
  service_type,area,note,
  contact,channel_message_id
FROM ads
WHERE category='road'
  AND status='active'
  AND from_city = ?
  AND to_city = ?
  AND (total_seats - occupied_seats) > 0
  AND datetime(ride_date || ' ' || departure_time || ':00') > datetime(?)
  AND date(ride_date) >= date(?)
ORDER BY ride_date ASC, departure_time ASC
LIMIT ?`,
		fromCity, toCity, nowLocalSQLite, nowLocalSQLite, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Ad
	for rows.Next() {
		var a models.Ad
		if err := scanAd(rows, &a); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AdsRepo) SearchServiceActive(ctx context.Context, serviceType, area string, limit int) ([]models.Ad, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT
  id,user_id,category,status,created_at,updated_at,
  from_city,to_city,ride_date,departure_time,car_type,total_seats,occupied_seats,
  service_type,area,note,
  contact,channel_message_id
FROM ads
WHERE category='service'
  AND status='active'
  AND service_type = ?
  AND area = ?
ORDER BY created_at DESC
LIMIT ?`, serviceType, area, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Ad
	for rows.Next() {
		var a models.Ad
		if err := scanAd(rows, &a); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AdsRepo) UpdateTaxiPassengerCount(ctx context.Context, id string, userID int64, occupied int, status models.AdStatus, updatedAtRFC3339 string) (models.Ad, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE ads
SET occupied_seats = ?, status = ?, updated_at = ?
WHERE id = ? AND user_id = ? AND category='road'`,
		occupied, string(status), updatedAtRFC3339, id, userID)
	if err != nil {
		return models.Ad{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		// either not found or not owned
		_, err := r.GetByID(ctx, id)
		if err != nil {
			return models.Ad{}, err
		}
		return models.Ad{}, repository.ErrForbidden
	}
	return r.GetByID(ctx, id)
}

func (r *AdsRepo) UpdateServiceFields(ctx context.Context, id string, userID int64, serviceType, area, note *string, contact *string, updatedAtRFC3339 string) (models.Ad, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE ads
SET service_type = ?, area = ?, note = ?, contact = ?, updated_at = ?
WHERE id = ? AND user_id = ? AND category='service'`,
		serviceType, area, note, contact, updatedAtRFC3339, id, userID)
	if err != nil {
		return models.Ad{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		_, err := r.GetByID(ctx, id)
		if err != nil {
			return models.Ad{}, err
		}
		return models.Ad{}, repository.ErrForbidden
	}
	return r.GetByID(ctx, id)
}

func (r *AdsRepo) UpdateStatus(ctx context.Context, id string, userID int64, status models.AdStatus, updatedAtRFC3339 string) (models.Ad, error) {
	res, err := r.db.ExecContext(ctx, `
UPDATE ads
SET status = ?, updated_at = ?
WHERE id = ? AND user_id = ?`,
		string(status), updatedAtRFC3339, id, userID)
	if err != nil {
		return models.Ad{}, err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		_, err := r.GetByID(ctx, id)
		if err != nil {
			return models.Ad{}, err
		}
		return models.Ad{}, repository.ErrForbidden
	}
	return r.GetByID(ctx, id)
}

func (r *AdsRepo) UpdateChannelMessageID(ctx context.Context, id string, userID int64, channelMessageID int, updatedAtRFC3339 string) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE ads
SET channel_message_id = ?, updated_at = ?
WHERE id = ? AND user_id = ?`,
		channelMessageID, updatedAtRFC3339, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		_, err := r.GetByID(ctx, id)
		if err != nil {
			return err
		}
		return repository.ErrForbidden
	}
	return nil
}

func (r *AdsRepo) MarkReplaced(ctx context.Context, id string, userID int64, updatedAtRFC3339 string) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE ads SET status='replaced', updated_at = ?
WHERE id = ? AND user_id = ?`, updatedAtRFC3339, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		_, err := r.GetByID(ctx, id)
		if err != nil {
			return err
		}
		return repository.ErrForbidden
	}
	return nil
}

func (r *AdsRepo) MarkDeleted(ctx context.Context, id string, userID int64, updatedAtRFC3339 string) error {
	res, err := r.db.ExecContext(ctx, `
UPDATE ads SET status='deleted', updated_at = ?
WHERE id = ? AND user_id = ?`, updatedAtRFC3339, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		_, err := r.GetByID(ctx, id)
		if err != nil {
			return err
		}
		return repository.ErrForbidden
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAd(s rowScanner, a *models.Ad) error {
	var (
		category string
		status   string
		created  string
		updated  string
	)
	err := s.Scan(
		&a.ID, &a.UserID, &category, &status, &created, &updated,
		&a.FromCity, &a.ToCity, &a.RideDate, &a.DepartureTime, &a.CarType, &a.TotalSeats, &a.OccupiedSeats,
		&a.ServiceType, &a.Area, &a.Note,
		&a.Contact, &a.ChannelMessageID,
	)
	if err != nil {
		return err
	}
	a.Category = models.AdCategory(category)
	a.Status = models.AdStatus(status)
	if t, err := time.Parse(time.RFC3339, created); err == nil {
		a.CreatedAt = t
	}
	if t, err := time.Parse(time.RFC3339, updated); err == nil {
		a.UpdatedAt = t
	}
	return nil
}

