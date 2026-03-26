package models

import "time"

type UserRole string

const (
	RoleProvider UserRole = "provider"
	RoleClient   UserRole = "client"
)

type User struct {
	TelegramID int64
	Username   *string
	FirstName  *string
	LastName   *string
	Role       *UserRole
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

