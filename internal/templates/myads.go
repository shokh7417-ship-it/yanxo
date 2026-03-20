package templates

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"yanxo/internal/models"
)

func TaxiMyAdCard(ad models.Ad) string {
	from := deref(ad.FromCity)
	to := deref(ad.ToCity)
	date := deref(ad.RideDate)
	tm := deref(ad.DepartureTime)
	car := deref(ad.CarType)
	occ := 0
	tot := 4
	if ad.OccupiedSeats != nil {
		occ = *ad.OccupiedSeats
	}
	if ad.TotalSeats != nil {
		tot = *ad.TotalSeats
	}
	avail := tot - occ
	if avail < 0 {
		avail = 0
	}
	status := string(ad.Status)
	s := fmt.Sprintf("🚗 %s → %s\n📅 %s  🕒 %s\n🚘 %s\n👥 Band: %d/%d\n💺 Yana %d odam kerak\nHolat: %s",
		from, to, date, tm, car, occ, tot, avail, statusToUz(status),
	)
	if ad.Contact != nil && *ad.Contact != "" {
		s += "\n📞 " + *ad.Contact
	}
	return s
}

func TaxiManageInline(ad models.Ad) tgbotapi.InlineKeyboardMarkup {
	id := ad.ID
	row1 := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("➖ 1", "my:taxi:dec:"+id),
		tgbotapi.NewInlineKeyboardButtonData("➕ 1", "my:taxi:inc:"+id),
		tgbotapi.NewInlineKeyboardButtonData("⛔ To‘ldi", "my:taxi:full:"+id),
	)
	row2 := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🚗 Jo‘nab ketdim", "my:taxi:departed:"+id),
		tgbotapi.NewInlineKeyboardButtonData("🗑 O‘chirish", "my:taxi:delete:"+id),
	)
	return tgbotapi.NewInlineKeyboardMarkup(row1, row2)
}

func statusToUz(status string) string {
	switch strings.ToLower(status) {
	case "active":
		return "Aktiv"
	case "full":
		return "To‘ldi"
	case "expired":
		return "Tugagan"
	case "replaced":
		return "Yangilangan"
	case "deleted":
		return "O‘chirilgan"
	default:
		return status
	}
}

