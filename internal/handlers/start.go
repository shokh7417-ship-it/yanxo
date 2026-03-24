package handlers

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"yanxo/internal/session"
	"yanxo/internal/templates"
)

type StartHandler struct {
	ctx Context
}

func NewStartHandler(ctx Context) *StartHandler { return &StartHandler{ctx: ctx} }

func (h *StartHandler) Start(ctx context.Context, m *tgbotapi.Message) error {
	h.ctx.Store.Clear(m.From.ID)
	// Hard reset UX: drop any previously shown custom keyboard before showing the fresh main menu.
	clearKB := tgbotapi.NewMessage(m.Chat.ID, "🔄 Qayta boshlandi.")
	clearKB.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
	_, _ = h.ctx.Bot.Send(clearKB)

	msg := tgbotapi.NewMessage(m.Chat.ID, templates.WelcomeText())
	msg.ReplyMarkup = templates.MainMenuKeyboard()
	_, err := h.ctx.Bot.Send(msg)
	return err
}

func (h *StartHandler) Cancel(ctx context.Context, m *tgbotapi.Message) error {
	h.ctx.Store.Clear(m.From.ID)
	msg := tgbotapi.NewMessage(m.Chat.ID, "Bekor qilindi. Asosiy menyu.")
	msg.ReplyMarkup = templates.MainMenuKeyboard()
	_, err := h.ctx.Bot.Send(msg)
	return err
}

func (h *StartHandler) RouteMenu(ctx context.Context, m *tgbotapi.Message) error {
	text := strings.TrimSpace(m.Text)
	switch text {
	case templates.BtnTaxiCreate:
		st := session.State{Flow: session.FlowTaxiCreate, Step: session.StepTaxiFromCity}
		st.Taxi.TotalSeats = 4
		st.Taxi.OccupiedSeats = 0
		h.ctx.Store.Set(m.From.ID, st)
		msg := tgbotapi.NewMessage(m.Chat.ID, "Qayerdan? (shahar)")
		msg.ReplyMarkup = templates.CityKeyboard()
		_, err := h.ctx.Bot.Send(msg)
		return err
	case templates.BtnServiceCreate:
		h.ctx.Store.Set(m.From.ID, session.State{Flow: session.FlowServiceCreate, Step: session.StepServiceCategory})
		msg := tgbotapi.NewMessage(m.Chat.ID, "Xizmat yo‘nalishini tanlang:")
		msg.ReplyMarkup = templates.ServiceCategoryKeyboard()
		_, err := h.ctx.Bot.Send(msg)
		return err
	case templates.BtnSearch:
		msg := tgbotapi.NewMessage(m.Chat.ID, "Nimani qidiramiz?")
		msg.ReplyMarkup = templates.SearchMenuKeyboard()
		_, err := h.ctx.Bot.Send(msg)
		return err
	case templates.BtnMyAds:
		return NewMyAdsHandler(h.ctx).Show(ctx, m)
	case templates.BtnOpenChannel:
		url := strings.TrimSpace(h.ctx.Cfg.ChannelURL)
		if url == "" {
			// fallback (no public username available)
			link := templates.ChannelLinkHint(h.ctx.Cfg.ChannelID)
			_, err := h.ctx.Bot.Send(tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("Kanal: %s", link)))
			return err
		}
		msg := tgbotapi.NewMessage(m.Chat.ID, "Kanalni ochish:")
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("📢 Kanalni ochish", url),
			),
		)
		_, err := h.ctx.Bot.Send(msg)
		return err
	default:
		if text == templates.BtnSearchTaxi {
			st := session.State{Flow: session.FlowTaxiSearch, Step: session.StepTaxiSearchFrom}
			h.ctx.Store.Set(m.From.ID, st)
			msg := tgbotapi.NewMessage(m.Chat.ID, "Qayerdan? (shahar)")
			msg.ReplyMarkup = templates.CityKeyboard()
			_, err := h.ctx.Bot.Send(msg)
			return err
		}
		if text == templates.BtnSearchService {
			st := session.State{Flow: session.FlowServiceSearch, Step: session.StepServiceSearchCategory}
			h.ctx.Store.Set(m.From.ID, st)
			msg := tgbotapi.NewMessage(m.Chat.ID, "Qidirish: xizmat yo‘nalishini tanlang:")
			msg.ReplyMarkup = templates.ServiceCategoryKeyboard()
			_, err := h.ctx.Bot.Send(msg)
			return err
		}
		if text == templates.BtnBack {
			msg := tgbotapi.NewMessage(m.Chat.ID, "Asosiy menyu.")
			msg.ReplyMarkup = templates.MainMenuKeyboard()
			_, err := h.ctx.Bot.Send(msg)
			return err
		}

		// If user is in an active session, we will handle later; for now show menu.
		msg := tgbotapi.NewMessage(m.Chat.ID, "Menyudan tanlang yoki /start bosing.")
		msg.ReplyMarkup = templates.MainMenuKeyboard()
		_, err := h.ctx.Bot.Send(msg)
		return err
	}
}

func (h *StartHandler) RouteCallback(ctx context.Context, q *tgbotapi.CallbackQuery, m *tgbotapi.Message) error {
	// Placeholder for inline buttons later (my ads management, confirm, etc.)
	_ = q
	msg := tgbotapi.NewMessage(m.Chat.ID, "Hali inline tugmalar ulanmagan. /start")
	msg.ReplyMarkup = templates.MainMenuKeyboard()
	_, err := h.ctx.Bot.Send(msg)
	return err
}

