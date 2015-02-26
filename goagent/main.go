package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/phuslu/goproxy/httpproxy"
	"github.com/phuslu/goproxy/rootca"
	"net/http"
	"time"
)

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	ca, err := rootca.NewCA("GoAgent", 3*365*24*time.Hour, 2048)
	if err != nil {
		glog.Fatalf("rootca.NewCA(\"GoAgent\") failed: %s", err)
	}

	addr := ":1080"
	ln, err := httpproxy.Listen("tcp4", addr)
	if err != nil {
		glog.Fatalf("Listen(\"tcp4\", %s) failed: %s", addr, err)
	}
	h := httpproxy.Handler{
		Listener: ln,
		Net:      &httpproxy.SimpleNetwork{},
		RequestFilters: []httpproxy.RequestFilter{
			&httpproxy.StripRequestFilter{CA: ca},
			&httpproxy.DirectRequestFilter{},
		},
		ResponseFilters: []httpproxy.ResponseFilter{
			&httpproxy.RawResponseFilter{},
		},
	}
	s := &http.Server{
		Handler:        h,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	glog.Infof("ListenAndServe on %s\n", h.Listener.Addr().String())
	glog.Exitln(s.Serve(h.Listener))
}
