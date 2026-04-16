package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"youtube-bot/internal/poller"
	"youtube-bot/internal/telegram"
	"youtube-bot/internal/youtube"
)

func loadEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

func main() {
	loadEnv(".env")

	apiKey := requireEnv("YOUTUBE_API_KEY")
	handle := requireEnv("YOUTUBE_CHANNEL_HANDLE")
	botToken := requireEnv("TELEGRAM_BOT_TOKEN")

	yt := youtube.NewClient(apiKey, handle)
	tg := telegram.NewClient(botToken)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Printf("received signal %v, shutting down", sig)
		cancel()
	}()

	if err := tg.WaitForStart(ctx); err != nil {
		log.Fatalf("waiting for /start: %v", err)
	}

	p := poller.New(yt, tg, 30*time.Minute)
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
