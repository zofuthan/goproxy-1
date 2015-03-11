package net2

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"time"
)

type Net2 interface {
	NetResolveIPAddr(network, addr string) (*net.IPAddr, error)
	NetDialTimeout(network, address string, timeout time.Duration) (net.Conn, error)
	TlsDialTimeout(network, address string, config *tls.Config, timeout time.Duration) (*tls.Conn, error)
	HttpClientDo(req *http.Request) (*http.Response, error)
	CopyResponseBody(w io.Writer, res *http.Response) (int64, error)
	GetTimeout() time.Duration
	SetTimeout()
	GetAddressAlias(addr string) (alias string)
}
