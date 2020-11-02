package helpers

import (
	"fmt"
	"net"

	"net/url"

	"github.com/sirupsen/logrus"
)

var privateIPBlocks []*net.IPNet

// Credits to https://stackoverflow.com/questions/41240761/check-if-ip-address-is-in-private-network-space/50825191#50825191
func init() {
	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927 link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Errorf("parse error on %q: %v", cidr, err))
		}
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

func IsPrivateIP(ipStr string) bool {
	u, err := url.Parse("http://" + ipStr)
	if err != nil {
		return false
	}

	ip := net.ParseIP(u.Hostname())

	defer func() {
		if !res {
			logrus.Warnf("Rejected request from %s", ip.String())
		}
	}()

	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}
