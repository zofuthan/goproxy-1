package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/phuslu/goproxy/httpproxy"
	"github.com/phuslu/goproxy/httpproxy/filters"
	_ "github.com/phuslu/goproxy/httpproxy/filters/direct"
	_ "github.com/phuslu/goproxy/httpproxy/filters/imagez"
	_ "github.com/phuslu/goproxy/httpproxy/filters/mock"
	"github.com/phuslu/goproxy/netutil"
	"net/http"
	"time"
)

func main() {
	addr := *flag.String("addr", "127.0.0.1:8000", "GoProxy Listen Address")
	flag.Set("logtostderr", "true")
	flag.Parse()

	ln, err := netutil.Listen("tcp", addr)
	if err != nil {
		glog.Fatalf("Listen(\"tcp\", %s) failed: %s", addr, err)
	}

	directFilter, err := filters.NewFilter("direct")
	if err != nil {
		glog.Fatalf("filters.NewFilter(\"direct\") failed: %s", err)
	}
	h := httpproxy.Handler{
		Listener:       ln,
		RequestFilters: []filters.RequestFilter{},
		RoundTripFilters: []filters.RoundTripFilter{
			directFilter.(filters.RoundTripFilter),
		},
		ResponseFilters: []filters.ResponseFilter{},
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
