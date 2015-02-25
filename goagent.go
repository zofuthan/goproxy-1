package main

import (
	"flag"
	"github.com/golang/glog"
	"net/http"
	"time"
)

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	// ca := NewCA("GoAgent", 2048)
	// ca.Create("CA.crt", 365*24*time.Hour)
	// cert, _ := ca.Issue("www.google.com", 365*24*time.Hour)
	// glog.Infof("cert %#v", cert)

	addr := ":1080"
	ln, err := Listen("tcp4", addr)
	if err != nil {
		glog.Fatalf("Listen(\"tcp\", %s) failed: %s", addr, err)
	}
	h := Handler{
		Listener: ln,
		Net:      &SimpleNetwork{},
		RequestFilters: []RequestFilter{
			&StripRequestFilter{},
			&DirectRequestFilter{},
		},
		ResponseFilters: []ResponseFilter{
			&AlwaysRawResponseFilter{
				Sites: []string{"www.baidu.com"},
			},
			&ImageResponseFilter{},
			&RawResponseFilter{},
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
