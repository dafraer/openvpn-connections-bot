package notifier

import (
	"bufio"
	"context"
	"os"
	"time"

	"go.uber.org/zap"
)

const (
	readingInterval   = time.Second * 10
	openVPNStatusPath = "/var/log/openvpn/status.log"
	endLine           = "ROUTING TABLE"
)

type Notifier struct {
	connections map[Name]*Connection
	ownerID     int64
	msg         chan Message
	logger      *zap.SugaredLogger
}

func New(ownerID int64, msg chan Message) *Notifier {
	m := make(map[Name]*Connection)
	return &Notifier{
		ownerID:     ownerID,
		connections: m,
		msg:         msg,
	}
}

func (n *Notifier) Listen(ctx context.Context) {
	n.logger.Infow("Status Monitor is running")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(readingInterval)
			f, err := os.Open(openVPNStatusPath)
			if err != nil {
				n.logger.Errorw("Error reading from file", "error", err)
				continue
			}
			n.processStatusFile(ctx, f)
			if err := f.Close(); err != nil {
				n.logger.Errorw("Error closing file", "error", err)
			}
		}

	}

}

func (n *Notifier) processStatusFile(ctx context.Context, file *os.File) {
	scanner := bufio.NewScanner(file)

	//Skip beginning of the file
	scanner.Scan()
	scanner.Scan()
	scanner.Scan()
	tmp := make(map[Name]struct{})
	for scanner.Scan() {
		line := scanner.Text()
		if line == endLine {
			//Check if anyone disconnected
			for k, v := range n.connections {
				if _, ok := tmp[k]; !ok {
					n.logger.Infow("Disconnected", "connection", v)
					msg := formatDisconnectedMessage(v)
					n.sendMessage(msg, n.ownerID)
					delete(n.connections, k)
				}
			}
			return
		}
		connection, err := n.parseConnection(line)
		if err != nil {
			n.logger.Errorw("Error parsing connection string", "error", err)
			return
		}

		tmp[connection.Name] = struct{}{}

		//Check if its a new connection
		if _, ok := n.connections[connection.Name]; !ok {
			n.logger.Infow("New connection", "connection", connection)
			//add real address
			realAddr, err := getAddrFromIP(ctx, connection.IP)
			if err != nil {
				n.logger.Errorw("Error calling address API", "error", err)
			}

			connection.Address = realAddr
			n.connections[connection.Name] = connection
			msg := formatConnectedMessage(connection)
			n.sendMessage(msg, n.ownerID)
		}
	}

}
