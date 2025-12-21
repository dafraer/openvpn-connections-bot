// Package bot contains bot's business logic
package bot

import (
	"context"

	"go.uber.org/zap"

	"github.com/dafraer/openvpn-connections-bot/notifier"
	tgbotapi "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Bot struct {
	b        *tgbotapi.Bot
	logger   *zap.SugaredLogger
	notifier *notifier.Notifier
	msg      chan notifier.Message
}

// New creates a new bot
func New(token string, logger *zap.SugaredLogger, notifier *notifier.Notifier, msg chan notifier.Message) (*Bot, error) {
	//Create bot using provided dependencies

	bot := &Bot{logger: logger, notifier: notifier, msg: msg}

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
	go b.notifier.Listen(ctx)
	go b.b.Start(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case m := <-b.msg:
			b.sendMessage(ctx, m.Msg, m.ChatID)
		}
	}
}

// defaultHandler
func (b *Bot) defaultHandler(ctx context.Context, _ *tgbotapi.Bot, update *models.Update) {
	b.logger.Infow("New update recieved", "update", update)
	if update.Message != nil {
		switch update.Message.Text {
		case "/ping":
			b.processPing(ctx, update)
		}
	}
}

func (b *Bot) processPing(ctx context.Context, update *models.Update) {
	b.sendMessage(ctx, "pong", update.Message.Chat.ID)
}
