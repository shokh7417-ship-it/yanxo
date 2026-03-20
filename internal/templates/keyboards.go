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

func CityKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
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
	)
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
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("⏭ O‘tkazib yuborish")),
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
			tgbotapi.NewInlineKeyboardButtonData("⏭ O‘tkazib yuborish", "contact:skip"),
		),
	)
}

func ContactChoiceNoUsername() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📞 Telefon kiritish", "contact:enter_phone"),
			tgbotapi.NewInlineKeyboardButtonData("⏭ O‘tkazib yuborish", "contact:skip"),
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

