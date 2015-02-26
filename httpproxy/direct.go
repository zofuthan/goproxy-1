package httpproxy

import (
	"fmt"
	"github.com/golang/glog"
	"net/http"
)

type DirectRequestFilter struct {
	RequestFilter
}

func (f *DirectRequestFilter) HandleRequest(h *Handler, args *http.Header, rw http.ResponseWriter, req *http.Request) (*http.Response, error) {
	if req.Method != "CONNECT" {
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
	} else {
		glog.Infof("%s \"DIRECT %s %s %s\" - -", req.RemoteAddr, req.Method, req.Host, req.Proto)
		response := &http.Response{
			StatusCode:    200,
			ProtoMajor:    1,
			ProtoMinor:    1,
			Header:        http.Header{},
			ContentLength: -1,
		}
		return response, nil
	}
}

func (f *DirectRequestFilter) Filter(req *http.Request) (args *http.Header, err error) {
	return nil, nil
}
