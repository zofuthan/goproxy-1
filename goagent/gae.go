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
	"net/url"
)

const (
	appspotDomain string = "appspot.com"
	goagentPath   string = "/_gh/"
)

var reqWriteExcludeHeader = map[string]bool{
	"Vary":                true,
	"Via":                 true,
	"X-Forwarded-For":     true,
	"Proxy-Authorization": true,
	"Proxy-Connection":    true,
	"Upgrade":             true,
	"X-Chrome-Variations": true,
	"Connection":          true,
	"Cache-Control":       true,
}

type GAERequestFilter struct {
	AppIDs []string
	Scheme string
}

func (g *GAERequestFilter) pickAppID() string {
	return g.AppIDs[0]
}

func (g *GAERequestFilter) encodeRequest(req *http.Request) (*http.Request, error) {
	var err error
	var b bytes.Buffer
	var w io.Writer
	var gw *gzip.Writer

	if req.TransferEncoding == nil || req.ContentLength < 1*1024*1024 {
		gw = gzip.NewWriter(&b)
		w = gw
	} else {
		w = &b
	}

	_, err = fmt.Fprintf(w, "%s %s %s\r\n", req.Method, req.URL.String(), "HTTP/1.1")
	if err != nil {
		return nil, err
	}
	err = req.Header.WriteSubset(w, reqWriteExcludeHeader)
	if err != nil {
		return nil, err
	}
	_, err = io.WriteString(w, "\r\n")
	if err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if gw != nil {
		_, err = io.Copy(w, req.Body)
		if err != nil {
			return nil, err
		}
		err = gw.Flush()
		if err != nil {
			return nil, err
		}
		bodyReader = &b
	} else {
		bodyReader = io.MultiReader(&b, req.Body)
	}

	u := &url.URL{
		Scheme: g.Scheme,
		Host:   fmt.Sprintf("%s.%s", g.pickAppID(), appspotDomain),
		Path:   goagentPath,
	}
	if gw != nil {
		u.Path += "gzip"
	}
	req1 := &http.Request{
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Method:        "POST",
		URL:           u,
		Host:          u.Host,
		ContentLength: int64(b.Len()),
		Body:          ioutil.NopCloser(bodyReader),
		Header: http.Header{
			"User-Agent": []string{"B"},
		},
	}
	if gw != nil {
		req1.Header.Set("X-Content-Encoding", "gzip")
	}
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

func (g *GAERequestFilter) HandleRequest(h *httpproxy.Handler, args *httpproxy.FilterArgs, rw http.ResponseWriter, req *http.Request) (*http.Response, error) {
	req1, err := g.encodeRequest(req)
	if err != nil {
		rw.WriteHeader(502)
		fmt.Fprintf(rw, "Error: %s\n", err)
		return nil, err
	}
	req1.Header = req.Header
	res, err := h.Transport.RoundTrip(req1)
	if err != nil {
		return nil, err
	} else {
		glog.Infof("%s \"GAE %s %s %s\" %d %s", req.RemoteAddr, req.Method, req.URL.String(), req.Proto, res.StatusCode, res.Header.Get("Content-Length"))
	}
	return g.decodeResponse(res)
}

func (g *GAERequestFilter) Filter(req *http.Request) (args *httpproxy.FilterArgs, err error) {
	return nil, nil
}
