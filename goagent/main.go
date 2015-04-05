package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/phuslu/goproxy/httpproxy"
	"github.com/phuslu/goproxy/httpproxy/filters"
	_ "github.com/phuslu/goproxy/httpproxy/filters/direct"
	_ "github.com/phuslu/goproxy/httpproxy/filters/imagez"
	_ "github.com/phuslu/goproxy/httpproxy/filters/mock"
	_ "github.com/phuslu/goproxy/httpproxy/filters/strip"
	"github.com/phuslu/goproxy/netutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	common, err := ReadConfigFile("proxy.ini")
	if err != nil {
		glog.Fatalf("ReadConfigFile() failed: %s", err)
	}

	addr := net.JoinHostPort(common.ListenIp, strconv.Itoa(common.ListenPort))
	ln, err := netutil.Listen("tcp4", addr)
	if err != nil {
		glog.Fatalf("Listen(\"tcp4\", %s) failed: %s", addr, err)
	}

	resolver := netutil.NewResolver(nil)
	for name, iplist := range common.IplistMap {
		resolver.SetHost(name, iplist)
	}
	for host, name := range common.HostMap {
		resolver.SetCNAME(host, name)
	}

	dialer := &netutil.Dialer{
		Timeout:     30 * time.Second,
		KeepAlive:   30 * time.Second,
		DNSResolver: resolver,
	}

	stripFiler, err := filters.NewFilter("strip")
	if err != nil {
		glog.Fatalf("filters.NewFilter(\"strip\") failed: %s", err)
	}
	directFilter, err := filters.NewFilter("direct")
	if err != nil {
		glog.Fatalf("filters.NewFilter(\"direct\") failed: %s", err)
	}
	h := httpproxy.Handler{
		Listener: ln,
		Transport: &http.Transport{
			Dial:                  dialer.Dial,
			DialTLS:               dialer.DialTLS,
			TLSHandshakeTimeout:   2 * time.Second,
			ResponseHeaderTimeout: 2 * time.Second,
			DisableKeepAlives:     true,
			DisableCompression:    true,
			Proxy:                 nil,
		},
		RequestFilters: []filters.RequestFilter{
			stripFiler.(filters.RequestFilter),
		},
		FetchFilters: []filters.FetchFilter{
			directFilter.(filters.FetchFilter),
		},
		ResponseFilters: []filters.ResponseFilter{},
	}
	s := &http.Server{
		Handler:        h,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	common.WriteSummary(os.Stderr)
	glog.Infof("ListenAndServe on %s\n", h.Listener.Addr().String())
	glog.Exitln(s.Serve(h.Listener))
}
