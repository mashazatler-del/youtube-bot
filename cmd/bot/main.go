package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"youtube-bot/internal/poller"
	"youtube-bot/internal/telegram"
	"youtube-bot/internal/youtube"
)

func main() {
	apiKey := requireEnv("YOUTUBE_API_KEY")
	handle := requireEnv("YOUTUBE_CHANNEL_HANDLE")
	botToken := requireEnv("TELEGRAM_BOT_TOKEN")
	chatID := requireEnv("TELEGRAM_CHAT_ID")

	yt := youtube.NewClient(apiKey, handle)
	tg := telegram.NewClient(botToken, chatID)

	p := poller.New(yt, tg, 30*time.Minute)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Printf("received signal %v, shutting down", sig)
		cancel()
	}()

	log.Println("bot started")
	p.Run(ctx)
	log.Println("bot stopped")
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return val
}
