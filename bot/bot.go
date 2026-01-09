// Package bot contains bot's business logic
package bot

import (
	"bufio"
	"context"
	"os"
	"strconv"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/dafraer/openvpn-connections-bot/notifier"
	tgbotapi "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Bot struct {
	b              *tgbotapi.Bot
	logger         *zap.SugaredLogger
	notifier       *notifier.Notifier
	msg            chan notifier.Message
	usageFileMutex *sync.Mutex
	ownerID        int64
	usageFilePath  string
}

// New creates a new bot
func New(token string, logger *zap.SugaredLogger, notifier *notifier.Notifier, msg chan notifier.Message, usageFileMutex *sync.Mutex, ownerID int64, usageFilePath string) (*Bot, error) {
	//Create bot using provided dependencies

	bot := &Bot{usageFilePath: usageFilePath, logger: logger, notifier: notifier, msg: msg, usageFileMutex: usageFileMutex, ownerID: ownerID}

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
		case "/usage":
			b.processUsage(ctx, update)
		}
	}
}

func (b *Bot) processPing(ctx context.Context, update *models.Update) {
	b.sendMessage(ctx, "pong", update.Message.Chat.ID)
}

type Usage struct {
	Name     string
	Recieved uint64
	Sent     uint64
}

func (b *Bot) processUsage(ctx context.Context, update *models.Update) {
	if update.Message != nil && update.Message.Chat.ID != b.ownerID {
		return
	}
	b.usageFileMutex.Lock()

	defer b.usageFileMutex.Unlock()
	f, err := os.Open(b.usageFilePath)
	if err != nil {
		b.logger.Errorw("Error opening file", "error", err)
		return
	}
	defer f.Close()

	//Parse file
	scanner := bufio.NewScanner(f)
	users := make(map[string]*Usage)
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), ",")
		if len(s) < 3 {
			b.logger.Errorw("Wrong usage file format")
			return
		}
		recieved, err := strconv.ParseUint(s[1], 10, 64)
		if err != nil {
			b.logger.Errorw("Error parsing usage file", "error", err)
		}
		sent, err := strconv.ParseUint(s[2], 10, 64)
		if err != nil {
			b.logger.Errorw("Error parsing usage file", "error", err)
		}
		u := Usage{Name: s[0], Recieved: recieved, Sent: sent}
		users[s[0]] = &u
	}
	b.sendMessage(ctx, formatUsageMessage(users), b.ownerID)
}
