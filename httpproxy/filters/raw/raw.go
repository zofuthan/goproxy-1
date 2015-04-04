package httpproxy

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

func (f *RawResponseFilter) HandleResponse(h *Handler, args *FilterArgs, rw http.ResponseWriter, res *http.Response, resError error) error {
	if res.Request.Method != "CONNECT" {
		if resError != nil {
			rw.WriteHeader(502)
			fmt.Fprintf(rw, "Error: %s\n", resError)
			return resError
		}
		for key, values := range res.Header {
			for _, value := range values {
				rw.Header().Add(key, value)
			}
		}
		rw.WriteHeader(res.StatusCode)
		io.Copy(rw, res.Body)
	} else {
		if resError != nil {
			rw.WriteHeader(502)
			fmt.Fprintf(rw, "Error: %s\n", resError)
			glog.Infof("NetDialTimeout %s failed %s", res.Request.Host, resError)
			return resError
		}
		remoteConn, err := h.Transport.Dial("tcp", res.Request.Host)
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

func (f *RawResponseFilter) Filter(res *http.Response) (args *FilterArgs, err error) {
	return nil, nil
}
