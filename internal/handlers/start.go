package handlers

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"yanxo/internal/session"
	"yanxo/internal/templates"
	"yanxo/internal/models"
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

	user, err := h.GetUserAndEnsureRole(ctx, m)
	if err != nil {
		return err
	}
	if user.Role == nil || *user.Role == "" {
		return h.showRoleSelection(m.Chat.ID)
	}
	return h.sendRoleMenu(m.Chat.ID, *user.Role)
}

func (h *StartHandler) Cancel(ctx context.Context, m *tgbotapi.Message) error {
	h.ctx.Store.Clear(m.From.ID)
	user, err := h.GetUserAndEnsureRole(ctx, m)
	if err != nil {
		return err
	}
	if user.Role == nil || *user.Role == "" {
		return h.showRoleSelection(m.Chat.ID)
	}
	msg := tgbotapi.NewMessage(m.Chat.ID, "Bekor qilindi.")
	if *user.Role == models.RoleProvider {
		msg.ReplyMarkup = templates.ProviderMenuKeyboard()
	} else {
		msg.ReplyMarkup = templates.ClientMenuKeyboard()
	}
	_, err = h.ctx.Bot.Send(msg)
	return err
}

func (h *StartHandler) RouteMenu(ctx context.Context, m *tgbotapi.Message) error {
	text := strings.TrimSpace(m.Text)
	user, err := h.GetUserAndEnsureRole(ctx, m)
	if err != nil {
		return err
	}

	if text == templates.BtnRoleProvider {
		if _, err := h.ctx.Users.SetRole(ctx, m.From.ID, models.RoleProvider); err != nil {
			return err
		}
		return h.sendRoleMenu(m.Chat.ID, models.RoleProvider)
	}
	if text == templates.BtnRoleClient {
		if _, err := h.ctx.Users.SetRole(ctx, m.From.ID, models.RoleClient); err != nil {
			return err
		}
		return h.sendRoleMenu(m.Chat.ID, models.RoleClient)
	}
	if text == templates.BtnRoleSwitch {
		h.ctx.Store.Clear(m.From.ID)
		if _, err := h.ctx.Users.ClearRole(ctx, m.From.ID); err != nil {
			return err
		}
		return h.showRoleSelection(m.Chat.ID)
	}

	if user.Role == nil || *user.Role == "" {
		return h.showRoleSelection(m.Chat.ID)
	}

	switch text {
	case templates.BtnCreateAd:
		if *user.Role != models.RoleProvider {
			return h.sendRoleMenu(m.Chat.ID, *user.Role)
		}
		msg := tgbotapi.NewMessage(m.Chat.ID, "Qaysi turdagi e'lon beramiz?")
		msg.ReplyMarkup = templates.ProviderCreateKeyboard()
		_, err := h.ctx.Bot.Send(msg)
		return err
	case templates.BtnTaxiCreate:
		if *user.Role != models.RoleProvider {
			return h.sendRoleMenu(m.Chat.ID, *user.Role)
		}
		st := session.State{Flow: session.FlowTaxiCreate, Step: session.StepTaxiFromCity}
		st.Taxi.TotalSeats = 4
		st.Taxi.OccupiedSeats = 0
		h.ctx.Store.Set(m.From.ID, st)
		msg := tgbotapi.NewMessage(m.Chat.ID, "Qayerdan? (shahar)")
		msg.ReplyMarkup = templates.CityKeyboard()
		_, err := h.ctx.Bot.Send(msg)
		return err
	case templates.BtnServiceCreate:
		if *user.Role != models.RoleProvider {
			return h.sendRoleMenu(m.Chat.ID, *user.Role)
		}
		h.ctx.Store.Set(m.From.ID, session.State{Flow: session.FlowServiceCreate, Step: session.StepServiceCategory})
		msg := tgbotapi.NewMessage(m.Chat.ID, "Xizmat yo‘nalishini tanlang:")
		msg.ReplyMarkup = templates.ServiceCategoryKeyboard()
		_, err := h.ctx.Bot.Send(msg)
		return err
	case templates.BtnSearch:
		if *user.Role != models.RoleClient {
			return h.sendRoleMenu(m.Chat.ID, *user.Role)
		}
		msg := tgbotapi.NewMessage(m.Chat.ID, "Nimani qidiramiz?")
		msg.ReplyMarkup = templates.SearchMenuKeyboard()
		_, err := h.ctx.Bot.Send(msg)
		return err
	case templates.BtnMyAds:
		if *user.Role != models.RoleProvider {
			return h.sendRoleMenu(m.Chat.ID, *user.Role)
		}
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
			if *user.Role != models.RoleClient {
				return h.sendRoleMenu(m.Chat.ID, *user.Role)
			}
			st := session.State{Flow: session.FlowTaxiSearch, Step: session.StepTaxiSearchFrom}
			h.ctx.Store.Set(m.From.ID, st)
			msg := tgbotapi.NewMessage(m.Chat.ID, "Qayerdan? (shahar)")
			msg.ReplyMarkup = templates.CityKeyboard()
			_, err := h.ctx.Bot.Send(msg)
			return err
		}
		if text == templates.BtnSearchService {
			if *user.Role != models.RoleClient {
				return h.sendRoleMenu(m.Chat.ID, *user.Role)
			}
			st := session.State{Flow: session.FlowServiceSearch, Step: session.StepServiceSearchCategory}
			h.ctx.Store.Set(m.From.ID, st)
			msg := tgbotapi.NewMessage(m.Chat.ID, "Qidirish: xizmat yo‘nalishini tanlang:")
			msg.ReplyMarkup = templates.ServiceCategoryKeyboard()
			_, err := h.ctx.Bot.Send(msg)
			return err
		}
		if text == templates.BtnBack {
			return h.sendRoleMenu(m.Chat.ID, *user.Role)
		}

		// If user is in an active session, we will handle later; for now show menu.
		return h.sendRoleMenu(m.Chat.ID, *user.Role)
	}
}

func (h *StartHandler) RouteCallback(ctx context.Context, q *tgbotapi.CallbackQuery, m *tgbotapi.Message) error {
	// Placeholder for inline buttons later (my ads management, confirm, etc.)
	user, err := h.ctx.Users.GetUserAndEnsureRole(ctx, q.From)
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(m.Chat.ID, "Hali inline tugmalar ulanmagan. /start")
	if user.Role == nil || *user.Role == "" {
		msg.ReplyMarkup = templates.RoleSelectionKeyboard()
	} else if *user.Role == models.RoleProvider {
		msg.ReplyMarkup = templates.ProviderMenuKeyboard()
	} else {
		msg.ReplyMarkup = templates.ClientMenuKeyboard()
	}
	_, err = h.ctx.Bot.Send(msg)
	return err
}

func (h *StartHandler) GetUserAndEnsureRole(ctx context.Context, m *tgbotapi.Message) (models.User, error) {
	return h.ctx.Users.GetUserAndEnsureRole(ctx, m.From)
}

func (h *StartHandler) showRoleSelection(chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, templates.RoleSelectText())
	msg.ReplyMarkup = templates.RoleSelectionKeyboard()
	_, err := h.ctx.Bot.Send(msg)
	return err
}

func (h *StartHandler) sendRoleMenu(chatID int64, role models.UserRole) error {
	msg := tgbotapi.NewMessage(chatID, templates.WelcomeText())
	if role == models.RoleProvider {
		msg.ReplyMarkup = templates.ProviderMenuKeyboard()
	} else {
		msg.ReplyMarkup = templates.ClientMenuKeyboard()
	}
	_, err := h.ctx.Bot.Send(msg)
	return err
}

