package bot

import (
	"context"
	"errors"
	"log"
	"math"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"yanxo/internal/config"
	"yanxo/internal/location"
	libsqlrepo "yanxo/internal/repository/libsql"
	"yanxo/internal/service"
	"yanxo/internal/session"
	"yanxo/internal/utils"
)

type App struct {
	cfg   config.Config
	bot   *tgbotapi.BotAPI
	db    *libsqlrepo.DB
	ads   *service.AdsService
	users *service.UsersService
	store *session.Store
	rt    *Router
}

func NewApp(cfg config.Config) (*App, error) {
	tg, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, err
	}
	tg.Debug = false

	return &App{
		cfg: cfg,
		bot: tg,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	// Listen on PORT before DB/migrations/Telegram — otherwise Render health checks get 502 during startup.
	var healthSrv *http.Server
	if addr := a.cfg.HTTPListenAddr; addr != "" {
		healthSrv = startHealthServer(addr)
	}

	for attempt := 1; ; attempt++ {
		err := a.initRuntime(ctx)
		if err == nil {
			break
		}
		wait := time.Duration(math.Min(float64(attempt), 30)) * time.Second
		log.Printf("app init failed: %v (retry in %s)", err, wait)
		select {
		case <-ctx.Done():
			if healthSrv != nil {
				shutdownHealthServer(context.Background(), healthSrv)
			}
			return nil
		case <-time.After(wait):
		}
	}

	// Set command panel (menu next to input)
	commands := []tgbotapi.BotCommand{
		tgbotapi.BotCommand{Command: "start", Description: "Botni ishga tushirish"},
		tgbotapi.BotCommand{Command: "cancel", Description: "Joriy amalni bekor qilish"},
	}
	// Set for default scope and for all private chats (Telegram Desktop is usually private chat).
	if _, err := a.bot.Request(tgbotapi.NewSetMyCommands(commands...)); err != nil {
		log.Printf("setMyCommands(default) warning: %v", err)
	}
	if _, err := a.bot.Request(tgbotapi.NewSetMyCommandsWithScope(tgbotapi.NewBotCommandScopeAllPrivateChats(), commands...)); err != nil {
		log.Printf("setMyCommands(all_private_chats) warning: %v", err)
	}
	if got, err := a.bot.GetMyCommands(); err != nil {
		log.Printf("getMyCommands warning: %v", err)
	} else {
		log.Printf("commands registered: %v", got)
	}

	log.Printf("bot started as @%s", a.bot.Self.UserName)

	err := a.loop(ctx)
	if healthSrv != nil {
		shutdownHealthServer(context.Background(), healthSrv)
	}
	_ = a.db.Close()
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func (a *App) initRuntime(ctx context.Context) error {
	// Webhook + long polling aralashmasin; eski pending update’lar tozalanadi.
	if _, whErr := a.bot.Request(tgbotapi.DeleteWebhookConfig{DropPendingUpdates: true}); whErr != nil {
		log.Printf("deleteWebhook warning: %v", whErr)
	} else {
		log.Printf("telegram: webhook o‘chirildi, long polling rejimi")
	}

	db, err := libsqlrepo.Open(ctx, a.cfg.TursoDatabaseURL, a.cfg.TursoAuthToken)
	if err != nil {
		return err
	}
	a.db = db

	migrationsDir := filepath.Join(".", "migrations")
	var migErr error
	for attempt := 1; attempt <= 3; attempt++ {
		migErr = libsqlrepo.RunMigrations(ctx, a.db.SQL, migrationsDir)
		if migErr == nil {
			break
		}
		log.Printf("migrations attempt=%d failed: %v", attempt, migErr)
		// Backoff for transient network/TLS issues.
		select {
		case <-ctx.Done():
			break
		case <-time.After(time.Duration(attempt) * 2 * time.Second):
		}
	}
	if migErr != nil {
		_ = a.db.Close()
		return migErr
	}

	repo := libsqlrepo.NewAdsRepo(a.db.SQL)
	a.ads = service.NewAdsService(repo, utils.RealClock{})
	usersRepo := libsqlrepo.NewUsersRepo(a.db.SQL)
	a.users = service.NewUsersService(usersRepo)
	a.store = session.NewStore()

	locRepo := libsqlrepo.NewLocationRepo(a.db.SQL)
	var seedErr error
	for attempt := 1; attempt <= 3; attempt++ {
		seedErr = location.SeedLocations(ctx, locRepo)
		if seedErr == nil {
			break
		}
		log.Printf("seed attempt=%d failed: %v", attempt, seedErr)
		select {
		case <-ctx.Done():
			break
		case <-time.After(time.Duration(attempt) * 2 * time.Second):
		}
	}
	if seedErr != nil {
		_ = a.db.Close()
		return seedErr
	}
	resolver := location.NewResolver(locRepo)

	a.rt = NewRouter(a.cfg, a.bot, a.ads, a.users, a.store, resolver)
	return nil
}

func (a *App) loop(ctx context.Context) error {
	var offset int
	backoff := 2 * time.Second

	for {
		select {
		case <-ctx.Done():
			a.bot.StopReceivingUpdates()
			return ctx.Err()
		default:
		}

		u := tgbotapi.NewUpdate(offset)
		u.Timeout = 60

		updates, err := a.bot.GetUpdates(u)
		if err != nil {
			if strings.Contains(err.Error(), "Conflict") {
				log.Printf("getUpdates Conflict: boshqa joyda ham shu BOT_TOKEN bilan bot ishlamoqda (mahalliy go run + Render yoki 2 ta deploy). Bittasini to‘xtating. Xato: %v", err)
			} else {
				log.Printf("getUpdates error: %v (retry in %s)", err, backoff)
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			if backoff < 10*time.Second {
				backoff += 1 * time.Second
			}
			continue
		}
		backoff = 2 * time.Second

		for _, upd := range updates {
			if upd.UpdateID >= offset {
				offset = upd.UpdateID + 1
			}
			a.rt.HandleUpdate(ctx, upd)
		}
	}
}

