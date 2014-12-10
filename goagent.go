package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// ca := NewCA("GoAgent", 2048)
	// ca.Create("CA.crt", 365*24*time.Hour)
	// cert, _ := ca.Issue("www.google.com", 365*24*time.Hour)
	// log.Printf("cert %#v", cert)

	addr := ":1080"
	ln, err := Listen("tcp4", addr)
	if err != nil {
		log.Fatalf("Listen(\"tcp\", %s) failed: %s", addr, err)
	}
	h := Handler{
		Listener: ln,
		Log:      log.New(os.Stderr, "INFO - ", 3),
		Net:      &SimpleNetwork{},
		RequestFilters: []RequestFilter{
			&StripRequestFilter{},
			&DirectRequestFilter{},
		},
		ResponseFilters: []ResponseFilter{
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
	h.Log.Printf("ListenAndServe on %s\n", h.Listener.Addr().String())
	h.Log.Fatal(s.Serve(h.Listener))
}
