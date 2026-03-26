package repository

import (
	"context"

	"yanxo/internal/models"
)

type AdsRepository interface {
	Create(ctx context.Context, ad models.Ad) error
	GetByID(ctx context.Context, id string) (models.Ad, error)
	ListByUser(ctx context.Context, userID int64, category *models.AdCategory, statuses []models.AdStatus, limit int) ([]models.Ad, error)

	// Taxi search
	// nowLocalSQLite must be in "YYYY-MM-DD HH:MM:SS" (local time) for SQLite datetime() comparison.
	SearchTaxiActive(ctx context.Context, fromCity, toCity string, nowLocalSQLite string, limit int) ([]models.Ad, error)
	SearchServiceActive(ctx context.Context, serviceType, area string, limit int) ([]models.Ad, error)

	// Updates
	UpdateTaxiPassengerCount(ctx context.Context, id string, userID int64, occupied int, status models.AdStatus, updatedAtRFC3339 string) (models.Ad, error)
	UpdateStatus(ctx context.Context, id string, userID int64, status models.AdStatus, updatedAtRFC3339 string) (models.Ad, error)
	UpdateChannelMessageID(ctx context.Context, id string, userID int64, channelMessageID int, updatedAtRFC3339 string) error
	UpdateServiceFields(ctx context.Context, id string, userID int64, serviceType, area, note *string, contact *string, updatedAtRFC3339 string) (models.Ad, error)

	// Major edit lifecycle
	MarkReplaced(ctx context.Context, id string, userID int64, updatedAtRFC3339 string) error
	MarkDeleted(ctx context.Context, id string, userID int64, updatedAtRFC3339 string) error
}

type UsersRepository interface {
	GetByTelegramID(ctx context.Context, telegramID int64) (models.User, error)
	UpsertTelegramUser(ctx context.Context, telegramID int64, username, firstName, lastName *string, updatedAtRFC3339 string) error
	SetRole(ctx context.Context, telegramID int64, role models.UserRole, updatedAtRFC3339 string) (models.User, error)
	ClearRole(ctx context.Context, telegramID int64, updatedAtRFC3339 string) (models.User, error)
}

