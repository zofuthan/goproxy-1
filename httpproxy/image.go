package httpproxy

import (
	"fmt"
	"github.com/golang/glog"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"
)

type ImageResponseFilter struct {
	ResponseFilter
}

func (f *ImageResponseFilter) HandleResponse(h *Handler, args *FilterArgs, rw http.ResponseWriter, res *http.Response, resError error) error {
	if resError != nil {
		rw.WriteHeader(502)
		fmt.Fprintf(rw, "Error: %s\n", resError)
		glog.Infof("ImageResponseFilter HandleResponse %s failed %s", res.Request.Host, resError)
		return resError
	}
	if !strings.HasPrefix(res.Header.Get("Content-Type"), "image/") {
		io.Copy(rw, res.Body)
		return nil
	}
	img, _, err := image.Decode(res.Body)
	if err != nil {
		glog.Infof("ImageResponseFilter HandleResponse failed %s", err)
		return err
	}
	for key, values := range res.Header {
		for _, value := range values {
			if key == "Content-Type" {
				rw.Header().Set(key, "image/jpeg")
			} else {
				rw.Header().Set(key, value)
			}
		}
	}
	rw.Header().Set("Connection", "close")
	return jpeg.Encode(rw, img, &jpeg.Options{50})
}

func (f *ImageResponseFilter) Filter(res *http.Response) (args *FilterArgs, err error) {
	if strings.HasPrefix(res.Header.Get("Content-Type"), "image/") {
		return &FilterArgs{}, nil
	}
	return nil, nil
}
