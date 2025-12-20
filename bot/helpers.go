package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram/bot"
)

func (b *Bot) sendMessage(ctx context.Context, msg string, chatID int64) {
	_, err := b.b.SendMessage(ctx, &tgbotapi.SendMessageParams{ChatID: chatID, Text: "Go away"})
	if err != nil {
		b.logger.Errorw("Error sending message", "error", err, "ChatID", chatID)
	}
}

func (b *Bot) parseConnection(msg string) (*Connection, error) {
	args := strings.Split(msg, ",")
	if len(args) == 5 {
		connection := &Connection{}
		connection.Name = Name(args[0])
		connection.Address = args[1]
		bytesIn, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			b.logger.Errorw("Error parsing bytes recieved", "error", err)
			return nil, err
		}
		connection.BytesRecieved = bytesIn
		bytesOut, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			b.logger.Errorw("Error parsing bytes sent", "error", err)
			return nil, err
		}

		connection.BytesSent = bytesOut
		timeStamp := args[4]

		connectedSince, err := time.Parse("2006-01-02 15:04:05", timeStamp)
		if err != nil {
			b.logger.Errorw("Error parsing bytes sent", "error", err)
			return nil, err
		}
		connection.Since = connectedSince
		return connection, nil
	}
	return nil, fmt.Errorf("error parsing string, not enough args")
}

func formatConnectionMessage(conn *Connection) string {
	newConnMsg := `
	New connection 
	Name: %s 
	From: %s 
	Connected Since: %v
	`
	return fmt.Sprintf(newConnMsg, string(conn.Name), conn.Address, conn.Since)
}
