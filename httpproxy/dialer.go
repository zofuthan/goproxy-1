package httpproxy

import (
	"crypto/tls"
	"net"
	"time"
)

type timeoutError struct{}

func (e *timeoutError) Error() string   { return "i/o timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

var (
	errTimeout      error    = &timeoutError{}
	defaultResolver Resolver = NewResolver(nil)
)

type Dialer struct {
	Timeout     time.Duration
	Deadline    time.Time
	LocalAddr   net.Addr
	DualStack   bool
	KeepAlive   time.Duration
	TLSConfig   *tls.Config
	DNSResolver Resolver
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
	resolver := d.DNSResolver
	if resolver == nil {
		resolver = defaultResolver
	}
	if network == "tcp" || network == "tcp4" {
		host, port, err := net.SplitHostPort(addr)
		if err == nil {
			addrs, err := resolver.LookupHost(host)
			if err == nil {
				ipaddrs := make([]string, 0)
				for _, addr := range addrs {
					ipaddrs = append(ipaddrs, net.JoinHostPort(addr, port))
				}
				return d.dialMultiTLS(network, ipaddrs)
			}
		}
	}
	return d1.Dial(network, addr)
}

func (d *Dialer) dialMulti(network string, addrs []string) (net.Conn, error) {
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
	for _, raddr := range addrs {
		go func(raddr string) {
			c, err := d1.Dial(network, raddr)
			lane <- racer{c, err}
		}(raddr)
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
	resolver := d.DNSResolver
	if resolver == nil {
		resolver = defaultResolver
	}
	if network == "tcp" || network == "tcp4" {
		host, port, err := net.SplitHostPort(addr)
		if err == nil {
			addrs, err := resolver.LookupHost(host)
			if err == nil {
				ipaddrs := make([]string, 0)
				for _, addr := range addrs {
					ipaddrs = append(ipaddrs, net.JoinHostPort(addr, port))
				}
				return d.dialMultiTLS(network, ipaddrs)
			}
		}
	}
	return tls.DialWithDialer(d1, network, addr, d.TLSConfig)
}

func (d *Dialer) dialMultiTLS(network string, addrs []string) (net.Conn, error) {
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
	for _, raddr := range addrs {
		go func(raddr string) {
			config := d.TLSConfig
			if config == nil {
				config = &tls.Config{
					InsecureSkipVerify: true,
				}
			}
			if config.ServerName == "" {
				c := *config
				c.ServerName = "www.gov.cn"
				config = &c
			}
			conn, err := tls.DialWithDialer(d1, network, raddr, config)
			lane <- racer{conn, err}
		}(raddr)
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
