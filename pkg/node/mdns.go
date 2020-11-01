package node

import (
	"context"
	"strconv"
	"strings"

	"github.com/micro/mdns"
	"github.com/sirupsen/logrus"
)

//func main() {
//ip, port, err := queryMDNS()
//if err != nil {
//logrus.Error(err)
//return
//}

//logrus.Infof("Found %s:%d", ip, port)
//}

func queryMDNS() (string, string) {
	entriesCh := make(chan *mdns.ServiceEntry)

	logrus.Info("node: running mdns query")
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				mdns.Lookup("_stampzilla._tcp", entriesCh)
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
	logrus.Infof("node: got mdns query response %s:%s", entry.AddrV4.String(), port)
	return entry.AddrV4.String(), port
}
