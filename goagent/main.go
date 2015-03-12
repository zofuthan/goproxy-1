package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/phuslu/goproxy/httpproxy"
	"github.com/phuslu/goproxy/net2"
	"github.com/phuslu/goproxy/rootca"
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

	addr := ":1080"
	ln, err := httpproxy.Listen("tcp4", addr)
	if err != nil {
		glog.Fatalf("Listen(\"tcp4\", %s) failed: %s", addr, err)
	}
	h := httpproxy.Handler{
		Listener: ln,
		Net:      &net2.SimpleNetwork{},
		RequestFilters: []httpproxy.RequestFilter{
			&httpproxy.StripRequestFilter{CA: ca},
			&GAERequestFilter{
				AppIDs: []string{"goagenta"},
				Schema: "https",
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
