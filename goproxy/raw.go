package goproxy

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	"io"
	"net/http"
)

type RawResponseFilter struct {
	ResponseFilter
}

func (f *RawResponseFilter) HandleResponse(h *Handler, args *http.Header, rw http.ResponseWriter, req *http.Request, res *http.Response, resError error) error {
	if req.Method != "CONNECT" {
		if resError != nil {
			rw.WriteHeader(502)
			fmt.Fprintf(rw, "Error: %s\n", resError)
			return resError
		}
		glog.Infof("%s \"DIRECT %s %s %s\" %d %s", req.RemoteAddr, req.Method, req.URL.String(), req.Proto, res.StatusCode, res.Header.Get("Content-Length"))
		rw.WriteHeader(res.StatusCode)
		for key, values := range res.Header {
			for _, value := range values {
				rw.Header().Add(key, value)
			}
		}
		io.Copy(rw, res.Body)
	} else {
		if resError != nil {
			rw.WriteHeader(502)
			fmt.Fprintf(rw, "Error: %s\n", resError)
			glog.Infof("NetDialTimeout %s failed %s", req.Host, resError)
			return resError
		}
		remoteConn, err := h.Net.NetDialTimeout("tcp", req.Host, h.Net.GetTimeout())
		if err != nil {
			return err
		}
		hijacker, ok := rw.(http.Hijacker)
		if !ok {
			resError = errors.New("http.ResponseWriter does not implments Hijacker")
			rw.WriteHeader(502)
			fmt.Fprintf(rw, "Error: %s\n", resError)
			return resError
		}
		localConn, _, err := hijacker.Hijack()
		localConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		go io.Copy(remoteConn, localConn)
		io.Copy(localConn, remoteConn)
	}
	return nil
}

func (f *RawResponseFilter) Filter(req *http.Request, res *http.Response) (args *http.Header, err error) {
	return nil, nil
}
