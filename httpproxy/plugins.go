package httpproxy

import (
	"fmt"
	"net/http"
)

type RequestPlugin interface {
	HandleRequest(*http.Request, *PluginArgs) (*http.Response, error)
}

type ResponsePlugin interface {
	HandleResponse(*http.Request, *PluginArgs) (*http.Response, error)
}

type PluginArgs map[string]interface{}

func (f *PluginArgs) GetString(name string) (string, error) {
	v, ok := (*f)[name]
	if !ok {
		return "", fmt.Errorf("PluginArgs(%#v) cannot GetString(%#v)", f, name)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("PluginArgs(%#v) cannot convert %#v to string", f, v)
	}
	return s, nil
}

func (f *PluginArgs) GetInt(name string) (int, error) {
	v, ok := (*f)[name]
	if !ok {
		return 0, fmt.Errorf("PluginArgs(%#v) cannot GetInt(%#v)", f, name)
	}
	s, ok := v.(int)
	if !ok {
		return 0, fmt.Errorf("PluginArgs(%#v) cannot convert %#v to int", f, v)
	}
	return s, nil
}

func (f *PluginArgs) GetStringMap(name string) (map[string]string, error) {
	v, ok := (*f)[name]
	if !ok {
		return nil, fmt.Errorf("PluginArgs(%#v) cannot GetStringMap(%#v)", f, name)
	}
	s, ok := v.(map[string]string)
	if !ok {
		return nil, fmt.Errorf("PluginArgs(%#v) cannot convert %#v to map[string]string", f, v)
	}
	return s, nil
}

func (f *PluginArgs) GetHeader(name string) (*http.Header, error) {
	v, ok := (*f)[name]
	if !ok {
		return nil, fmt.Errorf("PluginArgs(%#v) cannot GetHeader(%#v)", f, name)
	}
	s, ok := v.(*http.Header)
	if !ok {
		return nil, fmt.Errorf("PluginArgs(%#v) cannot convert %#v to *http.Header", f, v)
	}
	return s, nil
}
