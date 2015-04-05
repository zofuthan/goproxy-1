package certutil

import (
	"crypto/tls"
	"fmt"
	"golang.org/x/net/publicsuffix"
	"strings"
	"time"
)

type CA interface {
	Dump(filename string) error
	Issue(host string, vaildFor time.Duration, rsaBits int) (*tls.Certificate, error)
	IssueFile(host string, vaildFor time.Duration, rsaBits int) (string, error)
}

func GetCommonName(domain string) (host string, err error) {
	eTLD_1, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		return
	}

	prefix := strings.TrimRight(strings.TrimSuffix(domain, eTLD_1), ".")
	if strings.Contains(prefix, ".") {
		host = fmt.Sprintf("%s.%s", strings.SplitN(prefix, ".", 2)[1], eTLD_1)
	} else {
		host = eTLD_1
	}
	return
}
