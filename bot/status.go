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
	IP            string
	Address       string
	BytesRecieved uint64
	BytesSent     uint64
	Since         string
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
	tmp := make(map[Name]struct{})
	for scanner.Scan() {
		line := scanner.Text()
		if line == endLine {
			return
		}
		connection, err := b.parseConnection(line)
		if err != nil {
			b.logger.Errorw("Error parsing connection string", "error", err)
			return
		}

		tmp[connection.Name] = struct{}{}

		//Check if its a new connection
		if _, ok := b.connections[connection.Name]; !ok {
			//add real address
			realAddr, err := getAddrFromIP(ctx, connection.IP)
			if err != nil {
				b.logger.Errorw("Error calling address API", "error", err)
			}

			connection.Address = realAddr
			b.mutex.Lock()
			b.connections[connection.Name] = connection
			b.mutex.Unlock()
			b.SendNewConnection(ctx, connection)
		}
	}
	//Check if anyone disconnected
	for k, v := range b.connections {
		if _, ok := tmp[k]; !ok {
			msg := formatDisconnected(v)
			b.sendMessage(ctx, msg, b.ownerID)
			delete(b.connections, k)
		}
	}
}

func (b *Bot) SendNewConnection(ctx context.Context, conn *Connection) {
	msg := formatNewConnectionMessage(conn)
	b.sendMessage(ctx, msg, b.ownerID)
}
