package httpproxy

import (
// "fmt"
// "github.com/golang/glog"
// "net/http"
)

type MockRequestFilter struct {
	RequestFilter
}

// func (f *MockRequestFilter) HandleRequest(h *Handler, args *FilterArgs, rw http.ResponseWriter, req *http.Request) (*http.Response, error) {
// 	status := args.Get("status")
// }

// func (f *MockRequestFilter) Filter(req *http.Request) (args *FilterArgs, err error) {
// 	return &FilterArgs{}, nil
// }
