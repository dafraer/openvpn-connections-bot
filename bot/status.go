package bot

import (
	"bufio"
	"context"
	"os"
	"time"
)

/*
Status file structure (for reference)

	OpenVPN CLIENT LIST
	Updated,2025-12-20 13:12:58
	Common Name,Real Address,Bytes Received,Bytes Sent,Connected Since
	<list>
	ROUTING TABLE
	Virtual Address,Common Name,Real Address,Last Ref
	<list>
	GLOBAL STATS
	Max bcast/mcast queue length,7
	END
*/
const (
	readingInterval   = time.Second * 10
	openVPNStatusPath = "/var/log/openvpn/status.log"
	endLine           = "ROUTING TABLE"
)

type Connection struct {
	Name          Name
	Address       string
	BytesRecieved uint64
	BytesSent     uint64
	Since         time.Time
}

func (b *Bot) MonitorStatus(ctx context.Context) {
	b.logger.Infow("Status Monitor is running")
	for {
		time.Sleep(readingInterval)
		f, err := os.Open(openVPNStatusPath)
		if err != nil {
			panic(err)
		}
		b.processStatusFile(ctx, f)
		if err := f.Close(); err != nil {
			panic(err)
		}
	}
}

func (b *Bot) processStatusFile(ctx context.Context, file *os.File) {
	scanner := bufio.NewScanner(file)

	//Skip beginning of the file
	scanner.Scan()
	scanner.Scan()
	scanner.Scan()
	for scanner.Scan() {
		line := scanner.Text()
		if line == endLine {
			return
		}
		connection, err := b.parseConnection(line)
		if err != nil {
			b.logger.Errorw("Error parsing connection string", "error", err)
		}
		if _, ok := b.connections[connection.Name]; !ok {
			b.connections[connection.Name] = connection
			b.SendNewConnection(ctx, connection)
		}
	}
}

func (b *Bot) SendNewConnection(ctx context.Context, conn *Connection) {
	msg := formatConnectionMessage(conn)
	b.sendMessage(ctx, msg, b.ownerID)
}
