package imagez

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/chai2010/webp"
	"github.com/ginuerzh/weedo"
	"github.com/golang/glog"
	"github.com/phuslu/goproxy/httpproxy/filters"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
)

var (
	defaultWeedClient *weedo.Client = nil
)

type Filter struct {
	UnderlayFilter string
	WeedClient     *weedo.Client
}

func init() {
	weedMaster := *flag.String("weedmaster", "localhost:9333", "weed master address")
	defaultWeedClient = weedo.NewClient(weedMaster)

	filters.Register("imagez", &filters.RegisteredFilter{
		New: NewFilter,
	})
}

func NewFilter() (filters.Filter, error) {
	return &Filter{
		UnderlayFilter: "direct",
		WeedClient:     defaultWeedClient,
	}, nil
}

func (p *Filter) FilterName() string {
	return "imagez"
}

func (p *Filter) RoundTrip(ctx *filters.Context, req *http.Request) (*filters.Context, *http.Response, error) {
	// p.WeedClient.GetUrl(fid)

	f1, err := filters.NewFilter(p.UnderlayFilter)
	if err != nil {
		return ctx, nil, err
	}
	f2, ok := f1.(filters.RoundTripFilter)
	if !ok {
		return ctx, nil, fmt.Errorf("%#v was not a filters.RoundTripFilter", f1)
	}

	ctx, resp, err := f2.RoundTrip(ctx, req)
	if err != nil {
		return ctx, nil, err
	}

	return ctx, resp, nil
}

func (p *Filter) Response(ctx *filters.Context, resp *http.Response) (*filters.Context, *http.Response, error) {
	switch resp.Header.Get("Content-Type") {
	case "image/gif":
	case "image/png":
	case "image/jpeg":
		break
	default:
		return ctx, resp, nil
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		glog.Errorf("%s decode %#v error: %v", p.FilterName(), resp, err)
		return ctx, nil, err
	}

	var b bytes.Buffer
	err = webp.Encode(&b, img, &webp.Options{Lossless: true})
	if err != nil {
		glog.Errorf("%s encode %#v error: %v", p.FilterName(), img, err)
		return ctx, nil, err
	}

	resp.Header.Set("Content-Type", "image/webp")
	resp.ContentLength = int64(b.Len())

	return ctx, resp, nil
}
