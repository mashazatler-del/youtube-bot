# YouTube Subscriber Notification Bot

## Overview

Telegram-бот на Go, который каждые 30 минут проверяет количество подписчиков YouTube-канала и присылает уведомление в Telegram, если число изменилось (рост или снижение).

## Конфигурация

Переменные окружения:

- `YOUTUBE_API_KEY` — ключ YouTube Data API v3
- `YOUTUBE_CHANNEL_HANDLE` — хэндл канала (например `mashazatler`)
- `TELEGRAM_BOT_TOKEN` — токен Telegram-бота
- `TELEGRAM_CHAT_ID` — ID чата для уведомлений

## Архитектура

Single-binary Go-сервис. Четыре пакета:

### main (cmd/bot/main.go)

- Читает конфиг из env
- Создаёт клиенты YouTube и Telegram
- Запускает poller
- Обрабатывает graceful shutdown (SIGINT/SIGTERM)

### youtube (internal/youtube/)

- `Client` — HTTP-клиент к YouTube Data API v3
- `GetSubscriberCount(ctx) (int64, error)` — запрос `channels.list?part=statistics&forHandle=...`, возвращает subscriberCount

### telegram (internal/telegram/)

- `Client` — HTTP-клиент к Telegram Bot API
- `SendMessage(ctx, text) error` — отправляет сообщение в указанный chat_id

### poller (internal/poller/)

- Тикер с интервалом 30 минут
- Хранит предыдущее значение подписчиков в переменной (`int64`)
- Первая проверка: запоминает значение, не уведомляет
- Последующие: сравнивает, при изменении отправляет сообщение

## Формат уведомлений

- Рост: `📈 Подписчики: 1500 → 1503 (+3)`
- Снижение: `📉 Подписчики: 1503 → 1500 (-3)`

## Хранение состояния

В памяти процесса. При перезапуске первая проверка запоминает текущее значение без уведомления.

## Зависимости

- Go standard library только (net/http, encoding/json, os, time, context, os/signal, fmt, log, strconv)
- Без внешних зависимостей

## YouTube API

- Endpoint: `GET https://www.googleapis.com/youtube/v3/channels?part=statistics&forHandle={handle}&key={apiKey}`
- Стоимость: 1 единица квоты за запрос
- 48 запросов/день при интервале 30 минут (квота — 10 000/день)

## Telegram Bot API

- Endpoint: `POST https://api.telegram.org/bot{token}/sendMessage`
- Body: `{"chat_id": "...", "text": "..."}`
