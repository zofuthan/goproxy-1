// Copyright 2012 Phus Lu. All rights reserved.

package gae

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"appengine"
	"appengine/urlfetch"
)

const (
	Version  = "1.0"
	Password = ""

	FetchMaxSize       = 1024 * 1024 * 4
	Deadline     int64 = 30 * 1e9
)

func favicon(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func robots(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User-agent: *\nDisallow: /\n"))
}

func handler(w http.ResponseWriter, r *http.Request) {
	context := appengine.NewContext(r)

	if "gzip" == r.Header.Get("Content-Encoding") {
		b, err := gzip.NewReader(r.Body)
		r.Body = b
		if err != nil {
			context.Criticalf("gzip.NewReader(%#v) return %#v", r.Body, err)
		}
	}

	req, err := http.ReadRequest(r.Body)
	if err != nil {
		context.Criticalf("http.ReadRequest(%#v) return %#v", r.Body, err)
	}

	password := r.Header.Get("X-Fetch-Password")
	if password != "" {
		r.Header.Del(key)
	}
	if Password != "" && Password != password {
		http.Error(w, "Wrong Password.", 403)
	}

	deadline := time.Duration(Deadline)

	var errors []string
	for i := 0; i < 2; i++ {
		t := &urlfetch.Transport{Context: context, Deadline: deadline, AllowInvalidServerCertificate: true}
		resp, err := t.RoundTrip(req)
		if err != nil {
			message := err.String()
			errors = append(errors, message)
			if strings.Contains(message, "FETCH_ERROR") {
				context.Errorf("URLFetchServiceError_FETCH_ERROR(type=%T, deadline=%v, url=%v)", err, deadline, req.URL)
				time.Sleep(time.Second)
				deadline = time.Duration(Deadline * 2)
			} else if strings.Contains(message, "DEADLINE_EXCEEDED") {
				context.Errorf("URLFetchServiceError_DEADLINE_EXCEEDED(type=%T, deadline=%v, url=%v)", err, deadline, req.URL)
				time.Sleep(time.Second)
				deadline = time.Duration(Deadline * 2)
			} else if strings.Contains(message, "INVALID_URL") {
				http.Error(w, fmt.Sprintf("Invalid URL: %v", err), 501)
				return
			} else if strings.Contains(message, "RESPONSE_TOO_LARGE") {
				context.Errorf("URLFetchServiceError_RESPONSE_TOO_LARGE(type=%T, deadline=%v, url=%v)", err, deadline, req.URL)
				req.Header.Set("Range", fmt.Sprintf("bytes=0-%d", FetchMaxSize))
				deadline = time.Duration(Deadline * 2)
			} else {
				context.Errorf("URLFetchServiceError UNKOWN(type=%T, deadline=%v, url=%v, error=%v)", err, deadline, req.URL, err)
				time.Sleep(4 * time.Second)
			}
			continue
		}
		if resp.TransferEncoding != nil || resp.ContentLength >= 1024*1024 {
			w.Header().Add("Content-Encoding", "gzip")
			w, err = gzip.NewWriter(w)
			if err != nil {
				context.Criticalf("gzip.NewWriter(%#v) return %#v", w, err)
			}
		}
		fmt.Fprintf(w, "%s %s\r\n", resp.Proto, resp.Status)
		for key, values := range resp.Header {
			for _, value := range values {
				fmt.Fprintf(w, "%s: %s\r\n", key, value)
			}
		}
		w.Write("\r\n")
		io.Copy(w, resp.Body)
		return
	}
	http.Error(w, fmt.Sprintf("Go Server Fetch Failed: %v", errors), 502)
}

func root(w http.ResponseWriter, r *http.Request) {
	context := appengine.NewContext(r)
	version, _ := strconv.ParseInt(strings.Split(appengine.VersionID(context), ".")[1], 10, 64)
	ctime := time.Unix(version/(1<<28)+8*3600, 0).Format(time.RFC3339)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "GoAgent go server %s works, deployed at %s\n", Version, ctime)
}

func init() {
	http.HandleFunc("/favicon.ico", favicon)
	http.HandleFunc("/robots.txt", robots)
	http.HandleFunc("/_gh/", handler)
	http.HandleFunc("/", root)
}
