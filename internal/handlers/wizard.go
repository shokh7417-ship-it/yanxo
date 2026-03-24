package handlers

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"yanxo/internal/models"
	"yanxo/internal/session"
	"yanxo/internal/templates"
)

type WizardHandler struct {
	ctx Context
}

func NewWizardHandler(ctx Context) *WizardHandler { return &WizardHandler{ctx: ctx} }

func (h *WizardHandler) HandleWizardMessage(ctx context.Context, m *tgbotapi.Message, st session.State) bool {
	switch st.Flow {
	case session.FlowTaxiCreate:
		return h.handleTaxiCreate(ctx, m, st)
	case session.FlowServiceCreate:
		return h.handleServiceCreate(ctx, m, st)
	case session.FlowTaxiSearch:
		return h.handleTaxiSearch(ctx, m, st)
	case session.FlowServiceSearch:
		return h.handleServiceSearch(ctx, m, st)
	default:
		return false
	}
}

func (h *WizardHandler) HandleWizardCallback(ctx context.Context, q *tgbotapi.CallbackQuery, msg *tgbotapi.Message) bool {
	data := strings.TrimSpace(q.Data)
	if data == "" {
		return false
	}

	// Search result actions (stateless)
	if strings.HasPrefix(data, "sr:") {
		return h.handleSearchResultCallback(ctx, q, msg, data)
	}

	// Confirm / cancel
	if strings.HasPrefix(data, "confirm:") || strings.HasPrefix(data, "cancel:") ||
		strings.HasPrefix(data, "contact:") || data == "noop" {
		st, ok := h.ctx.Store.Get(q.From.ID)
		if !ok {
			return false
		}
		switch st.Flow {
		case session.FlowTaxiCreate:
			return h.handleTaxiCallback(ctx, q, msg, st, data)
		case session.FlowServiceCreate:
			return h.handleServiceCallback(ctx, q, msg, st, data)
		default:
			return false
		}
	}
	return false
}

func (h *WizardHandler) handleSearchResultCallback(ctx context.Context, q *tgbotapi.CallbackQuery, msg *tgbotapi.Message, data string) bool {
	parts := strings.SplitN(data, ":", 3)
	if len(parts) != 3 {
		return false
	}
	action := parts[1]
	adID := parts[2]

	ctx2, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	ad, err := h.ctx.Ads.Get(ctx2, adID)
	if err != nil {
		log.Printf("sr callback get ad error: %v", err)
		_ = h.sendText(msg.Chat.ID, "Xatolik. Keyinroq urinib ko‘ring.")
		return true
	}

	switch action {
	case "contact":
		if ad.Contact == nil || *ad.Contact == "" {
			_ = h.sendText(msg.Chat.ID, "Aloqa ma’lumoti yo‘q.")
			return true
		}
		_ = h.sendText(msg.Chat.ID, "📞 Aloqa: "+*ad.Contact)
		return true
	case "post":
		link := templates.ChannelPostLink(h.ctx.Cfg.ChannelID, h.ctx.Cfg.ChannelUsername, ad.ChannelMessageID)
		if link == "" {
			_ = h.sendText(msg.Chat.ID, "Post link topilmadi.")
			return true
		}
		out := tgbotapi.NewMessage(msg.Chat.ID, "📄 Postni ochish:")
		out.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("📄 Postni ochish", link),
			),
		)
		if _, err := h.ctx.Bot.Send(out); err != nil {
			log.Printf("sr post send error: %v", err)
			_ = h.sendText(msg.Chat.ID, "Xatolik: link yuborilmadi.")
		}
		return true
	default:
		return false
	}
}

// ---------------- Taxi create ----------------

func (h *WizardHandler) handleTaxiCreate(ctx context.Context, m *tgbotapi.Message, st session.State) bool {
	text := strings.TrimSpace(m.Text)
	// Contact cards have no m.Text; only allow empty text on the phone step when sharing contact.
	if text == "" && !(st.Step == session.StepTaxiContact && m.Contact != nil) {
		_ = h.sendText(m.Chat.ID, "Iltimos, qiymat kiriting.")
		return true
	}

	switch st.Step {
	case session.StepTaxiFromCity:
		if strings.EqualFold(text, "boshqa") {
			_ = h.sendText(m.Chat.ID, "Shahar nomini yozing:")
			return true
		}
		canonical, err := h.ctx.Resolver.Resolve(ctx, text)
		if err != nil {
			_ = h.sendText(m.Chat.ID, "Xatolik. Keyinroq urinib ko‘ring.")
			return true
		}
		if canonical == "" {
			_ = h.sendText(m.Chat.ID, "Shahar topilmadi. Boshqa yozib ko‘ring yoki to‘g‘ri yozing (masalan: Toshkent, Xiva).")
			return true
		}
		st.Taxi.FromCity = canonical
		st.Step = session.StepTaxiToCity
		h.ctx.Store.Set(m.From.ID, st)
		_ = h.sendMarkup(m.Chat.ID, "Qayerga? (shahar)", templates.CityKeyboard())
		return true
	case session.StepTaxiToCity:
		if strings.EqualFold(text, "boshqa") {
			_ = h.sendText(m.Chat.ID, "Shahar nomini yozing:")
			return true
		}
		canonical, err := h.ctx.Resolver.Resolve(ctx, text)
		if err != nil {
			_ = h.sendText(m.Chat.ID, "Xatolik. Keyinroq urinib ko‘ring.")
			return true
		}
		if canonical == "" {
			_ = h.sendText(m.Chat.ID, "Shahar topilmadi. Boshqa yozib ko‘ring yoki to‘g‘ri yozing (masalan: Toshkent, Xiva).")
			return true
		}
		st.Taxi.ToCity = canonical
		// default date: today
		st.Taxi.RideDate = time.Now().In(time.Local).Format("2006-01-02")
		st.Step = session.StepTaxiRideDate
		h.ctx.Store.Set(m.From.ID, st)
		_ = h.sendText(m.Chat.ID, fmt.Sprintf("Sana (YYYY-MM-DD). Default: %s\nAgar default bo‘lsa, “✅ Bugun” tugmasini bosing.", st.Taxi.RideDate))
		_ = h.sendMarkup(m.Chat.ID, "Sana tanlang:", templates.TaxiDateKeyboard())
		return true
	case session.StepTaxiRideDate:
		if text == "✅ Bugun" {
			// keep default
		} else {
			if !isDateYYYYMMDD(text) {
				_ = h.sendMarkup(m.Chat.ID, "Sana noto‘g‘ri. Format: YYYY-MM-DD\nMasalan: 2026-03-19", templates.TaxiDateKeyboard())
				return true
			}
			st.Taxi.RideDate = text
		}
		st.Step = session.StepTaxiDepartureTime
		h.ctx.Store.Set(m.From.ID, st)
		_ = h.sendText(m.Chat.ID, "Jo‘nash vaqti (HH:MM). Masalan: 18:30")
		return true
	case session.StepTaxiDepartureTime:
		if !isTimeHHMM(text) {
			_ = h.sendText(m.Chat.ID, "Vaqt noto‘g‘ri. Format: HH:MM (24 soat). Masalan: 06:45 yoki 18:30")
			return true
		}
		st.Taxi.DepartureTime = text
		st.Taxi.TotalSeats = 4
		st.Taxi.OccupiedSeats = 0
		st.Step = session.StepTaxiCarType
		h.ctx.Store.Set(m.From.ID, st)
		_ = h.sendMarkup(m.Chat.ID, "Mashina turi:", templates.CarTypeKeyboard())
		return true
	case session.StepTaxiCarType:
		car := text
		if car == "" {
			_ = h.sendText(m.Chat.ID, "Mashina turini yozing yoki tugmadan tanlang (masalan: Cobalt, Gentra).")
			return true
		}
		st.Taxi.CarType = car
		st.Step = session.StepTaxiTotalSeats
		h.ctx.Store.Set(m.From.ID, st)
		_ = h.sendMarkup(m.Chat.ID, "Jami o‘rinlar soni? (default 4)\nAgar 4 bo‘lsa, “✅ 4” tugmasini bosing.", templates.TotalSeatsKeyboard())
		return true
	case session.StepTaxiTotalSeats:
		if text == "✅ 4" {
			st.Taxi.TotalSeats = 4
			st.Step = session.StepTaxiContact
			h.ctx.Store.Set(m.From.ID, st)
			return h.askContactTaxi(ctx, m, st)
		}

		n, err := strconv.Atoi(text)
		if err != nil || n < 1 || n > 8 {
			_ = h.sendText(m.Chat.ID, "O‘rinlar soni noto‘g‘ri. 1-8 orasida kiriting. Default: 4")
			return true
		}
		st.Taxi.TotalSeats = n
		st.Step = session.StepTaxiContact
		h.ctx.Store.Set(m.From.ID, st)
		return h.askContactTaxi(ctx, m, st)
	case session.StepTaxiContact:
		if m.Contact != nil && m.Contact.PhoneNumber != "" {
			ph := normalizePhone(m.Contact.PhoneNumber)
			st.Taxi.Contact = &ph
			st.Step = session.StepTaxiPreview
			h.ctx.Store.Set(m.From.ID, st)
			return h.showTaxiPreview(ctx, m.Chat.ID, m.From, st)
		}
		// manual phone input
		if !isPhoneLike(text) {
			_ = h.sendText(m.Chat.ID, "Telefon noto‘g‘ri ko‘rinadi. Masalan: +998901234567 yoki 90 123 45 67")
			return true
		}
		contact := normalizePhone(text)
		st.Taxi.Contact = &contact
		st.Step = session.StepTaxiPreview
		h.ctx.Store.Set(m.From.ID, st)
		return h.showTaxiPreview(ctx, m.Chat.ID, m.From, st)
	default:
		return false
	}
}

func (h *WizardHandler) handleTaxiCallback(ctx context.Context, q *tgbotapi.CallbackQuery, msg *tgbotapi.Message, st session.State, data string) bool {
	if strings.HasPrefix(data, "car:") {
		car := strings.TrimPrefix(data, "car:")
		st.Taxi.CarType = car
		st.Step = session.StepTaxiTotalSeats
		h.ctx.Store.Set(q.From.ID, st)
		_ = h.sendText(msg.Chat.ID, "Jami o‘rinlar soni? (default 4)\nAgar 4 bo‘lsa, “✅ 4” tugmasini bosing.")
		_ = h.sendMarkup(msg.Chat.ID, "Tanlang:", templates.TotalSeatsKeyboard())
		return true
	}

	if strings.HasPrefix(data, "contact:") {
		action := strings.TrimPrefix(data, "contact:")
		switch action {
		case "use_username":
			if q.From.UserName != "" {
				u := "@" + q.From.UserName
				st.Taxi.Contact = &u
			}
			st.Step = session.StepTaxiPreview
			h.ctx.Store.Set(q.From.ID, st)
			return h.showTaxiPreview(ctx, msg.Chat.ID, q.From, st)
		case "enter_phone":
			st.Step = session.StepTaxiContact
			h.ctx.Store.Set(q.From.ID, st)
			// ask for phone share or input
			_ = h.sendMarkup(msg.Chat.ID, "Telefon yuboring yoki raqamni yozing:", templates.PhoneRequestKeyboard())
			return true
		}
	}

	if strings.HasPrefix(data, "confirm:taxi") {
		// Prevent duplicate actions: clear state before any DB/network work.
		h.ctx.Store.Clear(q.From.ID)

		// Create ad, post to channel, store channel_message_id
		ad, err := h.ctx.Ads.CreateTaxi(ctx, q.From.ID,
			st.Taxi.FromCity, st.Taxi.ToCity, st.Taxi.RideDate, st.Taxi.DepartureTime, st.Taxi.CarType,
			st.Taxi.TotalSeats, st.Taxi.OccupiedSeats, st.Taxi.Contact)
		if err != nil {
			_ = h.sendText(msg.Chat.ID, "Xatolik: e’lon saqlanmadi. Keyinroq urinib ko‘ring.")
			return true
		}
		chMsgID, postErr := h.postToChannelTaxi(ad)
		if postErr == nil {
			ad.ChannelMessageID = &chMsgID
			_ = h.ctx.Ads.UpdateChannelMessageID(ctx, ad.ID, q.From.ID, chMsgID)
		}

		// Remove old inline confirm/cancel buttons from preview message.
		empty := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{}}
		editKB := tgbotapi.NewEditMessageReplyMarkup(msg.Chat.ID, msg.MessageID, empty)
		_, _ = h.ctx.Bot.Send(editKB)

		// Success: attach inline "Postni ochish" button.
		success := tgbotapi.NewMessage(msg.Chat.ID, "✅ E’lon joylandi!")
		success.ReplyMarkup = templates.PostOpenInline(ad, h.ctx.Cfg.ChannelID, h.ctx.Cfg.ChannelUsername)
		_, _ = h.ctx.Bot.Send(success)

		_ = h.sendMainMenu(msg.Chat.ID)
		return true
	}
	if strings.HasPrefix(data, "cancel:taxi") {
		// Prevent duplicate actions: clear state before any further logic.
		h.ctx.Store.Clear(q.From.ID)

		// Remove preview buttons after cancel (in case router-level edit fails).
		empty := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{}}
		editKB := tgbotapi.NewEditMessageReplyMarkup(msg.Chat.ID, msg.MessageID, empty)
		_, _ = h.ctx.Bot.Send(editKB)

		out := tgbotapi.NewMessage(msg.Chat.ID, "Bekor qilindi. Asosiy menyu.")
		out.ReplyMarkup = templates.MainMenuKeyboard()
		_, _ = h.ctx.Bot.Send(out)
		return true
	}
	return false
}

func (h *WizardHandler) askContactTaxi(ctx context.Context, m *tgbotapi.Message, st session.State) bool {
	_ = ctx
	if m.From.UserName != "" {
		txt := fmt.Sprintf("Aloqa uchun Telegram username topildi: @%s", m.From.UserName)
		_ = h.sendMarkup(m.Chat.ID, txt, templates.ContactChoiceWithUsername())
		return true
	}
	_ = h.sendMarkup(m.Chat.ID, "Aloqa (ixtiyoriy):", templates.ContactChoiceNoUsername())
	return true
}

func (h *WizardHandler) showTaxiPreview(ctx context.Context, chatID int64, u *tgbotapi.User, st session.State) bool {
	_ = ctx
	text := templates.FormatTaxiPreview(
		st.Taxi.FromCity, st.Taxi.ToCity,
		st.Taxi.RideDate, st.Taxi.DepartureTime,
		st.Taxi.CarType, st.Taxi.OccupiedSeats, st.Taxi.TotalSeats,
		st.Taxi.Contact,
	)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = templates.ConfirmKeyboard("taxi")
	_, _ = h.ctx.Bot.Send(msg)
	return true
}

func (h *WizardHandler) postToChannelTaxi(ad models.Ad) (int, error) {
	text := templates.TaxiChannelPost(ad)
	ch := tgbotapi.NewMessage(h.ctx.Cfg.ChannelID, text)
	ch.ParseMode = "HTML"
	sent, err := h.ctx.Bot.Send(ch)
	if err != nil {
		return 0, err
	}
	return sent.MessageID, nil
}

// ---------------- Service create ----------------

func (h *WizardHandler) handleServiceCreate(ctx context.Context, m *tgbotapi.Message, st session.State) bool {
	text := strings.TrimSpace(m.Text)
	if text == "" && !(st.Step == session.StepServiceContact && m.Contact != nil) {
		_ = h.sendText(m.Chat.ID, "Iltimos, qiymat kiriting.")
		return true
	}

	switch st.Step {
	case session.StepServiceCategory:
		if text == templates.ServiceWizardCancel {
			h.ctx.Store.Clear(m.From.ID)
			out := tgbotapi.NewMessage(m.Chat.ID, "Bekor qilindi. Asosiy menyu.")
			out.ReplyMarkup = templates.MainMenuKeyboard()
			_, _ = h.ctx.Bot.Send(out)
			return true
		}
		if text == templates.ServiceTypeOtherBtn {
			st.Step = session.StepServiceCustomType
			h.ctx.Store.Set(m.From.ID, st)
			_ = h.sendMarkup(m.Chat.ID, "Xizmat turini yozing (masalan: santexnik, elektrik):", templates.ServiceCustomTypeKeyboard())
			return true
		}
		cat, ok := templates.ServiceCategoryFromButton(text)
		if !ok {
			_ = h.sendMarkup(m.Chat.ID, "Iltimos, pastdagi tugmalardan tanlang.", templates.ServiceCategoryKeyboard())
			return true
		}
		st.Service.PickCategory = cat
		st.Step = session.StepServicePick
		h.ctx.Store.Set(m.From.ID, st)
		_ = h.sendMarkup(m.Chat.ID, "Aniq xizmat turini tanlang:", templates.ServicePickKeyboard(cat))
		return true
	case session.StepServicePick:
		if text == templates.ServicePickBackBtn {
			st.Service.PickCategory = ""
			st.Step = session.StepServiceCategory
			h.ctx.Store.Set(m.From.ID, st)
			_ = h.sendMarkup(m.Chat.ID, "Xizmat yo‘nalishini tanlang:", templates.ServiceCategoryKeyboard())
			return true
		}
		if !templates.IsKnownServicePick(st.Service.PickCategory, text) {
			_ = h.sendMarkup(m.Chat.ID, "Iltimos, ro‘yxatdan tanlang yoki «⬅️ Kategoriyalar».", templates.ServicePickKeyboard(st.Service.PickCategory))
			return true
		}
		st.Service.ServiceType = text
		st.Service.PickCategory = ""
		st.Step = session.StepServiceArea
		h.ctx.Store.Set(m.From.ID, st)
		_ = h.sendMarkup(m.Chat.ID, "Hudud qayer? Shahar/tumanni tanlang yoki «Boshqa» ni bosing.", templates.ServicePlaceKeyboard())
		return true
	case session.StepServiceCustomType:
		if text == templates.ServicePickBackBtn {
			st.Step = session.StepServiceCategory
			h.ctx.Store.Set(m.From.ID, st)
			_ = h.sendMarkup(m.Chat.ID, "Xizmat yo‘nalishini tanlang:", templates.ServiceCategoryKeyboard())
			return true
		}
		if text == templates.ServiceWizardCancel {
			h.ctx.Store.Clear(m.From.ID)
			out := tgbotapi.NewMessage(m.Chat.ID, "Bekor qilindi. Asosiy menyu.")
			out.ReplyMarkup = templates.MainMenuKeyboard()
			_, _ = h.ctx.Bot.Send(out)
			return true
		}
		st.Service.ServiceType = text
		st.Step = session.StepServiceArea
		h.ctx.Store.Set(m.From.ID, st)
		_ = h.sendMarkup(m.Chat.ID, "Hudud qayer? Shahar/tumanni tanlang yoki «Boshqa» ni bosing.", templates.ServicePlaceKeyboard())
		return true
	case session.StepServiceArea:
		if strings.EqualFold(text, "boshqa") {
			_ = h.sendText(m.Chat.ID, "Boshqa hudud nomini yozing (masalan: Chilonzor, Samarqand):")
			return true
		}
		st.Service.Area = text
		st.Step = session.StepServiceNote
		h.ctx.Store.Set(m.From.ID, st)
		_ = h.sendMarkup(m.Chat.ID, "Qo‘shimcha izoh (ixtiyoriy). “⏭ O‘tkazib yuborish” ni bosing yoki yozing:", templates.SkipKeyboard())
		return true
	case session.StepServiceNote:
		if text == "⏭ O‘tkazib yuborish" {
			st.Service.Note = nil
		} else {
			st.Service.Note = &text
		}
		st.Step = session.StepServiceContact
		h.ctx.Store.Set(m.From.ID, st)
		return h.askContactService(ctx, m, st)
	case session.StepServiceContact:
		if m.Contact != nil && m.Contact.PhoneNumber != "" {
			ph := normalizePhone(m.Contact.PhoneNumber)
			st.Service.Contact = &ph
			st.Step = session.StepServicePreview
			h.ctx.Store.Set(m.From.ID, st)
			return h.showServicePreview(ctx, m.Chat.ID, m.From, st)
		}
		if !isPhoneLike(text) {
			_ = h.sendText(m.Chat.ID, "Telefon noto‘g‘ri ko‘rinadi. Masalan: +998901234567 yoki 90 123 45 67")
			return true
		}
		contact := normalizePhone(text)
		st.Service.Contact = &contact
		st.Step = session.StepServicePreview
		h.ctx.Store.Set(m.From.ID, st)
		return h.showServicePreview(ctx, m.Chat.ID, m.From, st)
	default:
		return false
	}
}

func (h *WizardHandler) handleServiceCallback(ctx context.Context, q *tgbotapi.CallbackQuery, msg *tgbotapi.Message, st session.State, data string) bool {
	if data == "noop" {
		return true
	}
	if strings.HasPrefix(data, "contact:") {
		action := strings.TrimPrefix(data, "contact:")
		switch action {
		case "use_username":
			if q.From.UserName != "" {
				u := "@" + q.From.UserName
				st.Service.Contact = &u
			}
			st.Step = session.StepServicePreview
			h.ctx.Store.Set(q.From.ID, st)
			return h.showServicePreview(ctx, msg.Chat.ID, q.From, st)
		case "enter_phone":
			st.Step = session.StepServiceContact
			h.ctx.Store.Set(q.From.ID, st)
			_ = h.sendMarkup(msg.Chat.ID, "Telefon yuboring yoki raqamni yozing:", templates.PhoneRequestKeyboard())
			return true
		}
	}

	if strings.HasPrefix(data, "confirm:service") {
		// Prevent duplicate actions: clear state before any DB/network work.
		h.ctx.Store.Clear(q.From.ID)

		ad, err := h.ctx.Ads.CreateService(ctx, q.From.ID, st.Service.ServiceType, st.Service.Area, st.Service.Note, st.Service.Contact)
		if err != nil {
			_ = h.sendText(msg.Chat.ID, "Xatolik: e’lon saqlanmadi. Keyinroq urinib ko‘ring.")
			return true
		}
		chMsgID, postErr := h.postToChannelService(ad)
		if postErr == nil {
			ad.ChannelMessageID = &chMsgID
			_ = h.ctx.Ads.UpdateChannelMessageID(ctx, ad.ID, q.From.ID, chMsgID)
		}

		// Remove old inline confirm/cancel buttons from preview message.
		empty := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{}}
		editKB := tgbotapi.NewEditMessageReplyMarkup(msg.Chat.ID, msg.MessageID, empty)
		_, _ = h.ctx.Bot.Send(editKB)

		success := tgbotapi.NewMessage(msg.Chat.ID, "✅ E’lon joylandi!")
		success.ReplyMarkup = templates.PostOpenInline(ad, h.ctx.Cfg.ChannelID, h.ctx.Cfg.ChannelUsername)
		_, _ = h.ctx.Bot.Send(success)

		_ = h.sendMainMenu(msg.Chat.ID)
		return true
	}
	if strings.HasPrefix(data, "cancel:service") {
		// Prevent duplicate actions: clear state before any further logic.
		h.ctx.Store.Clear(q.From.ID)

		// Remove preview buttons after cancel (in case router-level edit fails).
		empty := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{}}
		editKB := tgbotapi.NewEditMessageReplyMarkup(msg.Chat.ID, msg.MessageID, empty)
		_, _ = h.ctx.Bot.Send(editKB)

		out := tgbotapi.NewMessage(msg.Chat.ID, "Bekor qilindi. Asosiy menyu.")
		out.ReplyMarkup = templates.MainMenuKeyboard()
		_, _ = h.ctx.Bot.Send(out)
		return true
	}
	return false
}

// ---------------- Search flows ----------------

func (h *WizardHandler) handleTaxiSearch(ctx context.Context, m *tgbotapi.Message, st session.State) bool {
	text := strings.TrimSpace(m.Text)
	if text == "" {
		_ = h.sendText(m.Chat.ID, "Iltimos, qiymat kiriting.")
		return true
	}
	switch st.Step {
	case session.StepTaxiSearchFrom:
		if strings.EqualFold(text, "boshqa") {
			_ = h.sendText(m.Chat.ID, "Shahar nomini yozing:")
			return true
		}
		log.Printf("taxi_search_from input=%q", text)
		ctx2, cancel := context.WithTimeout(ctx, 4*time.Second)
		defer cancel()
		canonical, err := h.ctx.Resolver.Resolve(ctx2, text)
		if err != nil {
			log.Printf("taxi_search_from resolve error: %v", err)
			_ = h.sendText(m.Chat.ID, "Xatolik: shahar aniqlanmadi. Keyinroq urinib ko‘ring.")
			return true
		}
		log.Printf("taxi_search_from canonical=%q", canonical)
		if canonical == "" {
			_ = h.sendText(m.Chat.ID, "Shahar topilmadi. Boshqa yozib ko‘ring (masalan: Toshkent, Xiva).")
			return true
		}
		st.Search.TaxiFrom = canonical
		st.Step = session.StepTaxiSearchTo
		h.ctx.Store.Set(m.From.ID, st)
		_ = h.sendMarkup(m.Chat.ID, "Qayerga? (shahar)", templates.CityKeyboard())
		return true
	case session.StepTaxiSearchTo:
		// Recovery keyboard actions for taxi no-results.
		if strings.EqualFold(text, templates.BtnBack) {
			st.Step = session.StepTaxiSearchFrom
			h.ctx.Store.Set(m.From.ID, st)
			_ = h.sendMarkup(m.Chat.ID, "Qayerdan? (shahar)", templates.CityKeyboard())
			return true
		}
		if text == "❌ Bekor qilish" {
			h.ctx.Store.Clear(m.From.ID)
			out := tgbotapi.NewMessage(m.Chat.ID, "Bekor qilindi. Asosiy menyu.")
			out.ReplyMarkup = templates.MainMenuKeyboard()
			_, _ = h.ctx.Bot.Send(out)
			return true
		}

		if strings.EqualFold(text, "boshqa") {
			_ = h.sendText(m.Chat.ID, "Shahar nomini yozing:")
			return true
		}
		log.Printf("taxi_search_to input=%q", text)
		ctx2, cancel := context.WithTimeout(ctx, 4*time.Second)
		defer cancel()
		canonical, err := h.ctx.Resolver.Resolve(ctx2, text)
		if err != nil {
			log.Printf("taxi_search_to resolve error: %v", err)
			_ = h.sendText(m.Chat.ID, "Xatolik: shahar aniqlanmadi. Keyinroq urinib ko‘ring.")
			return true
		}
		log.Printf("taxi_search_to canonical=%q", canonical)
		if canonical == "" {
			_ = h.sendText(m.Chat.ID, "Shahar topilmadi. Boshqa yozib ko‘ring (masalan: Toshkent, Xiva).")
			return true
		}
		st.Search.TaxiTo = canonical
		h.ctx.Store.Set(m.From.ID, st)

		log.Printf("taxi_search query from=%q to=%q", st.Search.TaxiFrom, st.Search.TaxiTo)
		ctx3, cancel2 := context.WithTimeout(ctx, 6*time.Second)
		defer cancel2()
		ads, err := h.ctx.Ads.SearchTaxi(ctx3, st.Search.TaxiFrom, st.Search.TaxiTo, 10)
		if err != nil {
			log.Printf("taxi_search db error: %v", err)
			_ = h.sendText(m.Chat.ID, "Xatolik: qidiruv bajarilmadi. Keyinroq urinib ko‘ring.")
			return true
		}
		log.Printf("taxi_search results=%d", len(ads))
		if len(ads) == 0 {
			// Keep the state for quick retry.
			msg := tgbotapi.NewMessage(m.Chat.ID,
				"❌ Mos taksi topilmadi.\n\nManzilni qayta kiriting.")
			_, _ = h.ctx.Bot.Send(msg)

			ask := tgbotapi.NewMessage(m.Chat.ID, "Qayerga? (shahar)")
			ask.ReplyMarkup = templates.ServiceSearchNoResultsKeyboard()
			_, _ = h.ctx.Bot.Send(ask)

			return true
		}
		// Show results count with main menu keyboard immediately.
		out := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("Topildi: %d ta e’lon", len(ads)))
		out.ReplyMarkup = templates.MainMenuKeyboard()
		if _, err := h.ctx.Bot.Send(out); err != nil {
			log.Printf("taxi_search send topildi error: %v", err)
		}
		for _, ad := range ads {
			txt := templates.TaxiSearchResultCard(ad)
			out := tgbotapi.NewMessage(m.Chat.ID, txt)
			out.ReplyMarkup = templates.SearchResultInline(ad, h.ctx.Cfg.ChannelID, h.ctx.Cfg.ChannelUsername)
			if _, err := h.ctx.Bot.Send(out); err != nil {
				log.Printf("taxi_search send result error: %v", err)
				_ = h.sendText(m.Chat.ID, "Xatolik: natija yuborilmadi. Keyinroq urinib ko‘ring.")
				return true
			}
		}
		// Search flow finished; clear state so user is idle.
		h.ctx.Store.Clear(m.From.ID)
		return true
	default:
		return false
	}
}

func (h *WizardHandler) handleServiceSearch(ctx context.Context, m *tgbotapi.Message, st session.State) bool {
	text := strings.TrimSpace(m.Text)
	if text == "" {
		_ = h.sendText(m.Chat.ID, "Iltimos, qiymat kiriting.")
		return true
	}
	switch st.Step {
	case session.StepServiceSearchCategory:
		if text == templates.ServiceWizardCancel {
			h.ctx.Store.Clear(m.From.ID)
			out := tgbotapi.NewMessage(m.Chat.ID, "Bekor qilindi. Asosiy menyu.")
			out.ReplyMarkup = templates.MainMenuKeyboard()
			_, _ = h.ctx.Bot.Send(out)
			return true
		}
		if text == templates.ServiceTypeOtherBtn {
			st.Step = session.StepServiceSearchCustomType
			h.ctx.Store.Set(m.From.ID, st)
			_ = h.sendMarkup(m.Chat.ID, "Qidirilayotgan xizmat turini yozing:", templates.ServiceCustomTypeKeyboard())
			return true
		}
		cat, ok := templates.ServiceCategoryFromButton(text)
		if !ok {
			_ = h.sendMarkup(m.Chat.ID, "Iltimos, pastdagi tugmalardan tanlang.", templates.ServiceCategoryKeyboard())
			return true
		}
		st.Search.ServicePickCategory = cat
		st.Step = session.StepServiceSearchPick
		h.ctx.Store.Set(m.From.ID, st)
		_ = h.sendMarkup(m.Chat.ID, "Aniq xizmat turini tanlang:", templates.ServicePickKeyboard(cat))
		return true
	case session.StepServiceSearchPick:
		if text == templates.ServicePickBackBtn {
			st.Search.ServicePickCategory = ""
			st.Step = session.StepServiceSearchCategory
			h.ctx.Store.Set(m.From.ID, st)
			_ = h.sendMarkup(m.Chat.ID, "Qidirish: xizmat yo‘nalishini tanlang:", templates.ServiceCategoryKeyboard())
			return true
		}
		if !templates.IsKnownServicePick(st.Search.ServicePickCategory, text) {
			_ = h.sendMarkup(m.Chat.ID, "Iltimos, ro‘yxatdan tanlang yoki «⬅️ Kategoriyalar».", templates.ServicePickKeyboard(st.Search.ServicePickCategory))
			return true
		}
		st.Search.ServiceType = text
		st.Search.ServicePickCategory = ""
		st.Step = session.StepServiceSearchArea
		h.ctx.Store.Set(m.From.ID, st)
		_ = h.sendMarkup(m.Chat.ID, "Hududni tanlang yoki «Boshqa» ni bosing.", templates.ServiceSearchAreaKeyboard())
		return true
	case session.StepServiceSearchCustomType:
		if text == templates.ServicePickBackBtn {
			st.Step = session.StepServiceSearchCategory
			h.ctx.Store.Set(m.From.ID, st)
			_ = h.sendMarkup(m.Chat.ID, "Qidirish: xizmat yo‘nalishini tanlang:", templates.ServiceCategoryKeyboard())
			return true
		}
		if text == templates.ServiceWizardCancel {
			h.ctx.Store.Clear(m.From.ID)
			out := tgbotapi.NewMessage(m.Chat.ID, "Bekor qilindi. Asosiy menyu.")
			out.ReplyMarkup = templates.MainMenuKeyboard()
			_, _ = h.ctx.Bot.Send(out)
			return true
		}
		st.Search.ServiceType = text
		st.Step = session.StepServiceSearchArea
		h.ctx.Store.Set(m.From.ID, st)
		_ = h.sendMarkup(m.Chat.ID, "Hududni tanlang yoki «Boshqa» ni bosing.", templates.ServiceSearchAreaKeyboard())
		return true
	case session.StepServiceSearchArea:
		// Recovery keyboard actions for service no-results UX.
		if strings.EqualFold(text, templates.BtnBack) {
			st.Search.ServiceArea = ""
			st.Search.ServiceType = ""
			st.Search.ServicePickCategory = ""
			st.Step = session.StepServiceSearchCategory
			h.ctx.Store.Set(m.From.ID, st)
			_ = h.sendMarkup(m.Chat.ID, "Qidirish: xizmat yo‘nalishini tanlang:", templates.ServiceCategoryKeyboard())
			return true
		}
		if text == "❌ Bekor qilish" {
			h.ctx.Store.Clear(m.From.ID)
			out := tgbotapi.NewMessage(m.Chat.ID, "Bekor qilindi. Asosiy menyu.")
			out.ReplyMarkup = templates.MainMenuKeyboard()
			_, _ = h.ctx.Bot.Send(out)
			return true
		}
		if strings.EqualFold(text, "boshqa") {
			_ = h.sendText(m.Chat.ID, "Boshqa hudud nomini yozing:")
			return true
		}

		st.Search.ServiceArea = text

		ctx2, cancel := context.WithTimeout(ctx, 6*time.Second)
		defer cancel()
		ads, err := h.ctx.Ads.SearchService(ctx2, st.Search.ServiceType, st.Search.ServiceArea, 10)
		if err != nil {
			log.Printf("service_search db error: %v", err)
			_ = h.sendText(m.Chat.ID, "Xatolik: qidiruv bajarilmadi. Keyinroq urinib ko‘ring.")
			return true
		}
		if len(ads) == 0 {
			// Keep state to allow recovery retry (area step).
			h.ctx.Store.Set(m.From.ID, st)

			// 1) Message
			msg := tgbotapi.NewMessage(m.Chat.ID, "❌ Mos usta topilmadi.\n\nHududni qayta kiriting.")
			_, _ = h.ctx.Bot.Send(msg)

			// 2) Ask again immediately with keyboard
			ask := tgbotapi.NewMessage(m.Chat.ID, "Hududni qayta tanlang yoki «Boshqa» ni bosing.")
			ask.ReplyMarkup = templates.ServiceSearchAreaKeyboard()
			_, _ = h.ctx.Bot.Send(ask)
			return true
		}
		// Show results count with main menu keyboard immediately.
		out := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("Topildi: %d ta e’lon", len(ads)))
		out.ReplyMarkup = templates.MainMenuKeyboard()
		if _, err := h.ctx.Bot.Send(out); err != nil {
			log.Printf("service_search send topildi error: %v", err)
		}
		for _, ad := range ads {
			txt := templates.ServiceSearchResultCard(ad)
			out := tgbotapi.NewMessage(m.Chat.ID, txt)
			out.ReplyMarkup = templates.SearchResultInline(ad, h.ctx.Cfg.ChannelID, h.ctx.Cfg.ChannelUsername)
			_, _ = h.ctx.Bot.Send(out)
		}
		// Search flow finished; clear state.
		h.ctx.Store.Clear(m.From.ID)
		return true
	default:
		return false
	}
}

func (h *WizardHandler) askContactService(ctx context.Context, m *tgbotapi.Message, st session.State) bool {
	_ = ctx
	_ = st
	if m.From.UserName != "" {
		txt := fmt.Sprintf("Aloqa uchun Telegram username topildi: @%s", m.From.UserName)
		_ = h.sendMarkup(m.Chat.ID, txt, templates.ContactChoiceWithUsername())
		return true
	}
	_ = h.sendMarkup(m.Chat.ID, "Aloqa (ixtiyoriy):", templates.ContactChoiceNoUsername())
	return true
}

func (h *WizardHandler) showServicePreview(ctx context.Context, chatID int64, u *tgbotapi.User, st session.State) bool {
	_ = ctx
	_ = u
	text := templates.FormatServicePreview(st.Service.ServiceType, st.Service.Area, st.Service.Note, st.Service.Contact)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = templates.ConfirmKeyboard("service")
	_, _ = h.ctx.Bot.Send(msg)
	return true
}

func (h *WizardHandler) postToChannelService(ad models.Ad) (int, error) {
	text := templates.ServiceChannelPost(ad)
	ch := tgbotapi.NewMessage(h.ctx.Cfg.ChannelID, text)
	ch.ParseMode = "HTML"
	sent, err := h.ctx.Bot.Send(ch)
	if err != nil {
		return 0, err
	}
	return sent.MessageID, nil
}

// ---------------- helpers ----------------

func (h *WizardHandler) sendText(chatID int64, text string) error {
	_, err := h.ctx.Bot.Send(tgbotapi.NewMessage(chatID, text))
	if err != nil {
		log.Printf("sendText error chat_id=%d: %v", chatID, err)
	}
	return err
}

func (h *WizardHandler) sendTextRemoveKeyboard(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
	_, err := h.ctx.Bot.Send(msg)
	if err != nil {
		log.Printf("sendTextRemoveKeyboard error chat_id=%d: %v", chatID, err)
	}
	return err
}

func (h *WizardHandler) sendMarkup(chatID int64, text string, markup any) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = markup
	_, err := h.ctx.Bot.Send(msg)
	if err != nil {
		log.Printf("sendMarkup error chat_id=%d: %v", chatID, err)
	}
	return err
}

func (h *WizardHandler) sendMainMenu(chatID int64) error {
	msg := tgbotapi.NewMessage(chatID, "Asosiy menyu:")
	msg.ReplyMarkup = templates.MainMenuKeyboard()
	_, err := h.ctx.Bot.Send(msg)
	if err != nil {
		log.Printf("sendMainMenu error chat_id=%d: %v", chatID, err)
	}
	return err
}

var (
	reDate = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	reTime = regexp.MustCompile(`^\d{2}:\d{2}$`)
	rePhone = regexp.MustCompile(`^[+0-9][0-9\s-]{6,}$`)
)

func isDateYYYYMMDD(s string) bool {
	if !reDate.MatchString(s) {
		return false
	}
	_, err := time.ParseInLocation("2006-01-02", s, time.Local)
	return err == nil
}

func isTimeHHMM(s string) bool {
	if !reTime.MatchString(s) {
		return false
	}
	hh, err := strconv.Atoi(s[0:2])
	if err != nil {
		return false
	}
	mm, err := strconv.Atoi(s[3:5])
	if err != nil {
		return false
	}
	return hh >= 0 && hh <= 23 && mm >= 0 && mm <= 59
}

func isPhoneLike(s string) bool { return rePhone.MatchString(strings.TrimSpace(s)) }

func normalizePhone(s string) string {
	s = strings.TrimSpace(s)
	// Collapse any Unicode whitespace (Telegram contacts often use spaced +998 90 …).
	s = strings.Join(strings.Fields(s), "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "\u00a0", "") // non-breaking space if any slipped through
	return s
}

