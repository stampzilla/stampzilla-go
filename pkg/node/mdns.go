package node

import (
	"fmt"
	"time"

	"github.com/jonaz/mdns"
)

//func main() {
//ip, port, err := queryMDNS()
//if err != nil {
//logrus.Error(err)
//return
//}

//logrus.Infof("Found %s:%d", ip, port)
//}

func queryMDNS() (string, int, error) {
	entriesCh := make(chan *mdns.ServiceEntry, 1)

	// Start the lookup
	mdns.Lookup("_stampzilla._tcp", entriesCh)
	select {
	case entry := <-entriesCh:
		close(entriesCh)
		return entry.Addr.String(), entry.Port, nil
	case <-time.After(time.Second * 2):
	}

	return "", 0, fmt.Errorf("No stampzilla server was found with auto discovery :(")
}
