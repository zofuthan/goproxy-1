package mock

import (
	"bytes"
	"github.com/golang/glog"
	"github.com/phuslu/goproxy/httpproxy/filters"
	"io/ioutil"
	"net/http"
)

type Filter struct {
	filters.FetchFilter
}

func init() {
	filters.Register("mock", &filters.RegisteredFilter{
		New: NewFilter,
	})
}

func NewFilter() (filters.Filter, error) {
	return &Filter{}, nil
}

func (f *Filter) FilterName() string {
	return "mock"
}

func (f *Filter) Fetch(ctx *filters.Context, req *http.Request) (*http.Response, error) {
	statusCode, err := ctx.GetInt("StatusCode")
	if err != nil {
		return nil, err
	}
	header, err := ctx.GetHeader("Header")
	if err != nil {
		return nil, err
	}
	body, err := ctx.GetString("Body")
	if err != nil {
		body = ""
	}
	resp := &http.Response{
		StatusCode:    statusCode,
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        *header,
		ContentLength: int64(len(body)),
		Body:          ioutil.NopCloser(bytes.NewBufferString(body)),
		Request:       req,
	}
	glog.Infof("%s \"MOCK %s %s %s\" %d %s", req.RemoteAddr, req.Method, req.URL.String(), req.Proto, resp.StatusCode, resp.Header.Get("Content-Length"))
	return resp, nil
}
