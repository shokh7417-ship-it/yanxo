package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"yanxo/internal/config"
	"yanxo/internal/location"
	"yanxo/internal/service"
	"yanxo/internal/session"
)

type Context struct {
	Cfg      config.Config
	Bot      *tgbotapi.BotAPI
	Ads      *service.AdsService
	Store    *session.Store
	Resolver *location.Resolver
}

