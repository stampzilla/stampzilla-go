package node

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/mdns"
	"github.com/sirupsen/logrus"
)

// func main() {
// ip, port, err := queryMDNS()
// if err != nil {
// logrus.Error(err)
// return
//}

// logrus.Infof("Found %s:%d", ip, port)
//}

func queryMDNS() (string, string, string) {
	entriesCh := make(chan *mdns.ServiceEntry)

	logrus.Info("node: running mdns query")
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// mdns.Lookup("_stampzilla._tcp", entriesCh)
				params := mdns.DefaultParams("_stampzilla._tcp")
				params.Entries = entriesCh
				params.Timeout = time.Second * 5
				params.DisableIPv6 = true
				err := mdns.Query(params)
				if err != nil {
					logrus.Error(err)
				}
			}
		}
	}()

	var entry *mdns.ServiceEntry
	for {
		entry = <-entriesCh
		if strings.Contains(entry.Name, "_stampzilla._tcp") { // Ignore answers that are not what we are looking for
			break
		}
	}
	cancel()
	port := strconv.Itoa(entry.Port)
	tlsPort := ""
	for _, v := range entry.InfoFields {
		tmp := strings.SplitN(v, "=", 2)
		if len(tmp) != 2 {
			continue
		}
		if tmp[0] == "tlsPort" {
			tlsPort = tmp[1]
		}
	}
	logrus.Infof("node: got mdns query response %s:%s (tlsPort=%s)", entry.AddrV4.String(), port, tlsPort)
	return entry.AddrV4.String(), port, tlsPort
}
