package telegram

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type MessageKey struct {
	ChatID    int64
	MessageID int
}

type Telegram struct {
	Bot    *bot.Bot
	chatID int64

	pending sync.Map // MessageKey -> chan string
}

func NewTelegram(token string, chatID int64) (*Telegram, error) {
	t := &Telegram{
		chatID: chatID,
	}

	b, err := bot.New(
		token,
		bot.WithDefaultHandler(t.handleUpdate),
	)
	if err != nil {
		return nil, err
	}

	t.Bot = b

	return t, nil
}

// handleUpdate receives Telegram updates via long polling.
func (t *Telegram) handleUpdate(
	ctx context.Context,
	b *bot.Bot,
	update *models.Update,
) {
	if update == nil {
		return
	}

	msg := update.Message
	if msg == nil || msg.Text == "" {
		return
	}

	// Accept messages only from our chat.
	if msg.Chat.ID != t.chatID {
		return
	}

	// If the user used Telegram's "Reply" feature, route to that specific
	// waiter. Otherwise fall back to delivering the text to any pending
	// waiter in this chat (private 1-on-1 chat has at most one at a time).
	if msg.ReplyToMessage != nil {
		key := MessageKey{ChatID: msg.Chat.ID, MessageID: msg.ReplyToMessage.ID}
		if t.deliver(key, msg.Text) {
			return
		}
	}

	t.deliverAny(msg.Chat.ID, msg.Text)
}

// deliver sends text to the waiter registered under key. Returns true if a
// waiter was found.
func (t *Telegram) deliver(key MessageKey, text string) bool {
	value, ok := t.pending.Load(key)
	if !ok {
		return false
	}

	ch := value.(chan string)
	select {
	case ch <- text:
	default:
	}
	return true
}

// deliverAny sends text to any pending waiter in the given chat.
func (t *Telegram) deliverAny(chatID int64, text string) {
	t.pending.Range(func(k, value any) bool {
		key := k.(MessageKey)
		if key.ChatID != chatID {
			return true
		}
		ch := value.(chan string)
		select {
		case ch <- text:
		default:
		}
		return false // stop after delivering to the first match
	})
}

// waitReply waits for a reply to a specific message.
func (t *Telegram) waitReply(
	ctx context.Context,
	messageID int,
) (string, error) {
	key := MessageKey{
		ChatID:    t.chatID,
		MessageID: messageID,
	}

	ch := make(chan string, 1)

	t.pending.Store(key, ch)
	defer t.pending.Delete(key)

	select {
	case reply := <-ch:
		return reply, nil

	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// SendMessageAndWaitReply sends a text message and waits for a reply.
func (t *Telegram) SendMessageAndWaitReply(
	ctx context.Context,
	text string,
) (string, error) {
	msg, err := t.Bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: t.chatID,
		Text:   text,
	})
	if err != nil {
		return "", err
	}

	return t.waitReply(ctx, msg.ID)
}

// SendPhotoAndWaitReply sends an image and waits for a reply.
func (t *Telegram) SendPhotoAndWaitReply(
	ctx context.Context,
	image []byte,
	caption string,
) (string, error) {
	msg, err := t.Bot.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID: t.chatID,
		Photo: &models.InputFileUpload{
			Filename: "screenshot.jpg",
			Data:     bytes.NewReader(image),
		},
		Caption: caption,
	})
	if err != nil {
		return "", err
	}

	fmt.Printf("[tg] sent photo, message id=%d, waiting for reply...\n", msg.ID)

	return t.waitReply(ctx, msg.ID)
}
