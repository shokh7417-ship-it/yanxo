package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	BotToken        string
	ChannelID       int64
	ChannelURL      string
	ChannelUsername string
	TursoDatabaseURL string
	TursoAuthToken  string
	// HTTPListenAddr is a bind address for the optional health HTTP server, e.g. ":8080".
	// Set via HEALTH_ADDR or PORT (Render). Empty means the health server is disabled.
	HTTPListenAddr string
}

func FromEnv() (Config, error) {
	var c Config

	c.BotToken = strings.TrimSpace(os.Getenv("BOT_TOKEN"))
	if c.BotToken == "" {
		return c, errors.New("BOT_TOKEN is required")
	}

	ch := strings.TrimSpace(os.Getenv("CHANNEL_ID"))
	if ch == "" {
		return c, errors.New("CHANNEL_ID is required")
	}
	chID, err := strconv.ParseInt(ch, 10, 64)
	if err != nil {
		return c, errors.New("CHANNEL_ID must be int64 (e.g. -100...)")
	}
	c.ChannelID = chID

	c.ChannelURL = strings.TrimSpace(os.Getenv("CHANNEL_URL"))
	c.ChannelUsername = strings.TrimSpace(os.Getenv("CHANNEL_USERNAME"))
	if c.ChannelUsername == "" && c.ChannelURL != "" {
		c.ChannelUsername = extractChannelUsername(c.ChannelURL)
	}

	c.TursoDatabaseURL = strings.TrimSpace(os.Getenv("TURSO_DATABASE_URL"))
	if c.TursoDatabaseURL == "" {
		return c, errors.New("TURSO_DATABASE_URL is required")
	}

	c.TursoAuthToken = strings.TrimSpace(os.Getenv("TURSO_AUTH_TOKEN"))
	if c.TursoAuthToken == "" {
		return c, errors.New("TURSO_AUTH_TOKEN is required")
	}

	addr := strings.TrimSpace(os.Getenv("HEALTH_ADDR"))
	port := strings.TrimSpace(os.Getenv("PORT"))
	if addr == "" && port != "" {
		addr = ":" + port
	}
	c.HTTPListenAddr = addr

	return c, nil
}

func extractChannelUsername(channelURL string) string {
	// Accept: https://t.me/<username> , t.me/<username> , @username
	s := strings.TrimSpace(channelURL)
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "t.me/")
	s = strings.TrimPrefix(s, "telegram.me/")
	s = strings.TrimPrefix(s, "@")
	if i := strings.IndexByte(s, '/'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

