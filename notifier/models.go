package notifier

type Name string

type Connection struct {
	Name          Name
	IP            string
	Address       string
	BytesRecieved uint64
	BytesSent     uint64
	Since         string
}

type Message struct {
	ChatID int64
	Msg    string
}
