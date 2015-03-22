package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/phuslu/goproxy/httpproxy"
	"github.com/phuslu/goproxy/rootca"
	"net"
	"net/http"
	"os"
	"time"
)

func getCA() (*rootca.RootCA, error) {
	filename := "CA.crt"
	_, err := os.Stat(filename)
	var ca *rootca.RootCA
	if err == nil {
		ca, err = rootca.NewCAFromFile(filename)
		if err != nil {
			return nil, err
		}
	} else {
		ca, err = rootca.NewCA("GoAgent", 3*365*24*time.Hour, 2048)
		if err != nil {
			return nil, err
		}
		if err = ca.Dump("CA.crt"); err != nil {
			return nil, err
		}
	}
	return ca, nil
}

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	ca, err := getCA()
	if err != nil {
		glog.Fatalf("getCA() failed: %s", err)
	}

	common, err := ReadConfigFile("proxy.ini")
	if err != nil {
		glog.Fatalf("ReadConfigFile() failed: %s", err)
	}

	glog.Infof("common=%#v", common)

	addr := net.JoinHostPort(common.ListenIp, common.ListenPassword)
	ln, err := httpproxy.Listen("tcp4", addr)
	if err != nil {
		glog.Fatalf("Listen(\"tcp4\", %s) failed: %s", addr, err)
	}

	google_hk := []string{
		"58.176.217.88",
		"58.176.217.99",
		"58.176.217.104",
		"58.176.217.109",
		"58.176.217.114",
	}
	resolver := httpproxy.NewResolver(nil)
	resolver.SetHost("google_hk", google_hk)
	resolver.SetCNAME(".appspot.com", "google_hk")
	resolver.SetCNAME(".google.com", "google_hk")

	dialer := &httpproxy.Dialer{
		Timeout:     30 * time.Second,
		KeepAlive:   30 * time.Second,
		DNSResolver: resolver,
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
		RequestFilters: []httpproxy.RequestFilter{
			&httpproxy.StripRequestFilter{CA: ca},
			&GAERequestFilter{
				AppIDs: []string{"phuslua"},
				Scheme: "https",
			},
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
