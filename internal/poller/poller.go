package poller

import (
	"context"
	"fmt"
	"log"
	"time"
)

type SubscriberCounter interface {
	GetSubscriberCount(ctx context.Context) (int64, error)
}

type MessageSender interface {
	SendMessage(ctx context.Context, text string) error
}

type Poller struct {
	youtube  SubscriberCounter
	telegram MessageSender
	interval time.Duration
}

func New(youtube SubscriberCounter, telegram MessageSender, interval time.Duration) *Poller {
	return &Poller{
		youtube:  youtube,
		telegram: telegram,
		interval: interval,
	}
}

func (p *Poller) Run(ctx context.Context) {
	count, err := p.youtube.GetSubscriberCount(ctx)
	if err != nil {
		log.Printf("initial check failed: %v", err)
		return
	}
	prev := count
	log.Printf("initial subscriber count: %d", prev)

	startMsg := fmt.Sprintf("🚀 Бот запущен! Подписчиков: %d", prev)
	if err := p.telegram.SendMessage(ctx, startMsg); err != nil {
		log.Printf("send start message failed: %v", err)
	}

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count, err := p.youtube.GetSubscriberCount(ctx)
			if err != nil {
				log.Printf("check failed: %v", err)
				continue
			}

			if count != prev {
				diff := count - prev
				msg := formatMessage(prev, count, diff)
				if err := p.telegram.SendMessage(ctx, msg); err != nil {
					log.Printf("send message failed: %v", err)
				}
				prev = count
			}
		}
	}
}

func formatMessage(prev, current, diff int64) string {
	if diff > 0 {
		return fmt.Sprintf("📈 Подписчики: %d → %d (+%d)", prev, current, diff)
	}
	return fmt.Sprintf("📉 Подписчики: %d → %d (%d)", prev, current, diff)
}
