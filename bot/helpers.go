package bot

import (
	"context"

	tgbotapi "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (b *Bot) sendMessage(ctx context.Context, msg string, chatID int64) {
	_, err := b.b.SendMessage(ctx, &tgbotapi.SendMessageParams{ChatID: chatID, Text: msg, ParseMode: models.ParseModeMarkdown})
	if err != nil {
		b.logger.Errorw("Error sending message", "error", err, "ChatID", chatID)
	}
}
