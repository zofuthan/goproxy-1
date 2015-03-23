package httpproxy

import (
	"github.com/golang/glog"
	"net/http"
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

// type ForcehttpsRequestFilter struct {
// 	*
// }
