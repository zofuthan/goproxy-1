package main

import (
	"net/http"
	"strings"
)

type DirectRequestFilter struct {
	RequestFilter
}

func (d *DirectRequestFilter) Filter(req *http.Request) (pluginName string, pluginArgs *http.Header, err error) {
	return "direct", nil, nil
}

type DirectResponseFilter struct {
	RequestFilter
}

func (d *DirectResponseFilter) Filter(req *http.Request, res *http.Response) (pluginName string, pluginArgs *http.Header, err error) {
	return "direct", nil, nil
}

type StripRequestFilter struct {
	RequestFilter
}

func (d *StripRequestFilter) Filter(req *http.Request) (pluginName string, pluginArgs *http.Header, err error) {
	if req.Method == "CONNECT" {
		args := http.Header{
			"Foo": []string{"bar"},
			"key": []string{"value"},
		}
		return "strip", &args, nil
	}
	return "", nil, nil
}

type ImageResponseFilter struct {
	RequestFilter
}

func (d *ImageResponseFilter) Filter(req *http.Request, res *http.Response) (pluginName string, pluginArgs *http.Header, err error) {
	if strings.HasPrefix(res.Header.Get("Content-Type"), "image/") {
		return "image", nil, nil
	}
	return "", nil, nil
}
