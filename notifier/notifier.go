package notifier

import (
	"bufio"
	"context"
	"os"
	"sync"
	"time"

	"github.com/dafraer/openvpn-connections-bot/tracker"
	"go.uber.org/zap"
)

func New(ownerID int64, msg chan Message, logger *zap.SugaredLogger, tracker *tracker.Tracker, usageFileMutex *sync.Mutex, usageFilePath string) *Notifier {
	m := make(map[Name]*Connection)
	return &Notifier{
		ownerID:        ownerID,
		connections:    m,
		msg:            msg,
		logger:         logger,
		tracker:        tracker,
		usageFileMutex: usageFileMutex,
		usageFilePath:  usageFilePath,
	}
}

func (n *Notifier) Listen(ctx context.Context) {
	n.logger.Infow("Status Monitor is running")
	go n.tracker.Run(ctx)
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
	midLineFlag := false
	conns := make(map[Name]*Connection)
	for scanner.Scan() {
		line := scanner.Text()
		switch line {
		case endLine:
			//Check if anyone disconnected
			n.CheckDisconnected(conns)
			//We check for new ones only when we have full conections list incl virtual adressess
			for _, connection := range conns {
				n.CheckNewConnection(ctx, connection)
			}
			return
		case midLine:
			midLineFlag = true
			scanner.Scan()
			continue
		}

		if midLineFlag {
			name, virtAddr := n.parseVirtAddr(line)
			n.logger.Debugw("Parsed virtAddr", "name", name, "virtual address", virtAddr)
			conns[Name(name)].InternalIP = virtAddr
			continue
		}

		connection, err := n.parseConnection(line)
		if err != nil {
			n.logger.Errorw("Error parsing connection string", "error", err)
			return
		}

		conns[connection.Name] = connection
	}

}

func (n *Notifier) CheckDisconnected(tmp map[Name]*Connection) {
	for k, v := range n.connections {
		if _, ok := tmp[k]; !ok {
			n.logger.Infow("Disconnected", "connection", v)
			n.tracker.ReqAddr <- v.InternalIP
			domains := <-n.tracker.RespAddr
			v.Visited = domains
			msg := formatDisconnectedMessage(v)
			n.sendMessage(msg, n.ownerID)
			if err := n.updateUsage(v.Name, v.BytesRecieved, v.BytesSent); err != nil {
				n.logger.Errorw("Error updating usage file", "error", err)
			}
			delete(n.connections, k)
		}
	}
}

func (n *Notifier) CheckNewConnection(ctx context.Context, connection *Connection) {
	if _, ok := n.connections[connection.Name]; !ok {
		n.logger.Infow("New connection", "connection", connection)
		//add real address
		realAddr, err := getAddrFromIP(ctx, connection.IP)
		if err != nil {
			n.logger.Errorw("Error calling address API", "error", err)
		}

		connection.Address = realAddr
		n.connections[connection.Name] = connection

		n.tracker.NewAddr <- connection.InternalIP
		n.logger.Debugw("Added connection to tracker")

		msg := formatConnectedMessage(connection)
		n.sendMessage(msg, n.ownerID)
	}
}
