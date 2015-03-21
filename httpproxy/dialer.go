package main

import (
	"net"
)

type AggressiveDailer struct {
	*net.Dialer
}

type AggressiveTLSDialer struct {
}
