package httpproxy

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"
)

var googleIPList = []string{
	"58.176.217.88",
	"58.176.217.99",
	"58.176.217.104",
	"58.176.217.109",
	"58.176.217.114",
}

type timeoutError struct{}

func (e *timeoutError) Error() string   { return "i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

var errTimeout error = &timeoutError{}

type Dialer struct {
	Timeout   time.Duration
	Deadline  time.Time
	LocalAddr net.Addr
	DualStack bool
	KeepAlive time.Duration
	TLSConfig *tls.Config
}

func (d *Dialer) deadline() time.Time {
	if d.Timeout == 0 {
		return d.Deadline
	}
	timeoutDeadline := time.Now().Add(d.Timeout)
	if d.Deadline.IsZero() || timeoutDeadline.Before(d.Deadline) {
		return timeoutDeadline
	} else {
		return d.Deadline
	}
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
				//TODO: net.IPAddr???
				ipaddrs := make([]net.Addr, 0)
				for _, ip := range googleIPList {
					ipaddr, err := net.ResolveIPAddr("ip", fmt.Sprintf("%s:%s", ip, port))
					if err != nil {
						return nil, err
					}
					ipaddrs = append(ipaddrs, ipaddr)
				}
				return d.dialMulti("ip", ipaddrs)
			}
		}
	}
	return d1.Dial(network, addr)
}

func (d *Dialer) dialMulti(network string, addrs []net.Addr) (net.Conn, error) {
	d1 := &net.Dialer{
		Timeout:   d.Timeout,
		Deadline:  d.Deadline,
		LocalAddr: d.LocalAddr,
		DualStack: d.DualStack,
		KeepAlive: d.KeepAlive,
	}
	type racer struct {
		net.Conn
		error
	}
	lane := make(chan racer, len(addrs))
	for _, ra := range addrs {
		go func(ra net.Addr) {
			c, err := d1.Dial(network, ra.String())
			lane <- racer{c, err}
		}(ra)
	}
	lastErr := errTimeout
	nracers := len(addrs)
	for nracers > 0 {
		racer := <-lane
		if racer.error == nil {
			go func(n int) {
				for i := 0; i < n; i++ {
					racer := <-lane
					if racer.error == nil {
						racer.Close()
					}
				}
			}(nracers - 1)
			return racer.Conn, nil
		}
		lastErr = racer.error
		nracers--
	}
	return nil, lastErr
}

func (d *Dialer) DialTLS(network, addr string) (net.Conn, error) {
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
				//TODO: net.IPAddr???
				ipaddrs := make([]net.Addr, 0)
				for _, ip := range googleIPList {
					ipaddr, err := net.ResolveIPAddr("ip", fmt.Sprintf("%s:%s", ip, port))
					if err != nil {
						return nil, err
					}
					ipaddrs = append(ipaddrs, ipaddr)
				}
				return d.dialMultiTLS("ip", ipaddrs)
			}
		}
	}
	return tls.DialWithDialer(d1, network, addr, d.TLSConfig)
}

func (d *Dialer) dialMultiTLS(network string, addrs []net.Addr) (net.Conn, error) {
	d1 := &net.Dialer{
		Timeout:   d.Timeout,
		Deadline:  d.Deadline,
		LocalAddr: d.LocalAddr,
		DualStack: d.DualStack,
		KeepAlive: d.KeepAlive,
	}
	type racer struct {
		net.Conn
		error
	}
	lane := make(chan racer, len(addrs))
	for _, ra := range addrs {
		go func(ra net.Addr) {
			//TODO: network == "ip4" ??
			addr := ra.String()
			conn, err := d1.Dial(network, addr)
			if err != nil {
				lane <- racer{conn, err}
				return
			}
			colonPos := strings.LastIndex(addr, ":")
			if colonPos == -1 {
				colonPos = len(addr)
			}
			hostname := addr[:colonPos]
			config := d.TLSConfig
			if config == nil {
				config = &tls.Config{
					InsecureSkipVerify: true,
				}
			}
			if config.ServerName == "" {
				c := *config
				c.ServerName = hostname
				config = &c
			}
			tlsConn := tls.Client(conn, config)
			var errChannel chan error
			if d1.Timeout == 0 {
				err = tlsConn.Handshake()
			} else {
				errChannel = make(chan error, 2)
				time.AfterFunc(d1.Timeout, func() {
					errChannel <- &timeoutError{}
				})
				go func() {
					errChannel <- tlsConn.Handshake()
				}()
				err = <-errChannel
			}
			if err != nil {
				conn.Close()
			}
			lane <- racer{tlsConn, err}
		}(ra)
	}
	lastErr := errTimeout
	nracers := len(addrs)
	for nracers > 0 {
		racer := <-lane
		if racer.error == nil {
			go func(n int) {
				for i := 0; i < n; i++ {
					racer := <-lane
					if racer.error == nil {
						racer.Close()
					}
				}
			}(nracers - 1)
			return racer.Conn, nil
		}
		lastErr = racer.error
		nracers--
	}
	return nil, lastErr
}
