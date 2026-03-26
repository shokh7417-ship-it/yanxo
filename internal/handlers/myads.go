package handlers

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"yanxo/internal/models"
	"yanxo/internal/templates"
)

type MyAdsHandler struct {
	ctx Context
}

func NewMyAdsHandler(ctx Context) *MyAdsHandler { return &MyAdsHandler{ctx: ctx} }

func (h *MyAdsHandler) Show(ctx context.Context, m *tgbotapi.Message) error {
	userID := m.From.ID

	// Taxi ads: active/full/expired
	cat := models.CategoryRoad
	ads, err := h.ctx.Ads.ListByUser(ctx, userID, &cat, []models.AdStatus{
		models.StatusActive, models.StatusFull, models.StatusExpired,
	}, 50)
	if err != nil {
		_, _ = h.ctx.Bot.Send(tgbotapi.NewMessage(m.Chat.ID, "Xatolik: e’lonlar yuklanmadi. Keyinroq urinib ko‘ring."))
		return err
	}

	if len(ads) == 0 {
		msg := tgbotapi.NewMessage(m.Chat.ID, "Sizda hozircha e’lon yo‘q.")
		msg.ReplyMarkup = templates.ProviderMenuKeyboard()
		_, _ = h.ctx.Bot.Send(msg)
		return nil
	}
	// Old -> new ordering in "Mening e'lonlarim".
	sort.SliceStable(ads, func(i, j int) bool {
		return ads[i].CreatedAt.Before(ads[j].CreatedAt)
	})

	header := tgbotapi.NewMessage(m.Chat.ID, fmt.Sprintf("📋 Mening e’lonlarim (taksi): %d ta", len(ads)))
	header.ReplyMarkup = templates.ProviderMenuKeyboard()
	_, _ = h.ctx.Bot.Send(header)

	for _, ad := range ads {
		card := templates.TaxiMyAdCard(ad)
		msg := tgbotapi.NewMessage(m.Chat.ID, card)
		msg.ReplyMarkup = templates.TaxiManageInline(ad)
		_, _ = h.ctx.Bot.Send(msg)
	}
	return nil
}

func (h *MyAdsHandler) HandleCallback(ctx context.Context, q *tgbotapi.CallbackQuery, msg *tgbotapi.Message) bool {
	// data: my:taxi:<action>:<adID>
	data := q.Data
	const pfx = "my:taxi:"
	if len(data) < len(pfx) || data[:len(pfx)] != pfx {
		return false
	}
	rest := data[len(pfx):]
	parts := splitN(rest, ':', 2)
	if len(parts) != 2 {
		_, _ = h.ctx.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Xatolik: noto‘g‘ri tugma."))
		return true
	}
	action := parts[0]
	adID := parts[1]

	userID := q.From.ID
	var (
		ad  models.Ad
		err error
	)
	switch action {
	case "inc":
		ad, err = h.ctx.Ads.UpdateTaxiOccupiedDelta(ctx, adID, userID, 1)
	case "dec":
		ad, err = h.ctx.Ads.UpdateTaxiOccupiedDelta(ctx, adID, userID, -1)
	case "full":
		ad, err = h.ctx.Ads.SetTaxiFull(ctx, adID, userID)
	case "departed":
		ad, err = h.ctx.Ads.SetStatus(ctx, adID, userID, models.StatusExpired)
	case "delete":
		// Delete flow only: mark ad deleted, delete channel post, update chat UI, then return. Do NOT edit channel.
		before, gerr := h.ctx.Ads.Get(ctx, adID)
		if gerr != nil {
			_, _ = h.ctx.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Xatolik: amal bajarilmadi."))
			return true
		}
		if before.UserID != userID {
			_, _ = h.ctx.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Xatolik: amal bajarilmadi."))
			return true
		}
		if derr := h.ctx.Ads.MarkDeleted(ctx, adID, userID); derr != nil {
			_, _ = h.ctx.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Xatolik: amal bajarilmadi."))
			return true
		}
		if before.ChannelMessageID != nil && *before.ChannelMessageID > 0 {
			del := tgbotapi.NewDeleteMessage(h.ctx.Cfg.ChannelID, *before.ChannelMessageID)
			if _, derr2 := h.ctx.Bot.Request(del); derr2 != nil {
				if !strings.Contains(derr2.Error(), "message to delete not found") {
					log.Printf("channel delete error ad_id=%s channel_msg_id=%d: %v", adID, *before.ChannelMessageID, derr2)
					_, _ = h.ctx.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "⚠️ Kanaldagi postni o‘chirib bo‘lmadi. Bot kanalda “Delete messages” huquqiga ega ekanini tekshiring."))
				}
			}
		}
		// Update chat: show "E'lon o'chirildi" and remove inline buttons. Do NOT call editChannelPostTaxi.
		edit := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, "✅ E'lon o'chirildi.")
		_, _ = h.ctx.Bot.Send(edit)
		editKB := tgbotapi.NewEditMessageReplyMarkup(msg.Chat.ID, msg.MessageID, tgbotapi.NewInlineKeyboardMarkup())
		_, _ = h.ctx.Bot.Send(editKB)
		return true
	default:
		_, _ = h.ctx.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Bu amal hozircha mavjud emas."))
		return true
	}
	if err != nil {
		_, _ = h.ctx.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Xatolik: amal bajarilmadi."))
		return true
	}

	// Edit channel post text (minor updates and status changes)
	if err := h.editChannelPostTaxi(ad); err != nil {
		// If message is gone already, ignore; otherwise warn.
		if !strings.Contains(err.Error(), "message to edit not found") {
			log.Printf("channel edit error ad_id=%s channel_msg_id=%v: %v", ad.ID, ad.ChannelMessageID, err)
			_, _ = h.ctx.Bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "⚠️ Kanaldagi postni yangilab bo‘lmadi. Bot kanalda admin va edit huquqiga ega ekanini tekshiring."))
		}
	}

	// Update card in chat (edit message text + buttons)
	newText := templates.TaxiMyAdCard(ad)
	edit := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, newText)
	_, _ = h.ctx.Bot.Send(edit)

	editKB := tgbotapi.NewEditMessageReplyMarkup(msg.Chat.ID, msg.MessageID, templates.TaxiManageInline(ad))
	_, _ = h.ctx.Bot.Send(editKB)
	return true
}

func (h *MyAdsHandler) editChannelPostTaxi(ad models.Ad) error {
	if ad.ChannelMessageID == nil || *ad.ChannelMessageID <= 0 {
		return nil
	}
	txt := templates.TaxiChannelPost(ad)
	edit := tgbotapi.NewEditMessageText(h.ctx.Cfg.ChannelID, *ad.ChannelMessageID, txt)
	edit.ParseMode = "HTML"
	_, err := h.ctx.Bot.Send(edit)
	return err
}

func splitN(s string, sep byte, n int) []string {
	var out []string
	start := 0
	for i := 0; i < len(s) && len(out) < n; i++ {
		if s[i] == sep {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}

