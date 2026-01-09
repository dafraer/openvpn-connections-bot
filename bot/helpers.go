package bot

import (
	"context"
	"fmt"
	"slices"
	"strings"

	tgbotapi "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (b *Bot) sendMessage(ctx context.Context, msg string, chatID int64) {
	_, err := b.b.SendMessage(ctx, &tgbotapi.SendMessageParams{ChatID: chatID, Text: msg, ParseMode: models.ParseModeMarkdown})
	if err != nil {
		b.logger.Errorw("Error sending message", "error", err, "ChatID", chatID)
	}
}

type user struct {
	name  string
	total uint64
}

func formatUsageMessage(users map[string]*Usage) string {
	builder := strings.Builder{}
	builder.WriteString("Top users:\n")
	usersSlice := make([]user, 0, len(users))
	for _, v := range users {
		u := user{total: v.Recieved + v.Sent, name: v.Name}
		usersSlice = append(usersSlice, u)
	}
	cmpFunc := func(a, b user) int {
		if a.total < b.total {
			return 1
		}
		if a.total > b.total {
			return -1
		}
		return 0
	}
	slices.SortFunc(usersSlice, cmpFunc)
	for i, v := range usersSlice {
		builder.WriteString(fmt.Sprintf("%d\\. Name:%s, Total data used: %s\n", i+1, v.name, formatBytes(v.total)))
	}
	return builder.String()
}

func formatBytes(b uint64) string {
	res := ""
	switch {
	case b < 1000:
		res = fmt.Sprintf("%v B", b)
	case b >= 1000 && b < 1000_000:
		res = fmt.Sprintf("%v KB", b/1000)
	case b >= 1000_000 && b < 1000_000_000:
		res = fmt.Sprintf("%v MB", b/1000_000)
	case b >= 1000_000_000:
		res = fmt.Sprintf("%v GB", b/1000_000_000)
	}
	return res
}
