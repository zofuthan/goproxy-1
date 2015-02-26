package goproxy

import (
	"github.com/golang/glog"
	"net"
	"time"
)

type PushListener interface {
	net.Listener
	Push(net.Conn, error)
}

type listenerAcceptTuple struct {
	c net.Conn
	e error
}

type listener struct {
	net.Listener
	ln net.Listener
	ch chan listenerAcceptTuple
}

func Listen(network string, addr string) (net.Listener, error) {
	ln, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	l := &listener{
		ln: ln,
		ch: make(chan listenerAcceptTuple, 200),
	}
	// http://golang.org/src/pkg/net/http/server.go
	go func(ln net.Listener, ch chan listenerAcceptTuple) {
		var tempDelay time.Duration
		for {
			c, e := ln.Accept()
			ch <- listenerAcceptTuple{c, e}
			if e != nil {
				if ne, ok := e.(net.Error); ok && ne.Temporary() {
					if tempDelay == 0 {
						tempDelay = 5 * time.Millisecond
					} else {
						tempDelay *= 2
					}
					if max := 1 * time.Second; tempDelay > max {
						tempDelay = max
					}
					glog.Infof("http: Accept error: %v; retrying in %v", e, tempDelay)
					time.Sleep(tempDelay)
					continue
				}
				return
			}
		}
	}(l.ln, l.ch)
	return l, nil
}

func (l listener) Accept() (net.Conn, error) {
	t := <-l.ch
	return t.c, t.e
}

func (l listener) CLose() error {
	return l.ln.Close()
}

func (l listener) Addr() net.Addr {
	return l.ln.Addr()
}

func (l listener) Push(conn net.Conn, err error) {
	l.ch <- listenerAcceptTuple{conn, err}
}
