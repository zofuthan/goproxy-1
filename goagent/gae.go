package main

import (
	"fmt"
	//"github.com/golang/glog"
	"github.com/phuslu/goproxy/httpproxy"
	"net/http"
)

type GAERequestFilter struct {
}

func (f *GAERequestFilter) HandleRequest(h *httpproxy.Handler, args *http.Header, rw http.ResponseWriter, req *http.Request) (*http.Response, error) {
	if !req.URL.IsAbs() {
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
	newReq, err := http.NewRequest(req.Method, req.URL.String(), req.Body)
	if err != nil {
		rw.WriteHeader(502)
		fmt.Fprintf(rw, "Error: %s\n", err)
		return nil, err
	}
	newReq.Header = req.Header
	res, err := h.Net.HttpClientDo(newReq)
	return res, err
}

func (f *GAERequestFilter) Filter(req *http.Request) (args *http.Header, err error) {
	return nil, nil
}
