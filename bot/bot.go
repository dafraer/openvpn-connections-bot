// Package bot contains bot's business logic
package bot

import (
	"context"
	"fmt"
	"sync"

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
	mutex       *sync.RWMutex
}

// New creates a new bot
func New(token string, logger *zap.SugaredLogger, userID int64) (*Bot, error) {
	//Create bot using provided dependencies
	m := make(map[Name]*Connection)
	bot := &Bot{logger: logger, ownerID: userID, mutex: &sync.RWMutex{}, connections: m}

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
			b.processHelpHanler(ctx, update)
		case "/connections":
			b.processConnectionsHandler(ctx, update)
		}
	}
}

func (b *Bot) processStartHandler(ctx context.Context, update *models.Update) {
	b.sendMessage(ctx, "Go away", update.Message.Chat.ID)
}

func (b *Bot) processHelpHanler(ctx context.Context, update *models.Update) {
	b.sendMessage(ctx, "Help ain't coming", update.Message.Chat.ID)
}

func (b *Bot) processConnectionsHandler(ctx context.Context, update *models.Update) {
	msg, err := b.formatConnectionsJSON()
	if err != nil {
		b.logger.Errorw("Error formatting JSON", "error", err)
		b.sendMessage(ctx, "Sorry, error occured try again", b.ownerID)
	}
	stringMessage := fmt.Sprintf("```%v```", msg)
	b.sendMessage(ctx, stringMessage, b.ownerID)
}
