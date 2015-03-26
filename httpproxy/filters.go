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
	if req.URL.Scheme == "http" && f.ForcehttpsSites != nil {
		for _, suffix := range f.ForcehttpsSites {
			if strings.HasSuffix(req.Host, suffix) && !strings.HasPrefix(req.Referer(), "https:") {
				force := false
				if f.NoforcehttpsSites != nil {
					force = true
				} else if _, ok := f.NoforcehttpsSites[req.Host]; !ok {
					force = true
				}
				if force {
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

type FakehttpsRequestFilter struct {
	*StripRequestFilter
	FakehttpsSites   []string
	NofakehttpsSites map[string]struct{}
}

func (f *FakehttpsRequestFilter) Filter(req *http.Request) (args *FilterArgs, err error) {
	if req.URL.Scheme == "https" && f.FakehttpsSites != nil {
		for _, suffix := range f.FakehttpsSites {
			if strings.HasSuffix(req.Host, suffix) {
				fake := false
				if f.NofakehttpsSites == nil {
					fake = true
				} else if _, ok := f.NofakehttpsSites[req.Host]; !ok {
					fake = true
				}
				if fake {
					return &FilterArgs{}, nil
				}
			}
		}
	}
	return nil, nil
}
