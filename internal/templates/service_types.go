package templates

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// Usta e’loni: kategoriya va aniq xizmat turlari (reply keyboard).

const (
	ServiceCatBuildBtn = "🔧 Qurilish va ta’mirlash ustalari"
	ServiceCatAutoBtn  = "🚗 Avto ustalar"
	ServiceCatWoodBtn  = "🪵 Mebel va yog‘och ustalari"

	ServiceTypeOtherBtn = "✏️ Boshqa xizmat (o‘zim yozaman)"
	ServicePickBackBtn  = "⬅️ Kategoriyalar"
	ServiceWizardCancel = "❌ Bekor qilish"
)

// PickCategory session qiymatlari (ichki).
const (
	ServicePickCatBuild = "build"
	ServicePickCatAuto  = "auto"
	ServicePickCatWood  = "wood"
)

var serviceTypesBuild = []string{
	"Santexnik (quvurlar, suv tizimi)",
	"Elektrik (elektr tarmoqlari)",
	"Payvandchi (svarchik)",
	"G‘isht teruvchi (mason)",
	"Suvoqchi (shpaklyovka, suvoq ishlari)",
	"Kafelchi (plitka qo‘yuvchi)",
	"Tom yopuvchi ustalar",
}

var serviceTypesAuto = []string{
	"Motor ustasi (dvigatel)",
	"Hodovoy ustasi (podveska)",
	"Elektrik (avtoelektrik)",
	"Diagnostika ustasi",
	"Kuzov ustasi (kraska, rihlovka)",
	"Shina ustasi (balansirovka, montaj)",
}

var serviceTypesWood = []string{
	"Duradgor (stolyar)",
	"Mebel ustasi (shkaf, stol yasash)",
	"Eshik-deraza ustasi",
}

// ServiceCategoryKeyboard — birinchi qadam: yo‘nalish tanlash.
func ServiceCategoryKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(ServiceCatBuildBtn)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(ServiceCatAutoBtn)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(ServiceCatWoodBtn)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(ServiceTypeOtherBtn)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(ServiceWizardCancel)),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = true
	return kb
}

// ServicePickKeyboard — tanlangan kategoriya bo‘yicha xizmatlar + orqaga.
func ServicePickKeyboard(category string) tgbotapi.ReplyKeyboardMarkup {
	var list []string
	switch category {
	case ServicePickCatBuild:
		list = serviceTypesBuild
	case ServicePickCatAuto:
		list = serviceTypesAuto
	case ServicePickCatWood:
		list = serviceTypesWood
	default:
		list = nil
	}
	var rows [][]tgbotapi.KeyboardButton
	for i := 0; i < len(list); i += 2 {
		if i+1 < len(list) {
			rows = append(rows, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(list[i]),
				tgbotapi.NewKeyboardButton(list[i+1]),
			))
		} else {
			rows = append(rows, tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(list[i]),
			))
		}
	}
	rows = append(rows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(ServicePickBackBtn)))
	kb := tgbotapi.NewReplyKeyboard(rows...)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = true
	return kb
}

// ServiceCustomTypeKeyboard — erkin mat uchun faqat orqaga.
func ServiceCustomTypeKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(ServicePickBackBtn)),
		tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(ServiceWizardCancel)),
	)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = true
	return kb
}

// ServiceCategoryFromButton — tugma matni → ichki kategoriya kodi.
func ServiceCategoryFromButton(text string) (cat string, ok bool) {
	switch text {
	case ServiceCatBuildBtn:
		return ServicePickCatBuild, true
	case ServiceCatAutoBtn:
		return ServicePickCatAuto, true
	case ServiceCatWoodBtn:
		return ServicePickCatWood, true
	default:
		return "", false
	}
}

// IsKnownServicePick reports whether text is a valid option for this pick step.
func IsKnownServicePick(category, text string) bool {
	var list []string
	switch category {
	case ServicePickCatBuild:
		list = serviceTypesBuild
	case ServicePickCatAuto:
		list = serviceTypesAuto
	case ServicePickCatWood:
		list = serviceTypesWood
	default:
		return false
	}
	for _, s := range list {
		if s == text {
			return true
		}
	}
	return false
}
