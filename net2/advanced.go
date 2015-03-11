package net2

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"
)

type AdvancedNetwork struct {
	dnsCache   map[string]*net.IPAddr
	dnsCacheMu sync.Mutex
}

func NewAdvancedNetwork() *AdvancedNetwork {
	return &AdvancedNetwork{
		dnsCache: map[string]*net.IPAddr{},
	}
}

func (an *AdvancedNetwork) NetResolveIPAddr(network, addr string) (*net.IPAddr, error) {
	return net.ResolveIPAddr(network, addr)
}

func (an *AdvancedNetwork) NetDialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout(network, address, timeout)
}

func (an *AdvancedNetwork) TlsDialTimeout(network string, addr string, config *tls.Config, timeout time.Duration) (*tls.Conn, error) {
	return tls.Dial(network, addr, config)
}

func (an *AdvancedNetwork) HttpClientDo(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	return client.Do(req)
}

func (an *AdvancedNetwork) GetTimeout() time.Duration {
	return 8 * time.Second
}

func (s *AdvancedNetwork) SetTimeout() {
}

func (s *AdvancedNetwork) GetAddressAlias(addr string) (alias string) {
	return ""
}
