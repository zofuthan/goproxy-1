package httpproxy

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type MockRequestFilter struct {
}

func (f *MockRequestFilter) HandleRequest(h *Handler, args *FilterArgs, rw http.ResponseWriter, req *http.Request) (*http.Response, error) {
	status, err := args.GetString("status")
	if err != nil {
		return nil, err
	}
	header, err := args.GetHeader("header")
	if err != nil {
		return nil, err
	}
	content, err := args.GetString("content")
	if err != nil {
		return nil, err
	}
	return &http.Response{
		Status:        status,
		Header:        *header,
		Body:          ioutil.NopCloser(bytes.NewBufferString(content)),
		ContentLength: int64(len(content)),
		Request:       req,
		Close:         true,
	}, nil
}
