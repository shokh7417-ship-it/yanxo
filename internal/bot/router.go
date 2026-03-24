package bot

import (
	"context"
	"log"
	"runtime/debug"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"yanxo/internal/config"
	"yanxo/internal/handlers"
	"yanxo/internal/location"
	"yanxo/internal/service"
	"yanxo/internal/session"
	"yanxo/internal/templates"
)

type Router struct {
	cfg   config.Config
	bot   *tgbotapi.BotAPI
	ads   *service.AdsService
	store *session.Store

	start *handlers.StartHandler
	wiz   *handlers.WizardHandler
	myAds *handlers.MyAdsHandler
}

func NewRouter(cfg config.Config, bot *tgbotapi.BotAPI, ads *service.AdsService, store *session.Store, resolver *location.Resolver) *Router {
	hctx := handlers.Context{
		Cfg:      cfg,
		Bot:      bot,
		Ads:      ads,
		Store:    store,
		Resolver: resolver,
	}
	return &Router{
		cfg:   cfg,
		bot:   bot,
		ads:   ads,
		store: store,
		start: handlers.NewStartHandler(hctx),
		wiz:   handlers.NewWizardHandler(hctx),
		myAds: handlers.NewMyAdsHandler(hctx),
	}
}

func (r *Router) HandleUpdate(ctx context.Context, upd tgbotapi.Update) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("panic: %v\n%s", rec, string(debug.Stack()))
		}
	}()

	if upd.Message != nil {
		r.handleMessage(ctx, upd)
		return
	}
	if upd.CallbackQuery != nil {
		r.handleCallback(ctx, upd)
		return
	}
}

func (r *Router) handleMessage(ctx context.Context, upd tgbotapi.Update) {
	m := upd.Message
	if m.Chat == nil || m.From == nil {
		return
	}
	if m.Chat.IsGroup() || m.Chat.IsSuperGroup() {
		return
	}

	text := strings.TrimSpace(m.Text)
	log.Printf("msg from=%d chat=%d text=%q", m.From.ID, m.Chat.ID, text)
	if strings.HasPrefix(text, "/start") {
		_ = r.start.Start(ctx, m)
		return
	}
	if text == "/cancel" {
		_ = r.start.Cancel(ctx, m)
		return
	}

	// Global menu actions should always win over active wizard steps.
	if isGlobalMenuAction(text) {
		r.store.Clear(m.From.ID)
		_ = r.start.RouteMenu(ctx, m)
		return
	}

	// If in an active wizard flow, let wizard consume the input.
	if st, ok := r.store.Get(m.From.ID); ok && st.Flow != session.FlowNone && st.Step != session.StepNone {
		log.Printf("session user=%d flow=%s step=%s", m.From.ID, st.Flow, st.Step)
		if consumed := r.wiz.HandleWizardMessage(ctx, m, st); consumed {
			return
		}
	}
	_ = r.start.RouteMenu(ctx, m)
}

func isGlobalMenuAction(text string) bool {
	switch strings.TrimSpace(text) {
	case
		templates.BtnTaxiCreate,
		templates.BtnServiceCreate,
		templates.BtnSearch,
		templates.BtnMyAds,
		templates.BtnOpenChannel,
		templates.BtnSearchTaxi,
		templates.BtnSearchService,
		templates.BtnBack:
		return true
	default:
		return false
	}
}

func (r *Router) handleCallback(ctx context.Context, upd tgbotapi.Update) {
	q := upd.CallbackQuery
	if q.Message == nil || q.From == nil {
		return
	}

	chatID := q.Message.Chat.ID
	messageID := q.Message.MessageID

	// Instant UX: remove inline keyboard as soon as callback arrives,
	// before any business logic (publish/delete/cancel) runs.
	// IMPORTANT: edit the exact callback message (q.Message.*).
	if chatID != 0 && messageID != 0 {
		// Explicit empty inline_keyboard makes Telegram remove buttons reliably.
		empty := tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
		}
		editKB := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, empty)
		if _, err := r.bot.Request(editKB); err != nil {
			log.Printf("inline kb remove failed chat_id=%d msg_id=%d callback_id=%s: %v", chatID, messageID, q.ID, err)
		}
	}

	// For now, just ack + show menu
	ack := tgbotapi.NewCallback(q.ID, "")
	_, _ = r.bot.Request(ack)

	msg := q.Message

	if r.wiz.HandleWizardCallback(ctx, q, msg) {
		return
	}
	if r.myAds.HandleCallback(ctx, q, msg) {
		return
	}
	_ = r.start.RouteCallback(ctx, q, msg)
}

