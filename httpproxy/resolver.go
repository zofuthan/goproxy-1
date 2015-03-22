package httpproxy

import (
	"github.com/phuslu/goproxy/dnsclient"
	"net"
	"strings"
	"sync"
)

type Resolver interface {
	LookupHost(name string) (addrs []string, err error)
	LookupIP(name string) (addrs []net.IP, err error)
	LookupCNAME(name string) (cname string, err error)
	SetCNAME(name, cname string)
	SetHost(host string, addrs []string)
}

type resolver struct {
	dnsServers []string
	cnames     map[string]string
	hosts      map[string][]string
	rwLock     *sync.RWMutex
}

func NewResolver(dnsServers []string) Resolver {
	return &resolver{
		dnsServers: dnsServers,
		cnames:     make(map[string]string),
		hosts:      make(map[string][]string),
		rwLock:     &sync.RWMutex{},
	}
}

func (r *resolver) lookupHostInMemory(name string) (addrs []string, err error) {
	r.rwLock.RLock()
	defer r.rwLock.RUnlock()
	for suffix, cname := range r.cnames {
		if strings.HasSuffix(name, suffix) {
			name = cname
			break
		}
	}
	if hosts, ok := r.hosts[name]; ok {
		addrs = hosts
	}
	return addrs, nil
}

func (r *resolver) LookupHost(name string) (addrs []string, err error) {
	addrs, err = r.lookupHostInMemory(name)
	if err == nil && addrs != nil {
		return addrs, nil
	}
	options := &dnsclient.LookupOptions{
		DNSServers: r.dnsServers,
		Net:        "udp",
	}
	return dnsclient.LookupHost(name, options)
}

func (r *resolver) LookupIP(name string) (addrs []net.IP, err error) {
	options := &dnsclient.LookupOptions{
		DNSServers: r.dnsServers,
		Net:        "udp",
	}
	return dnsclient.LookupIP(name, options)
}

func (r *resolver) LookupCNAME(name string) (cname string, err error) {
	options := &dnsclient.LookupOptions{
		DNSServers: r.dnsServers,
		Net:        "udp",
	}
	return dnsclient.LookupCNAME(name, options)
}

func (r *resolver) SetCNAME(suffix, cname string) {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	r.cnames[suffix] = cname
}

func (r *resolver) SetHost(name string, addrs []string) {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	r.hosts[name] = addrs
}
