package templates

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"yanxo/internal/models"
)

func TaxiSearchResultCard(ad models.Ad) string {
	from := deref(ad.FromCity)
	to := deref(ad.ToCity)
	tm := deref(ad.DepartureTime)
	car := deref(ad.CarType)
	avail := 0
	if ad.TotalSeats != nil && ad.OccupiedSeats != nil {
		avail = *ad.TotalSeats - *ad.OccupiedSeats
		if avail < 0 {
			avail = 0
		}
	}
	s := fmt.Sprintf("🚗 %s → %s\n🕒 %s\n🚘 %s\n💺 Yana %d odam kerak", from, to, tm, car, avail)
	if ad.Contact != nil && *ad.Contact != "" {
		s += "\n📞 " + *ad.Contact
	}
	return s
}

func ServiceSearchResultCard(ad models.Ad) string {
	st := deref(ad.ServiceType)
	area := deref(ad.Area)
	s := fmt.Sprintf("🔧 %s\n📍 %s", st, area)
	if ad.Note != nil && *ad.Note != "" {
		s += "\n📝 " + *ad.Note
	}
	if ad.Contact != nil && *ad.Contact != "" {
		s += "\n📞 " + *ad.Contact
	}
	return s
}

func SearchResultInline(ad models.Ad, channelID int64, channelUsername string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Post open
	link := ChannelPostLink(channelID, channelUsername, ad.ChannelMessageID)
	if link != "" {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("📄 Postni ochish", link),
		))
	} else {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📄 Postni ochish", "sr:post:"+ad.ID),
		))
	}

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func deref(p *string) string {
	if p == nil {
		return "—"
	}
	return *p
}

