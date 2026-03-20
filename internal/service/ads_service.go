package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"yanxo/internal/models"
	"yanxo/internal/repository"
	"yanxo/internal/utils"
)

type AdsService struct {
	repo  repository.AdsRepository
	clock utils.Clock
}

func NewAdsService(repo repository.AdsRepository, clock utils.Clock) *AdsService {
	return &AdsService{repo: repo, clock: clock}
}

func (s *AdsService) CreateTaxi(ctx context.Context, userID int64, fromCity, toCity, rideDate, departureTime, carType string, totalSeats, occupiedSeats int, contact *string) (models.Ad, error) {
	now := s.clock.Now()
	ad := models.Ad{
		ID:            uuid.NewString(),
		UserID:        userID,
		Category:      models.CategoryRoad,
		Status:        models.StatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
		FromCity:      &fromCity,
		ToCity:        &toCity,
		RideDate:      &rideDate,
		DepartureTime: &departureTime,
		CarType:       &carType,
		TotalSeats:    &totalSeats,
		OccupiedSeats: &occupiedSeats,
		Contact:       contact,
	}
	ad.Status = ComputeTaxiStatus(now, ad)
	if err := s.repo.Create(ctx, ad); err != nil {
		return models.Ad{}, err
	}
	return ad, nil
}

func (s *AdsService) CreateService(ctx context.Context, userID int64, serviceType, area string, note *string, contact *string) (models.Ad, error) {
	now := s.clock.Now()
	ad := models.Ad{
		ID:          uuid.NewString(),
		UserID:      userID,
		Category:    models.CategoryService,
		Status:      models.StatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
		ServiceType: &serviceType,
		Area:        &area,
		Note:        note,
		Contact:     contact,
	}
	if err := s.repo.Create(ctx, ad); err != nil {
		return models.Ad{}, err
	}
	return ad, nil
}

func (s *AdsService) NowLocalSQLite() string {
	// SQLite datetime() compares best with "YYYY-MM-DD HH:MM:SS"
	return s.clock.Now().In(time.Local).Format("2006-01-02 15:04:05")
}

func (s *AdsService) SearchTaxi(ctx context.Context, fromCity, toCity string, limit int) ([]models.Ad, error) {
	return s.repo.SearchTaxiActive(ctx, fromCity, toCity, s.NowLocalSQLite(), limit)
}

func (s *AdsService) SearchService(ctx context.Context, serviceType, area string, limit int) ([]models.Ad, error) {
	return s.repo.SearchServiceActive(ctx, serviceType, area, limit)
}

func (s *AdsService) ListByUser(ctx context.Context, userID int64, category *models.AdCategory, statuses []models.AdStatus, limit int) ([]models.Ad, error) {
	return s.repo.ListByUser(ctx, userID, category, statuses, limit)
}

func (s *AdsService) Get(ctx context.Context, id string) (models.Ad, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AdsService) UpdateTaxiOccupiedDelta(ctx context.Context, id string, userID int64, delta int) (models.Ad, error) {
	ad, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return models.Ad{}, err
	}
	if ad.UserID != userID {
		return models.Ad{}, repository.ErrForbidden
	}
	if ad.Category != models.CategoryRoad {
		return models.Ad{}, errors.New("not a taxi ad")
	}
	if ad.TotalSeats == nil || ad.OccupiedSeats == nil {
		return models.Ad{}, errors.New("invalid taxi ad seats")
	}

	newOcc := *ad.OccupiedSeats + delta
	if newOcc < 0 {
		newOcc = 0
	}
	if newOcc > *ad.TotalSeats {
		newOcc = *ad.TotalSeats
	}

	updated := s.clock.Now()
	tmp := ad
	tmp.OccupiedSeats = &newOcc
	st := ComputeTaxiStatus(updated, tmp)
	return s.repo.UpdateTaxiPassengerCount(ctx, id, userID, newOcc, st, updated.Format(time.RFC3339))
}

func (s *AdsService) SetTaxiFull(ctx context.Context, id string, userID int64) (models.Ad, error) {
	ad, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return models.Ad{}, err
	}
	if ad.UserID != userID {
		return models.Ad{}, repository.ErrForbidden
	}
	if ad.TotalSeats == nil {
		return models.Ad{}, errors.New("invalid total seats")
	}
	occ := *ad.TotalSeats
	updated := s.clock.Now()
	return s.repo.UpdateTaxiPassengerCount(ctx, id, userID, occ, models.StatusFull, updated.Format(time.RFC3339))
}

func (s *AdsService) SetStatus(ctx context.Context, id string, userID int64, status models.AdStatus) (models.Ad, error) {
	updated := s.clock.Now()
	return s.repo.UpdateStatus(ctx, id, userID, status, updated.Format(time.RFC3339))
}

func (s *AdsService) MarkReplaced(ctx context.Context, id string, userID int64) error {
	return s.repo.MarkReplaced(ctx, id, userID, s.clock.Now().Format(time.RFC3339))
}

func (s *AdsService) MarkDeleted(ctx context.Context, id string, userID int64) error {
	return s.repo.MarkDeleted(ctx, id, userID, s.clock.Now().Format(time.RFC3339))
}

func (s *AdsService) UpdateChannelMessageID(ctx context.Context, id string, userID int64, channelMsgID int) error {
	return s.repo.UpdateChannelMessageID(ctx, id, userID, channelMsgID, s.clock.Now().Format(time.RFC3339))
}

func (s *AdsService) UpdateServiceFields(ctx context.Context, id string, userID int64, serviceType, area string, note *string, contact *string) (models.Ad, error) {
	return s.repo.UpdateServiceFields(ctx, id, userID, &serviceType, &area, note, contact, s.clock.Now().Format(time.RFC3339))
}

func ComputeTaxiStatus(now time.Time, ad models.Ad) models.AdStatus {
	// Deleted / replaced are preserved
	if ad.Status == models.StatusDeleted || ad.Status == models.StatusReplaced {
		return ad.Status
	}
	if ad.TotalSeats != nil && ad.OccupiedSeats != nil {
		if (*ad.TotalSeats - *ad.OccupiedSeats) <= 0 {
			return models.StatusFull
		}
	}
	// Expire if departure passed OR ride_date in past (daily fallback)
	if isTaxiExpired(now, ad) {
		return models.StatusExpired
	}
	return models.StatusActive
}

func isTaxiExpired(now time.Time, ad models.Ad) bool {
	if ad.RideDate == nil || ad.DepartureTime == nil {
		return false
	}
	dep, err := parseLocalDeparture(*ad.RideDate, *ad.DepartureTime)
	if err == nil && !dep.After(now.In(time.Local)) {
		return true
	}
	// fallback: after 23:59 of ride_date, it must be unavailable next day
	today := now.In(time.Local).Format("2006-01-02")
	if *ad.RideDate < today {
		return true
	}
	return false
}

func parseLocalDeparture(rideDate, departureTime string) (time.Time, error) {
	// rideDate: YYYY-MM-DD, departureTime: HH:MM
	if len(rideDate) != 10 || len(departureTime) != 5 {
		return time.Time{}, fmt.Errorf("bad date/time")
	}
	hh, err := strconv.Atoi(departureTime[0:2])
	if err != nil {
		return time.Time{}, err
	}
	mm, err := strconv.Atoi(departureTime[3:5])
	if err != nil {
		return time.Time{}, err
	}
	d, err := time.ParseInLocation("2006-01-02", rideDate, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(d.Year(), d.Month(), d.Day(), hh, mm, 0, 0, time.Local), nil
}

