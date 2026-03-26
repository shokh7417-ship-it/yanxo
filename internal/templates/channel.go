package templates

import (
	"fmt"
	"html"
	"math"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"yanxo/internal/models"
)

func TaxiChannelPost(ad models.Ad) string {
	from, to, date, tm, car := safe(ad.FromCity), safe(ad.ToCity), safe(ad.RideDate), safe(ad.DepartureTime), safe(ad.CarType)
	occ, tot := safeInt(ad.OccupiedSeats), safeIntDefault(ad.TotalSeats, 4)
	avail := tot - occ
	if avail < 0 {
		avail = 0
	}
	title := "🚗 YO‘L E’LONI"
	statusLine := taxiStatusLine(ad, occ, tot, avail)
	body := fmt.Sprintf("%s\n\n📍 %s → %s\n📅 %s\n🕒 %s\n🚘 %s\n👥 Band: %d/%d\n💺 Yana %d odam kerak\n%s",
		title,
		from, to,
		date,
		tm,
		car,
		occ, tot,
		avail,
		statusLine,
	)
	if ad.Contact != nil && *ad.Contact != "" {
		body += "\n📞 " + html.EscapeString(*ad.Contact)
	}
	return body
}

func ServiceChannelPost(ad models.Ad) string {
	title := "🔧 USTA XIZMATI"
	st := safe(ad.ServiceType)
	area := safe(ad.Area)
	body := fmt.Sprintf("%s\n\n🛠 %s\n📍 %s", title, st, area)
	if ad.Note != nil && *ad.Note != "" {
		body += "\n📝 " + html.EscapeString(*ad.Note)
	}
	if ad.Contact != nil && *ad.Contact != "" {
		body += "\n📞 " + html.EscapeString(*ad.Contact)
	}
	return body
}

func PostOpenInline(ad models.Ad, channelID int64, channelUsername string) tgbotapi.InlineKeyboardMarkup {
	link := ChannelPostLink(channelID, channelUsername, ad.ChannelMessageID)
	if link == "" {
		// if not known yet, show a noop button
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📄 Postni ochish", "noop"),
			),
		)
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("📄 Postni ochish", link),
		),
	)
}

func ChannelPostLink(channelID int64, channelUsername string, channelMessageID *int) string {
	if channelMessageID == nil || *channelMessageID <= 0 {
		return ""
	}
	if channelUsername = strings.TrimSpace(channelUsername); channelUsername != "" {
		channelUsername = strings.TrimPrefix(channelUsername, "@")
		return fmt.Sprintf("https://t.me/%s/%d", channelUsername, *channelMessageID)
	}
	// For private channels/supergroups with -100... id:
	// https://t.me/c/<internal_id>/<message_id>
	// internal_id = abs(channelID) - 1000000000000
	s := strconv.FormatInt(channelID, 10)
	if !strings.HasPrefix(s, "-100") {
		return ""
	}
	abs := int64(math.Abs(float64(channelID)))
	internalID := abs - 1000000000000
	if internalID <= 0 {
		return ""
	}
	return fmt.Sprintf("https://t.me/c/%d/%d", internalID, *channelMessageID)
}

func taxiStatusLine(ad models.Ad, occ, tot, avail int) string {
	switch ad.Status {
	case models.StatusFull:
		return "⛔ Holat: To‘ldi"
	case models.StatusExpired:
		return "🕒 Holat: Tugagan"
	case models.StatusDeleted:
		return "🗑 Holat: O‘chirilgan"
	case models.StatusReplaced:
		return "🔄 Holat: Yangilangan"
	default:
		if avail <= 0 || occ >= tot {
			return "⛔ Holat: To‘ldi"
		}
		return "✅ Holat: Aktiv"
	}
}

func safe(p *string) string {
	if p == nil {
		return "—"
	}
	return html.EscapeString(*p)
}

func safeInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

func safeIntDefault(p *int, def int) int {
	if p == nil {
		return def
	}
	return *p
}

