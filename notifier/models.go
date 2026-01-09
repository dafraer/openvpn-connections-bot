package notifier

import (
	"sync"
	"time"

	"github.com/dafraer/openvpn-connections-bot/tracker"
	"go.uber.org/zap"
)

const (
	readingInterval   = time.Second * 10
	openVPNStatusPath = "/var/log/openvpn/status.log"
	midLine           = "ROUTING TABLE"
	endLine           = "GLOBAL STATS"
)

type Name string

type Notifier struct {
	connections    map[Name]*Connection
	tracker        *tracker.Tracker
	ownerID        int64
	msg            chan Message
	logger         *zap.SugaredLogger
	usageFileMutex *sync.Mutex
	usageFilePath  string
}

type Connection struct {
	Name          Name
	IP            string
	Address       string
	BytesRecieved uint64
	BytesSent     uint64
	Since         string
	Visited       []string
	InternalIP    string
}

type Message struct {
	ChatID int64
	Msg    string
}
