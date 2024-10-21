package main

import (
	"log"
	"net/http"
	"time"

	"github.com/elazarl/goproxy"
)

func main() {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	// Block YouTube and Instagram during specific hours
	proxy.OnRequest(goproxy.DstHostIs(
		"www.youtube.com",
		"www.instagram.com")).DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			currentHour := time.Now().Hour()
			if currentHour >= 9 && currentHour <= 23 { // Block from 9 AM to 5 PM
				return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusForbidden, "Access to this site is blocked during working hours.")
			}
			return r, nil
		})

	log.Fatal(http.ListenAndServe(":8080", proxy))
}
