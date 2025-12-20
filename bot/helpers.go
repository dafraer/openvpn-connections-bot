package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram/bot"
)

const addrFromIPURL = "http://ip-api.com/json/%s?fields=540689"

func (b *Bot) sendMessage(ctx context.Context, msg string, chatID int64) {
	_, err := b.b.SendMessage(ctx, &tgbotapi.SendMessageParams{ChatID: chatID, Text: msg})
	if err != nil {
		b.logger.Errorw("Error sending message", "error", err, "ChatID", chatID)
	}
}

func (b *Bot) parseConnection(msg string) (*Connection, error) {
	args := strings.Split(msg, ",")
	if len(args) == 5 {
		connection := &Connection{}
		connection.Name = Name(args[0])
		connection.IP = args[1]
		bytesIn, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			b.logger.Errorw("Error parsing bytes recieved", "error", err)
			return nil, err
		}
		connection.BytesRecieved = bytesIn
		bytesOut, err := strconv.ParseUint(args[3], 10, 64)
		if err != nil {
			b.logger.Errorw("Error parsing bytes sent", "error", err)
			return nil, err
		}

		connection.BytesSent = bytesOut
		connection.Since = args[4]
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

func (b *Bot) formatConnectionsJSON() ([]byte, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	table, err := json.Marshal(b.connections)
	if err != nil {
		return nil, err
	}
	return table, nil
}

type AddrResponse struct {
	Status   string `json:"status"`
	Country  string `json:"country"`
	City     string `json:"city"`
	District string `json:"district"`
}

func getAddrFromIP(ctx context.Context, ip string) (string, error) {
	resp, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(addrFromIPURL, ip), http.NoBody)
	if err != nil {
		return "", err
	}
	var respStruct AddrResponse
	if err := json.NewDecoder(resp.Body).Decode(&respStruct); err != nil {
		return "", err
	}
	if respStruct.Status != "success" {
		return "", fmt.Errorf("unsuccessful API call")
	}
	return fmt.Sprintf("%s, %s, %s", respStruct.Country, respStruct.City, respStruct.District), nil
}
