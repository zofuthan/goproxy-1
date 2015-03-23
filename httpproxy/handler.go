package httpproxy

import (
	"fmt"
	"github.com/golang/glog"
	"net"
	"net/http"
)

type Handler struct {
	http.Handler
	Listener        net.Listener
	Transport       *http.Transport
	RequestFilters  []RequestFilter
	ResponseFilters []ResponseFilter
}

type FilterArgs map[string]interface{}

func (f *FilterArgs) GetString(name string) (string, error) {
	v, ok := (*f)[name]
	if !ok {
		return "", fmt.Errorf("FilterArgs(%#v) cannot GetString(%#v)", f, name)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("FilterArgs(%#v) cannot convert %#v to string", f, v)
	}
	return s, nil
}

func (f *FilterArgs) GetInt(name string) (int, error) {
	v, ok := (*f)[name]
	if !ok {
		return 0, fmt.Errorf("FilterArgs(%#v) cannot GetInt(%#v)", f, name)
	}
	s, ok := v.(int)
	if !ok {
		return 0, fmt.Errorf("FilterArgs(%#v) cannot convert %#v to int", f, v)
	}
	return s, nil
}

func (f *FilterArgs) GetStringMap(name string) (map[string]string, error) {
	v, ok := (*f)[name]
	if !ok {
		return nil, fmt.Errorf("FilterArgs(%#v) cannot GetStringMap(%#v)", f, name)
	}
	s, ok := v.(map[string]string)
	if !ok {
		return nil, fmt.Errorf("FilterArgs(%#v) cannot convert %#v to map[string]string", f, v)
	}
	return s, nil
}

func (f *FilterArgs) GetHeader(name string) (*http.Header, error) {
	v, ok := (*f)[name]
	if !ok {
		return nil, fmt.Errorf("FilterArgs(%#v) cannot GetHeader(%#v)", f, name)
	}
	s, ok := v.(*http.Header)
	if !ok {
		return nil, fmt.Errorf("FilterArgs(%#v) cannot convert %#v to *http.Header", f, v)
	}
	return s, nil
}

type RequestFilter interface {
	HandleRequest(*Handler, *FilterArgs, http.ResponseWriter, *http.Request) (*http.Response, error)
	Filter(req *http.Request) (args *FilterArgs, err error)
}

type ResponseFilter interface {
	HandleResponse(*Handler, *FilterArgs, http.ResponseWriter, *http.Response, error) error
	Filter(res *http.Response) (args *FilterArgs, err error)
}

func (h Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" && !req.URL.IsAbs() {
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
	for i, reqfilter := range h.RequestFilters {
		args, err := reqfilter.Filter(req)
		if err != nil {
			glog.Infof("ServeHTTP RequestFilter error: %v", err)
		}
		if args != nil || i == len(h.RequestFilters)-1 {
			res, err := reqfilter.HandleRequest(&h, args, rw, req)
			if err != nil {
				glog.Infof("ServeHTTP HandleRequest error: %v", err)
			}
			if res == nil {
				return
			}
			res.Request = req
			for j, resfilter := range h.ResponseFilters {
				if resfilter == nil {
					break
				}
				args, err := resfilter.Filter(res)
				if err != nil {
					glog.Infof("ServeHTTP ResponseFilter error: %v", err)
				}
				if args != nil || j == len(h.ResponseFilters)-1 {
					err := resfilter.HandleResponse(&h, args, rw, res, err)
					if err != nil {
						glog.Infof("ServeHTTP HandleResponse error: %v", err)
					}
					break
				}
			}
			break
		}
	}
}
