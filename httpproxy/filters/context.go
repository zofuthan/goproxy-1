package filters

import (
	"fmt"
	"net"
	"net/http"
)

type Context map[string]interface{}

func (f *Context) Get(name string) (interface{}, error) {
	v, ok := (*f)[name]
	if !ok {
		return nil, fmt.Errorf("Context(%#v) cannot Get(%#v)", f, name)
	}
	return v, nil
}

func (f *Context) GetString(name string) (string, error) {
	v, ok := (*f)[name]
	if !ok {
		return "", fmt.Errorf("Context(%#v) cannot GetString(%#v)", f, name)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("Context(%#v) cannot convert %#v to string", f, v)
	}
	return s, nil
}

func (f *Context) GetInt(name string) (int, error) {
	v, ok := (*f)[name]
	if !ok {
		return 0, fmt.Errorf("Context(%#v) cannot GetInt(%#v)", f, name)
	}
	s, ok := v.(int)
	if !ok {
		return 0, fmt.Errorf("Context(%#v) cannot convert %#v to int", f, v)
	}
	return s, nil
}

func (f *Context) GetStringMap(name string) (map[string]string, error) {
	v, ok := (*f)[name]
	if !ok {
		return nil, fmt.Errorf("Context(%#v) cannot GetStringMap(%#v)", f, name)
	}
	s, ok := v.(map[string]string)
	if !ok {
		return nil, fmt.Errorf("Context(%#v) cannot convert %#v to map[string]string", f, v)
	}
	return s, nil
}

func (f *Context) GetHeader(name string) (*http.Header, error) {
	v, ok := (*f)[name]
	if !ok {
		return nil, fmt.Errorf("Context(%#v) cannot GetHeader(%#v)", f, name)
	}
	s, ok := v.(*http.Header)
	if !ok {
		return nil, fmt.Errorf("Context(%#v) cannot convert %#v to *http.Header", f, v)
	}
	return s, nil
}

func (f *Context) GetResponseWriter() http.ResponseWriter {
	name := "__responsewriter__"
	v, ok := (*f)[name]
	if !ok {
		panic(fmt.Errorf("Context(%#v) cannot GetResponseWriter(%#v)", f, name))
	}
	rw, ok := v.(http.ResponseWriter)
	if !ok {
		panic(fmt.Errorf("Context(%#v) cannot convert %#v to http.ResponseWriter", f, v))
	}
	return rw
}

func (f *Context) GetTransport() *http.Transport {
	name := "__transport__"
	v, ok := (*f)[name]
	if !ok {
		panic(fmt.Errorf("Context(%#v) cannot GetTransport(%#v)", f, name))
	}
	tr, ok := v.(*http.Transport)
	if !ok {
		panic(fmt.Errorf("Context(%#v) cannot convert %#v to *http.Transport", f, v))
	}
	return tr
}

func (f *Context) GetListener() net.Listener {
	name := "__listener__"
	v, ok := (*f)[name]
	if !ok {
		panic(fmt.Errorf("Context(%#v) cannot GetListener(%#v)", f, name))
	}
	ln, ok := v.(net.Listener)
	if !ok {
		panic(fmt.Errorf("Context(%#v) cannot convert %#v to net.Listener", f, v))
	}
	return ln
}
