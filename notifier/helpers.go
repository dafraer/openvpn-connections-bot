package notifier

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const addrFromIPURL = "http://ip-api.com/json/%s?fields=540689"

func (n *Notifier) sendMessage(msg string, chatID int64) {
	n.msg <- Message{Msg: msg, ChatID: chatID}
}

func (n *Notifier) parseConnection(msg string) (*Connection, error) {
	args := strings.Split(msg, ",")
	if len(args) == 5 {
		connection := &Connection{}

		connection.Name = Name(args[0])
		connection.IP = args[1]

		bytesIn, err := strconv.ParseUint(args[2], 10, 64)
		if err != nil {
			n.logger.Errorw("Error parsing bytes recieved", "error", err)
			return nil, err
		}
		connection.BytesRecieved = bytesIn

		bytesOut, err := strconv.ParseUint(args[3], 10, 64)
		if err != nil {
			n.logger.Errorw("Error parsing bytes sent", "error", err)
			return nil, err
		}
		connection.BytesSent = bytesOut

		connection.Since = args[4]
		return connection, nil
	}
	return nil, fmt.Errorf("error parsing string, not enough args")
}

func formatConnectedMessage(conn *Connection) string {
	newConnMsg := "🟢 *New connection* \n• *Name:* `%s` \n• *IP:* `%s` \n• *From:* `%s` \n• *Since:* `%v`"
	return fmt.Sprintf(newConnMsg, string(conn.Name), conn.IP, conn.Address, conn.Since)
}

func formatDisconnectedMessage(conn *Connection) string {
	disconnected := "🔴 *Disconnected* \n• *Name:* `%s` \n• *IP:* `%s` • *From:* `%s` \n• *Since:* `%v` UTC\\+3 \n• *Recv:* `%v` bytes \n• *Sent:* `%v` bytes \n • *Visited:*\n•`%v`"
	//only send 2 subdomains
	for i, _ := range conn.Visited {
		conn.Visited[i] = strings.TrimSuffix(conn.Visited[i], ".")
		subDomains := strings.Split(conn.Visited[i], ".")
		if len(subDomains) >= 2 {
			conn.Visited[i] = subDomains[len(subDomains)-2] + "." + subDomains[len(subDomains)-1]
		}
	}
	m := make(map[string]struct{})
	for _, v := range conn.Visited {
		m[v] = struct{}{}
	}
	s := make([]string, 0, len(m))
	for k, _ := range m {
		s = append(s, k)
	}
	return fmt.Sprintf(disconnected, string(conn.Name), conn.IP, conn.Address, conn.Since, conn.BytesRecieved, conn.BytesSent, strings.Join(s, "\n`•`"))
}

type AddrResponse struct {
	Status   string `json:"status"`
	Country  string `json:"country"`
	City     string `json:"city"`
	District string `json:"district"`
}

func getAddrFromIP(ctx context.Context, ip string) (string, error) {
	ip = trimPort(ip)
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(addrFromIPURL, ip), http.NoBody)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var respStruct AddrResponse
	if err := json.NewDecoder(resp.Body).Decode(&respStruct); err != nil {
		return "", err
	}
	if respStruct.Status != "success" {
		return "", fmt.Errorf("unsuccessful API call")
	}
	return fmt.Sprintf("%s, %s, %s", respStruct.Country, respStruct.City, respStruct.District), nil
}

func trimPort(ip string) string {
	for i, c := range ip {
		if c == ':' {
			return ip[:i]
		}
	}
	return ip
}

func (n *Notifier) parseVirtAddr(line string) (name string, addr string) {
	elems := strings.Split(line, ",")
	return elems[1], elems[0]
}

type Usage struct {
	Name     string
	Recieved uint64
	Sent     uint64
}

func (n *Notifier) updateUsage(name Name, recievedBytes, sentBytes uint64) error {
	n.usageFileMutex.Lock()
	defer n.usageFileMutex.Unlock()
	f, err := os.Open(n.usageFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	//Parse file
	scanner := bufio.NewScanner(f)
	users := make(map[Name]*Usage)
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), ",")
		if len(s) < 3 {
			return fmt.Errorf("Wrong usage file format")
		}
		recieved, err := strconv.ParseUint(s[1], 10, 64)
		if err != nil {
			return err
		}
		sent, err := strconv.ParseUint(s[2], 10, 64)
		if err != nil {
			return err
		}
		u := Usage{Name: s[0], Recieved: recieved, Sent: sent}
		users[Name(s[0])] = &u
	}

	//update usage
	if users[name] == nil {
		users[name] = &Usage{}
	}
	user := users[name]
	user.Recieved += recievedBytes
	user.Sent += sentBytes

	//Write file
	res := strings.Builder{}
	for _, v := range users {
		res.WriteString(fmt.Sprintf("%s,%v,%v\n", string(v.Name), v.Recieved, v.Sent))
	}
	if err := os.WriteFile(n.usageFilePath, []byte(res.String()), 0644); err != nil {
		return err
	}

	return nil
}
