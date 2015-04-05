package gae

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/golang/glog"
	"github.com/phuslu/goproxy/httpproxy/filters"
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

type Filter struct {
	AppIDs []string
	Scheme string
}

func (f *Filter) pickAppID() string {
	return f.AppIDs[0]
}

func (f *Filter) encodeRequest(req *http.Request) (*http.Request, error) {
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
		Scheme: f.Scheme,
		Host:   fmt.Sprintf("%s.%s", f.pickAppID(), appspotDomain),
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

func (f *Filter) decodeResponse(res *http.Response) (*http.Response, error) {
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

func (f *Filter) Fetch(ctx *filters.Context, req *http.Request) (*filters.Context, *http.Response, error) {
	req1, err := f.encodeRequest(req)
	if err != nil {
		return ctx, nil, fmt.Errorf("GAE encodeRequest: %s", err.Error())
	}
	req1.Header = req.Header
	res, err := ctx.GetTransport().RoundTrip(req1)
	if err != nil {
		return ctx, nil, err
	} else {
		glog.Infof("%s \"GAE %s %s %s\" %d %s", req.RemoteAddr, req.Method, req.URL.String(), req.Proto, res.StatusCode, res.Header.Get("Content-Length"))
	}
	resp, err := f.decodeResponse(res)
	return ctx, resp, err
}
