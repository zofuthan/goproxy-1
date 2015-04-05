package strip

import (
	"crypto/tls"
	"fmt"
	"github.com/golang/glog"
	"github.com/phuslu/goproxy/certutil"
	"github.com/phuslu/goproxy/httpproxy/filters"
	"github.com/phuslu/goproxy/netutil"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

type Filter struct {
	filters.RequestFilter
}

const (
	CAFilename string        = "CA.crt"
	CAName     string        = "GoAgent"
	CAExpires  time.Duration = 3 * 365 * 24 * time.Hour
)

var (
	ca certutil.CA
)

func init() {
	var err error

	if _, err = os.Stat(CAFilename); err == nil {
		ca, err = certutil.NewStdCAFromFile(CAFilename)
		if err != nil {
			panic(err)
		}
	} else {
		ca, err = certutil.NewStdCA(CAName, CAExpires, 2048)
		if err != nil {
			panic(err)
		}
		if err = ca.Dump(CAFilename); err != nil {
			panic(err)
		}
	}

	filters.Register("strip", &filters.RegisteredFilter{
		New: NewFilter,
	})
}

func NewFilter() (filters.Filter, error) {
	return &Filter{}, nil
}

func (f *Filter) FilterName() string {
	return "strip"
}

func (f *Filter) Request(ctx *filters.Context, req *http.Request) (*filters.Context, *http.Request, error) {
	if req.Method != "CONNECT" {
		return ctx, req, nil
	}

	hijacker, ok := ctx.GetResponseWriter().(http.Hijacker)
	if !ok {
		return ctx, nil, fmt.Errorf("%#v does not implments Hijacker", ctx.GetResponseWriter())
	}

	conn, _, err := hijacker.Hijack()
	if err != nil {
		return ctx, nil, fmt.Errorf("http.ResponseWriter Hijack failed: %s", err)
	}

	_, err = io.WriteString(conn, "HTTP/1.1 200 OK\r\n\r\n")
	if err != nil {
		return ctx, nil, err
	}

	glog.Infof("%s \"STRIP %s %s %s\" - -", req.RemoteAddr, req.Method, req.Host, req.Proto)

	cert, err := ca.Issue(req.Host, 3*365*24*time.Hour, 2048)
	if err != nil {
		return ctx, nil, fmt.Errorf("tls.LoadX509KeyPair failed: %s", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*cert},
		ClientAuth:   tls.VerifyClientCertIfGiven,
	}
	tlsConn := tls.Server(conn, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		return ctx, nil, fmt.Errorf("tlsConn.Handshake error: %s", err)
	}

	if pln, ok := ctx.GetListener().(netutil.PushListener); ok {
		pln.Push(tlsConn, nil)
		// glog.Infof("%#v Push %#v\n", pln, tlsConn)
		return ctx, nil, nil
	}

	loConn, err := net.Dial("tcp", ctx.GetListener().Addr().String())
	if err != nil {
		return ctx, nil, fmt.Errorf("net.Dial failed: %s", err.Error())
	}

	go io.Copy(loConn, tlsConn)
	go io.Copy(tlsConn, loConn)
	return ctx, nil, nil
}
