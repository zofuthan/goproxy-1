package httpproxy

import (
	"github.com/golang/glog"
	"net/http"
	"strings"
)

type AlwaysRawResponseFilter struct {
	*RawResponseFilter
	Sites []string
}

func (f *AlwaysRawResponseFilter) Filter(res *http.Response) (args *FilterArgs, err error) {
	host := res.Request.Header.Get("Host")
	if host != "" && f.Sites != nil {
		for _, site := range f.Sites {
			glog.Infof("host %#v site %#v", host, site)
			if host == site {
				return &FilterArgs{}, nil
			}
		}
	}
	return nil, nil
}

type ForcehttpsRequestFilter struct {
	*MockRequestFilter
	ForcehttpsSites   []string
	NoforcehttpsSites map[string]struct{}
}

func (f *ForcehttpsRequestFilter) Filter(req *http.Request) (args *FilterArgs, err error) {
	if f.ForcehttpsSites != nil && f.NoforcehttpsSites != nil {
		for _, suffix := range f.ForcehttpsSites {
			if strings.HasSuffix(req.Host, suffix) {
				if _, ok := f.NoforcehttpsSites[req.Host]; !ok {
					url := strings.Replace(req.URL.String(), "http:", "https:", 1)
					return &FilterArgs{
						"StatusCode": 301,
						"Header": &http.Header{
							"Location": []string{url},
						},
					}, nil
				}
			}
		}
	}
	return nil, nil
}
