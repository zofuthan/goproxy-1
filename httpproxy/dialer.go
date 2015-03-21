package httpproxy

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"
)

var googleIP = "58.176.217.109"

type Dialer struct {
	Timeout   time.Duration
	Deadline  time.Time
	LocalAddr net.Addr
	DualStack bool
	KeepAlive time.Duration
}

func (d *Dialer) Dial(network, addr string) (net.Conn, error) {
	d1 := &net.Dialer{
		Timeout:   d.Timeout,
		Deadline:  d.Deadline,
		LocalAddr: d.LocalAddr,
		DualStack: d.DualStack,
		KeepAlive: d.KeepAlive,
	}
	if network == "tcp" || network == "tcp4" {
		host, port, err := net.SplitHostPort(addr)
		if err == nil {
			if strings.HasSuffix(host, ".appspot.com") {
				addr1 := fmt.Sprintf("%s:%s", googleIP, port)
				return d1.Dial(network, addr1)
			}
		}
	}
	return d1.Dial(network, addr)
}

func (d *Dialer) DialTLS(network, addr string) (net.Conn, error) {
	d1 := &net.Dialer{
		Timeout:   d.Timeout,
		Deadline:  d.Deadline,
		LocalAddr: d.LocalAddr,
		DualStack: d.DualStack,
		KeepAlive: d.KeepAlive,
	}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	if network == "tcp" || network == "tcp4" {
		host, port, err := net.SplitHostPort(addr)
		if err == nil {
			if strings.HasSuffix(host, ".appspot.com") {
				addr1 := fmt.Sprintf("%s:%s", googleIP, port)
				return tls.DialWithDialer(d1, network, addr1, tlsConfig)
			}
		}
	}
	return tls.DialWithDialer(d1, network, addr, tlsConfig)
}
