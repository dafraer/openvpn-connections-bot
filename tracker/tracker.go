package tracker

import (
	"context"
	"net"
	"net/netip"
	"time"

	"github.com/ti-mo/conntrack"
	"go.uber.org/zap"
)

type virtAddr string

type Tracker struct {
	domains  map[virtAddr]map[string]struct{}
	logger   *zap.SugaredLogger
	NewAddr  chan string
	ReqAddr  chan string
	RespAddr chan []string
}

func New(newAddr chan string, reqAddr chan string, respAddr chan []string, logger *zap.SugaredLogger) *Tracker {
	m := make(map[virtAddr]map[string]struct{})
	return &Tracker{
		domains:  m,
		ReqAddr:  reqAddr,
		NewAddr:  newAddr,
		RespAddr: respAddr,
		logger:   logger,
	}
}

func (t *Tracker) Run(ctx context.Context) {
	t.logger.Debugw("tracker is running")
	for {
		select {
		case <-ctx.Done():
			return
		case addr := <-t.NewAddr:
			t.domains[virtAddr(addr)] = make(map[string]struct{})
			t.logger.Debugw("Added new virtAddr", "address", addr)
		case addr := <-t.ReqAddr:
			resp := make([]string, 0, len(t.domains[virtAddr(addr)]))
			for k, _ := range t.domains[virtAddr(addr)] {
				resp = append(resp, k)
			}

			t.RespAddr <- resp
			delete(t.domains, virtAddr(addr))
		default:
			time.Sleep(time.Second)
			t.Track()
		}
	}
}

func (t *Tracker) Track() {
	for virtAddr, domainList := range t.domains {
		d := t.GetVisited(string(virtAddr))
		for _, v := range d {
			domainList[v] = struct{}{}
		}
	}
}

// GetVisited Vibe coded idk what it really does
// returns empty slice in case of an error
func (t *Tracker) GetVisited(srcIP string) []string {
	t.logger.Debugw("GetVisited called")
	src, err := netip.ParseAddr(srcIP)
	if err != nil {
		t.logger.Errorw("Error parsing ip address", "error", err)
		return []string{}
	}

	c, err := conntrack.Dial(nil)
	if err != nil {
		t.logger.Errorw("Error dialing conntrack", "error", err)
	}
	defer c.Close()

	// Dumps current conntrack entries.
	flows, err := c.Dump(nil)
	if err != nil {
		t.logger.Errorw("Error dumping entries", "error", err)
	}

	dsts := make([]string, 0, 128)

	for _, f := range flows {
		orig := f.TupleOrig.IP

		// Skip incomplete entries, just in case.
		if !orig.SourceAddress.IsValid() || !orig.DestinationAddress.IsValid() {
			continue
		}

		if orig.SourceAddress != src {
			continue
		}

		dst := orig.DestinationAddress.String()
		domains, _ := net.LookupAddr(dst)
		dsts = append(dsts, domains...)
	}
	return dsts
}
