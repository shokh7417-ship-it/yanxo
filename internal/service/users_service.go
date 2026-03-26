package service

import (
	"context"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"yanxo/internal/models"
	"yanxo/internal/repository"
)

type UsersService struct {
	repo repository.UsersRepository
}

func NewUsersService(repo repository.UsersRepository) *UsersService {
	return &UsersService{repo: repo}
}

func (s *UsersService) GetUserAndEnsureRole(ctx context.Context, tg *tgbotapi.User) (models.User, error) {
	username := strPtr(strings.TrimSpace(tg.UserName))
	firstName := strPtr(strings.TrimSpace(tg.FirstName))
	lastName := strPtr(strings.TrimSpace(tg.LastName))
	now := time.Now().Format(time.RFC3339)

	if err := s.repo.UpsertTelegramUser(ctx, tg.ID, username, firstName, lastName, now); err != nil {
		return models.User{}, err
	}
	return s.repo.GetByTelegramID(ctx, tg.ID)
}

func (s *UsersService) SetRole(ctx context.Context, telegramID int64, role models.UserRole) (models.User, error) {
	return s.repo.SetRole(ctx, telegramID, role, time.Now().Format(time.RFC3339))
}

func (s *UsersService) ClearRole(ctx context.Context, telegramID int64) (models.User, error) {
	return s.repo.ClearRole(ctx, telegramID, time.Now().Format(time.RFC3339))
}

func (s *UsersService) GetByTelegramID(ctx context.Context, telegramID int64) (models.User, error) {
	return s.repo.GetByTelegramID(ctx, telegramID)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

