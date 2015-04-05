package direct

import (
	"crypto/tls"
	"fmt"
	"github.com/golang/glog"
	"github.com/phuslu/goproxy/httpproxy/filters"
	"io"
	"net"
	"net/http"
	"time"
)

type Filter struct {
	filters.RoundTripFilter
	transport *http.Transport
}

func init() {
	filters.Register("direct", &filters.RegisteredFilter{
		New: NewFilter,
	})
}

func NewFilter() (filters.Filter, error) {
	return &Filter{
		transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}, nil
}

func (p *Filter) FilterName() string {
	return "direct"
}

func (p *Filter) RoundTrip(ctx *filters.Context, req *http.Request) (*filters.Context, *http.Response, error) {
	if req.Method != "CONNECT" {
		req1, err := http.NewRequest(req.Method, req.URL.String(), req.Body)
		if err != nil {
			return ctx, nil, fmt.Errorf("DIRECT RoundTrip %#v error: %#v", req, err)
		}
		req1.Header = req.Header
		res, err := p.transport.RoundTrip(req1)
		if err == nil {
			glog.Infof("%s \"DIRECT %s %s %s\" %d %s", req.RemoteAddr, req.Method, req.URL.String(), req.Proto, res.StatusCode, res.Header.Get("Content-Length"))
		}
		return ctx, res, err
	} else {
		glog.Infof("%s \"DIRECT %s %s %s\" - -", req.RemoteAddr, req.Method, req.Host, req.Proto)
		remote, err := p.transport.Dial("tcp", req.Host)
		if err != nil {
			return ctx, nil, err
		}
		hijacker, ok := ctx.GetResponseWriter().(http.Hijacker)
		if !ok {
			return ctx, nil, fmt.Errorf("http.ResponseWriter(%#v) does not implments Hijacker", ctx.GetResponseWriter())
		}
		local, _, err := hijacker.Hijack()
		local.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		go io.Copy(remote, local)
		io.Copy(local, remote)
		return ctx, nil, nil
	}
}
