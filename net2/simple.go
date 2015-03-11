package net2

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"time"
)

type SimpleNetwork struct {
}

func (sn *SimpleNetwork) NetResolveIPAddr(network, addr string) (*net.IPAddr, error) {
	return net.ResolveIPAddr(network, addr)
}

func (sn *SimpleNetwork) NetDialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout(network, address, timeout)
}

func (sn *SimpleNetwork) TlsDialTimeout(network string, addr string, config *tls.Config, timeout time.Duration) (*tls.Conn, error) {
	return tls.Dial(network, addr, config)
}

func (sn *SimpleNetwork) HttpClientDo(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	return client.Do(req)
}

func (sn *SimpleNetwork) CopyResponseBody(w io.Writer, res *http.Response) (int64, error) {
	return io.Copy(w, res.Body)
}

func (sn *SimpleNetwork) GetTimeout() time.Duration {
	return 8 * time.Second
}

func (sn *SimpleNetwork) SetTimeout() {
}

func (sn *SimpleNetwork) GetAddressAlias(addr string) (alias string) {
	return ""
}
