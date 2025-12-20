// Package bot contains bot's business logic
package bot

import (
	"context"

	"go.uber.org/zap"

	tgbotapi "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Name string

type Bot struct {
	b           *tgbotapi.Bot
	logger      *zap.SugaredLogger
	ownerID     int64
	connections map[Name]*Connection
}

// New creates a new bot
func New(token string, logger *zap.SugaredLogger, userID int64) (*Bot, error) {
	//Create bot using provided dependencies
	m := make(map[Name]*Connection)
	bot := &Bot{logger: logger, ownerID: userID, connections: m}

	//Create telegram bot with a default handler
	b, err := tgbotapi.New(token, tgbotapi.WithDefaultHandler(bot.defaultHandler))
	if err != nil {
		return nil, err
	}
	bot.b = b
	return bot, nil
}

// Run runs the bot using long polling
func (b *Bot) Run(ctx context.Context) {
	b.logger.Infow("Bot is running")
	go b.MonitorStatus(ctx)
	b.b.Start(ctx)
}

// defaultHandler
func (b *Bot) defaultHandler(ctx context.Context, _ *tgbotapi.Bot, update *models.Update) {
	b.logger.Infow("New update recieved", "update", update)
	if update.Message != nil && update.Message.Chat.ID == b.ownerID {
		switch update.Message.Text {
		case "/start":
			b.processStartHandler(ctx, update)
		case "/help":
			b.processHelpHandler(ctx, update)
		}
	}
}

func (b *Bot) processStartHandler(ctx context.Context, update *models.Update) {
	b.sendMessage(ctx, "Go away", update.Message.Chat.ID)
}

func (b *Bot) processHelpHandler(ctx context.Context, update *models.Update) {
	b.sendMessage(ctx, "Help ain't coming", update.Message.Chat.ID)
}
