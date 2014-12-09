package main

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"
)

type ImageResponsePlugin struct {
	ResponsePlugin
}

func (p ImageResponsePlugin) HandleResponse(c *PluginContext, rw http.ResponseWriter, req *http.Request, res *http.Response, resError error) error {
	if resError != nil {
		rw.WriteHeader(502)
		fmt.Fprintf(rw, "Error: %s\n", resError)
		c.H.Log.Printf("ImageResponsePlugin HandleResponse %s failed %s", req.Host, resError)
		return resError
	}
	if !strings.HasPrefix(res.Header.Get("Content-Type"), "image/") {
		io.Copy(rw, res.Body)
		return nil
	}
	img, _, err := image.Decode(res.Body)
	if err != nil {
		c.H.Log.Printf("ImageResponsePlugin HandleResponse failed %s", err)
		return err
	}
	rw.WriteHeader(200)
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
