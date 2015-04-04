package direct

import (
	"crypto/tls"
	"fmt"
	"github.com/golang/glog"
	"github.com/phuslu/goproxy/httpproxy"
	"github.com/phuslu/goproxy/httpproxy/plugins"
	"net/http"
)

type Plugin struct {
	transport http.RoundTripper
}

func init() {
	plugins.Register("direct", &plugins.RegisteredPlugin{
		New: NewPlugin,
	})
}

func NewPlugin() (plugins.Plugin, error) {
	return &Plugin{
		transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			DisableCompression: true,
		},
	}, nil
}

func (p *Plugin) PluginName() string {
	return "direct"
}

func (p *Plugin) Fetch(ctx *httpproxy.Context, req *http.Request) (*http.Response, error) {
	if req.Method != "CONNECT" {
		req1, err := http.NewRequest(req.Method, req.URL.String(), req.Body)
		if err != nil {
			return nil, fmt.Errorf("DIRECT Fetch %#v error: %#v", req, err)
		}
		req1.Header = req.Header
		res, err := p.transport.RoundTrip(req1)
		if err == nil {
			glog.Infof("%s \"DIRECT %s %s %s\" %d %s", req.RemoteAddr, req.Method, req.URL.String(), req.Proto, res.StatusCode, res.Header.Get("Content-Length"))
		}
		return res, err
	} else {
		glog.Infof("%s \"DIRECT %s %s %s\" - -", req.RemoteAddr, req.Method, req.Host, req.Proto)
		return &http.Response{
			StatusCode:    200,
			ProtoMajor:    1,
			ProtoMinor:    1,
			Header:        http.Header{},
			ContentLength: -1,
		}, nil
	}
}
