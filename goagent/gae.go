package main

import (
	"bufio"
	"bytes"
	"compress/flate"
	"encoding/binary"
	"fmt"
	"github.com/golang/glog"
	"github.com/phuslu/goproxy/httpproxy"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	appspotDomain string = "appspot.com"
	goagentPath   string = "/_gh"
)

type GAERequestFilter struct {
	AppIDs []string
	Schema string
}

func (g *GAERequestFilter) pickAppID() string {
	return g.AppIDs[0]
}

func (g *GAERequestFilter) encodeRequest(req *http.Request) (*http.Request, error) {
	var b bytes.Buffer
	var err error
	w, err := flate.NewWriter(&b, 9)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(w, "%s %s %s\r\n%s\r\n", req.Method, req.URL.String(), req.Proto, req.Header)
	err = w.Flush()
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(&b)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = binary.Write(&buf, binary.BigEndian, int16(len(data)))
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(data)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(&buf, req.Body)
	if err != nil {
		return nil, err
	}
	data, err = ioutil.ReadAll(&buf)
	if err != nil {
		return nil, err
	}
	return http.NewRequest(
		"POST",
		fmt.Sprintf("%s://%s.%s%s", g.Schema, g.pickAppID(), appspotDomain, goagentPath),
		bytes.NewReader(data))
}

func (g *GAERequestFilter) decodeResponse(res *http.Response) (*http.Response, error) {
	if res.StatusCode >= 300 {
		return res, nil
	}
	var length int16
	err := binary.Read(res.Body, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(flate.NewReader(&io.LimitedReader{res.Body, int64(length)}))
	return http.ReadResponse(r, res.Request)
}

func (g *GAERequestFilter) HandleRequest(h *httpproxy.Handler, args *http.Header, rw http.ResponseWriter, req *http.Request) (*http.Response, error) {
	gaeReq, err := g.encodeRequest(req)
	if err != nil {
		rw.WriteHeader(502)
		fmt.Fprintf(rw, "Error: %s\n", err)
		return nil, err
	}
	gaeReq.Header = req.Header
	res, err := h.Net.HttpClientDo(gaeReq)
	if err == nil {
		glog.Infof("%s \"GAE %s %s %s\" %d %s", req.RemoteAddr, req.Method, req.URL.String(), req.Proto, res.StatusCode, res.Header.Get("Content-Length"))
	}
	return res, err
}

func (g *GAERequestFilter) Filter(req *http.Request) (args *http.Header, err error) {
	return nil, nil
}
