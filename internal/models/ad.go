package models

import "time"

type AdCategory string

const (
	CategoryRoad    AdCategory = "road"
	CategoryService AdCategory = "service"
)

type AdStatus string

const (
	StatusActive   AdStatus = "active"
	StatusFull     AdStatus = "full"
	StatusExpired  AdStatus = "expired"
	StatusReplaced AdStatus = "replaced"
	StatusDeleted  AdStatus = "deleted"
)

type Ad struct {
	ID     string
	UserID int64

	Category AdCategory
	Status   AdStatus

	CreatedAt time.Time
	UpdatedAt time.Time

	// taxi fields
	FromCity      *string
	ToCity        *string
	RideDate      *string // YYYY-MM-DD
	DepartureTime *string // HH:MM
	CarType       *string
	TotalSeats    *int
	OccupiedSeats *int

	// service fields
	ServiceType *string
	Area        *string
	Note        *string

	// common
	Contact          *string
	ChannelMessageID *int
}

func (a Ad) AvailableSeats() *int {
	if a.TotalSeats == nil || a.OccupiedSeats == nil {
		return nil
	}
	v := *a.TotalSeats - *a.OccupiedSeats
	return &v
}

