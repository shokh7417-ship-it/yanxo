package templates

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func MainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnTaxiCreate),
			tgbotapi.NewKeyboardButton(BtnServiceCreate),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnSearch),
			tgbotapi.NewKeyboardButton(BtnMyAds),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnOpenChannel),
		),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = false
	return kb
}

func RoleSelectionKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnRoleProvider),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnRoleClient),
		),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = false
	return kb
}

func ProviderMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnCreateAd),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnMyAds),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnRoleSwitch),
		),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = false
	return kb
}

func ClientMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnSearch),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnRoleSwitch),
		),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = false
	return kb
}

func ProviderCreateKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnTaxiCreate),
			tgbotapi.NewKeyboardButton(BtnServiceCreate),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnBack),
		),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = false
	return kb
}

func SearchMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnSearchTaxi),
			tgbotapi.NewKeyboardButton(BtnSearchService),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnBack),
		),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = true
	return kb
}

// cityPlaceKeyboardRows — taksi/usta uchun umumiy shahar/hudud tugmalari.
func cityPlaceKeyboardRows() [][]tgbotapi.KeyboardButton {
	return [][]tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Xovos"),
			tgbotapi.NewKeyboardButton("Yangiyer"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Guliston"),
			tgbotapi.NewKeyboardButton("Toshkent"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Boshqa"),
		),
	}
}

func CityKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(cityPlaceKeyboardRows()...)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = true
	return kb
}

// ServicePlaceKeyboard — usta e’loni / usta qidiruv: hudud tanlash (shahar + «Boshqa»).
func ServicePlaceKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return CityKeyboard()
}

// ServiceSearchAreaKeyboard — usta qidiruv: hudud tugmalari + Orqaga / Bekor.
func ServiceSearchAreaKeyboard() tgbotapi.ReplyKeyboardMarkup {
	rows := cityPlaceKeyboardRows()
	rows = append(rows, tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(BtnBack),
		tgbotapi.NewKeyboardButton("❌ Bekor qilish"),
	))
	kb := tgbotapi.NewReplyKeyboard(rows...)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = true
	return kb
}

func ServiceSearchNoResultsKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(BtnBack),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ Bekor qilish"),
		),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = true
	return kb
}

func SkipKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⏭ O‘tkazib yuborish"),
		),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = true
	return kb
}

func TaxiDateKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ Bugun"),
		),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = true
	return kb
}

func CarTypeKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Cobalt"),
			tgbotapi.NewKeyboardButton("Gentra"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Nexia"),
			tgbotapi.NewKeyboardButton("Spark"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Malibu"),
			tgbotapi.NewKeyboardButton("Boshqa"),
		),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = true
	return kb
}

func TotalSeatsKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("✅ 4"),
			tgbotapi.NewKeyboardButton("3"),
			tgbotapi.NewKeyboardButton("5"),
		),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = true
	return kb
}

func PhoneRequestKeyboard() tgbotapi.ReplyKeyboardMarkup {
	btn := tgbotapi.NewKeyboardButtonContact("📲 Telefonni ulashish")
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(btn),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = true
	return kb
}

func ContactChoiceWithUsername() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Username ishlatish", "contact:use_username"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📞 Telefon kiritaman", "contact:enter_phone"),
		),
	)
}

func ContactChoiceNoUsername() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📞 Telefon kiritish", "contact:enter_phone"),
		),
	)
}

func ConfirmKeyboard(kind string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Joylash", "confirm:"+kind),
			tgbotapi.NewInlineKeyboardButtonData("❌ Bekor qilish", "cancel:"+kind),
		),
	)
}

