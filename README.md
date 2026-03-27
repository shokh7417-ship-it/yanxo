# Yanxo - Telegram Marketplace Bot (Go + Turso)

Yanxo - Uzbekistan lokal bozoriga mos Telegram bot. Bot ichida foydalanuvchilar:

- taksi yol e'lonini joylaydi va qidiradi;
- usta/xizmat e'lonini joylaydi va qidiradi;
- o'z e'lonlarini boshqaradi.

Kanal esa ommaviy feed sifatida ishlaydi: bot e'lonni kanalga post qiladi, foydalanuvchi esa post link orqali ochadi.

## Asosiy imkoniyatlar

- `/start` menyu: e'lon berish, qidirish, mening e'lonlarim, kanalni ochish.
- Taksi e'loni wizard: shaharlar, sana, vaqt, mashina turi, o'rinlar, aloqa, preview, kanalga joylash.
- Usta e'loni wizard: yo'nalish, aniq xizmat turi, hudud, izoh, aloqa, preview, kanalga joylash.
- Bot ichida qidiruv: taksi va usta e'lonlari bo'yicha natijalar, postni ochish va aloqa tugmalari.
- "Mening e'lonlarim": taksi e'lonini tez boshqarish (o'rin +/-, to'ldi, jo'nab ketdi, o'chirish).
- Health endpointlar: `/`, `/health`, `/healthz`.

## Texnologiyalar

- Go 1.24+
- Telegram Bot API (`go-telegram-bot-api/v5`)
- Turso / libSQL
- SQL migratsiyalar (`migrations/`)

## Loyiha tuzilmasi

- `cmd/bot` - dastur entrypoint.
- `internal/bot` - app lifecycle, polling, router, health server.
- `internal/handlers` - start/menu, wizard flow, my ads callback.
- `internal/service` - biznes mantiq (`AdsService`).
- `internal/repository` - repository abstraksiyalari va libSQL implementatsiyasi.
- `internal/location` - shahar normalizatsiyasi, matching, resolve/seed.
- `internal/templates` - bot matnlari, keyboardlar, post/card formatlari.
- `internal/session` - in-memory wizard holati.
- `migrations` - DB schema migratsiyalari.

## Talablar

- Go `1.24` yoki yuqori
- Telegram bot token (`@BotFather`)
- Telegram kanal (public yoki private), bot kanalda admin bo'lishi kerak
- Turso database va auth token

## Muhit o'zgaruvchilari

`.env.example` ni `.env` ga nusxalab to'ldiring:

```bash
BOT_TOKEN=123456:ABCDEF_your_bot_token
CHANNEL_ID=-1001234567890
CHANNEL_URL=https://t.me/your_channel
CHANNEL_USERNAME=your_channel
TURSO_DATABASE_URL=libsql://your-db-name.turso.io
TURSO_AUTH_TOKEN=your_turso_auth_token
```

Majburiy:

- `BOT_TOKEN`
- `CHANNEL_ID` (masalan `-1001234567890`)
- `TURSO_DATABASE_URL`
- `TURSO_AUTH_TOKEN`

Ixtiyoriy:

- `CHANNEL_URL` - "Kanalni ochish" tugmasi uchun.
- `CHANNEL_USERNAME` - post linkni qulay formatda yasash uchun.
- `PORT` yoki `HEALTH_ADDR` - health HTTP serverni yoqish uchun.
  - `PORT` bo'lsa, app avtomatik `:<PORT>` da eshitadi.
  - `HEALTH_ADDR` bo'lsa (masalan `:8080`), shu qiymat ishlatiladi.

## Local ishga tushirish

```bash
go mod tidy
go run ./cmd/bot
```

Nimalar bo'ladi:

- app startda webhookni o'chiradi va long polling rejimida ishlaydi;
- DB ga ulanadi;
- migratsiyalarni avtomatik ishga tushiradi;
- location ma'lumotlarini seed qiladi;
- update loopni boshlaydi.

## Docker

```bash
docker build -t yanxo-bot .
docker run --env-file .env -e PORT=8080 yanxo-bot
```

Health endpointlar:

- `GET /`
- `GET /health`
- `GET /healthz`

Javob: `ok` (oddiy matn, `text/plain`)

## Render deploy

2 xil variant:

1. **Web Service**
   - `PORT` Render tomonidan beriladi (app health serverni avtomatik yoqadi).
   - Health check path: `/health`.
2. **Background Worker**
   - HTTP kerak bo'lmasa `PORT`/`HEALTH_ADDR` bermasdan ishlatish mumkin.

## Kanal va Telegram sozlamalari

### `CHANNEL_ID` olish

- Botni kanalga admin qiling.
- Kanalga bitta post yuboring.
- Shu postni `@userinfobot` ga forward qiling.
- Bot qaytargan `-100...` formatdagi ID ni `CHANNEL_ID` ga yozing.

### Kanal izohlari (comments)

Kanal postlarida commentlar ko'rinishi uchun:

- Channel -> Settings -> Discussion -> Add a group

## Muhim ishlash eslatmalari

- Bir vaqtning o'zida faqat **bitta** jarayon shu `BOT_TOKEN` bilan `getUpdates` qilishi kerak.
- Aks holda logda `Conflict: terminated by other getUpdates request` chiqadi.
- Agar Render ishlayotgan bo'lsa, local `go run ./cmd/bot` ni to'xtating (yoki aksincha).

## Qisqa troubleshooting

- `config: BOT_TOKEN is required` -> `.env` to'g'ri joyda va to'g'ri to'ldirilganini tekshiring.
- `CHANNEL_ID must be int64` -> `CHANNEL_ID` qiymati `-100...` formatda bo'lishi kerak.
- Health check 502 (deploy boshida) -> qisqa muddatli bo'lishi mumkin; doimiy bo'lsa loglarni tekshiring.
- Kanal post edit/delete ishlamasa -> bot kanalda admin va kerakli huquqlarga ega ekanini tekshiring.

## MVP doirasi

Hozirgi implementatsiya asosan quyidagilarni yopadi:

- Taksi e'loni yaratish/qidirish/boshqarish;
- Usta e'loni yaratish/qidirish;
- Kanalga post qilish va post link ochish;
- Turso asosida saqlash va retrieval.