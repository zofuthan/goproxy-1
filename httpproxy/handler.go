package httpproxy

import (
	"github.com/golang/glog"
	"github.com/phuslu/goproxy/net2"
	"net"
	"net/http"
)

type Handler struct {
	http.Handler
	Listener        net.Listener
	Net             net2.Net2
	RequestFilters  []RequestFilter
	ResponseFilters []ResponseFilter
}

type RequestFilter interface {
	HandleRequest(*Handler, *http.Header, http.ResponseWriter, *http.Request) (*http.Response, error)
	Filter(req *http.Request) (args *http.Header, err error)
}

type ResponseFilter interface {
	HandleResponse(*Handler, *http.Header, http.ResponseWriter, *http.Response, error) error
	Filter(res *http.Response) (args *http.Header, err error)
}

func (h Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" && !req.URL.IsAbs() {
		if req.TLS != nil {
			req.URL.Scheme = "https"
			if req.Host != "" {
				req.URL.Host = req.Host
			} else {
				req.URL.Host = req.TLS.ServerName
			}
		} else {
			req.URL.Scheme = "http"
			req.URL.Host = req.Host
		}
	}
	for i, reqfilter := range h.RequestFilters {
		args, err := reqfilter.Filter(req)
		if err != nil {
			glog.Infof("ServeHTTP RequestFilter error: %v", err)
		}
		if args != nil || i == len(h.RequestFilters)-1 {
			res, err := reqfilter.HandleRequest(&h, args, rw, req)
			if err != nil {
				glog.Infof("ServeHTTP HandleRequest error: %v", err)
			}
			if res == nil {
				return
			}
			res.Request = req
			for j, resfilter := range h.ResponseFilters {
				if resfilter == nil {
					break
				}
				args, err := resfilter.Filter(res)
				if err != nil {
					glog.Infof("ServeHTTP ResponseFilter error: %v", err)
				}
				if args != nil || j == len(h.ResponseFilters)-1 {
					err := resfilter.HandleResponse(&h, args, rw, res, err)
					if err != nil {
						glog.Infof("ServeHTTP HandleResponse error: %v", err)
					}
					break
				}
			}
			break
		}
	}
}
