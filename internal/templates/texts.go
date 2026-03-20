package templates

import (
	"fmt"
	"math"
)

const (
	BtnTaxiCreate   = "🚗 Taksi e’loni berish"
	BtnServiceCreate = "🔧 Usta e’loni berish"
	BtnSearch       = "🔎 Qidirish"
	BtnMyAds        = "📋 Mening e’lonlarim"
	BtnOpenChannel  = "📢 Kanalni ochish"

	BtnSearchTaxi    = "🚗 Taksi qidirish"
	BtnSearchService = "🔧 Usta qidirish"
	BtnBack          = "⬅️ Orqaga"
)

func WelcomeText() string {
	return "Assalomu alaykum! Kerakli bo‘limni tanlang.\n\n/cancel — istalgan bosqichda bekor qilish."
}

func ChannelLinkHint(channelID int64) string {
	// If channel has public username we can't infer it from ID.
	// For private/supergroup style links: https://t.me/c/<internal_id>/<message_id>
	// internal_id = abs(channelID) - 1000000000000 (when channelID starts with -100)
	return fmt.Sprintf("Channel ID: %d", channelID)
}

func FormatTaxiPreview(from, to, date, tm, car string, occupied, total int, contact *string) string {
	avail := total - occupied
	if avail < 0 {
		avail = 0
	}
	s := fmt.Sprintf("🚗 YO‘L E’LONI\n\n📍 %s → %s\n📅 %s\n🕒 %s\n🚘 %s\n👥 Band: %d/%d\n💺 Yana %d odam kerak",
		from, to, date, tm, car, occupied, total, avail)
	if contact != nil && *contact != "" {
		s += fmt.Sprintf("\n📞 %s", *contact)
	}
	return s
}

func FormatServicePreview(serviceType, area string, note *string, contact *string) string {
	s := fmt.Sprintf("🔧 USTA XIZMATI\n\n🛠 %s\n📍 %s", serviceType, area)
	if note != nil && *note != "" {
		s += fmt.Sprintf("\n📝 %s", *note)
	}
	if contact != nil && *contact != "" {
		s += fmt.Sprintf("\n📞 %s", *contact)
	}
	return s
}

func ClampInt(v, min, max int) int {
	return int(math.Max(float64(min), math.Min(float64(max), float64(v))))
}

