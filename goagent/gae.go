package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/golang/glog"
	"github.com/phuslu/goproxy/httpproxy"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	appspotDomain string = "appspot.com"
	goagentPath   string = "/_gh/"
)

type GAERequestFilter struct {
	AppIDs []string
	Schema string
}

func (g *GAERequestFilter) pickAppID() string {
	return g.AppIDs[0]
}

func copyRequest(w io.Writer, req *http.Request) error {
	var err error
	_, err = fmt.Fprintf(w, "%s %s %s\r\n", req.Method, req.URL.String(), "HTTP/1.1")
	if err != nil {
		return err
	}
	for key, values := range req.Header {
		for _, value := range values {
			_, err = fmt.Fprintf(w, "%s: %s\r\n", key, value)
			if err != nil {
				return err
			}
		}
	}
	_, err = w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	_, err = io.Copy(w, req.Body)
	if err != nil {
		return err
	}
	return nil
}

func (g *GAERequestFilter) encodeRequest(req *http.Request) (*http.Request, error) {
	var b bytes.Buffer
	var err error
	var gw *gzip.Writer
	if req.TransferEncoding == nil || req.ContentLength < 1*1024*1024 {
		gw = gzip.NewWriter(&b)
		err = copyRequest(gw, req)
	} else {
		err = copyRequest(&b, req)
	}
	if err != nil {
		return nil, err
	}
	if gw != nil {
		gw.Flush()
	}
	url := fmt.Sprintf("%s://%s.%s%s", g.Schema, g.pickAppID(), appspotDomain, goagentPath)
	req1, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return nil, err
	}
	req1.Header.Set("Conntent-Length", strconv.Itoa(b.Len()))
	req1.ContentLength = int64(b.Len())
	return req1, nil
}

func (g *GAERequestFilter) decodeResponse(res *http.Response) (*http.Response, error) {
	if res.StatusCode != 200 {
		return res, nil
	}
	var err error
	var resp *http.Response
	if "gzip" == res.Header.Get("X-Content-Encoding") {
		r, err := gzip.NewReader(res.Body)
		if err != nil {
			return nil, err
		}
		resp, err = http.ReadResponse(bufio.NewReader(r), res.Request)
	} else {
		resp, err = http.ReadResponse(bufio.NewReader(ioutil.NopCloser(res.Body)), res.Request)
	}
	return resp, err
}

func (g *GAERequestFilter) HandleRequest(h *httpproxy.Handler, args *http.Header, rw http.ResponseWriter, req *http.Request) (*http.Response, error) {
	req1, err := g.encodeRequest(req)
	if err != nil {
		rw.WriteHeader(502)
		fmt.Fprintf(rw, "Error: %s\n", err)
		return nil, err
	}
	req1.Header = req.Header
	res, err := h.Net.HttpClientDo(req1)
	if err == nil {
		glog.Infof("%s \"GAE %s %s %s\" %d %s", req.RemoteAddr, req.Method, req.URL.String(), req.Proto, res.StatusCode, res.Header.Get("Content-Length"))
	}
	return g.decodeResponse(res)
}

func (g *GAERequestFilter) Filter(req *http.Request) (args *http.Header, err error) {
	return nil, nil
}
