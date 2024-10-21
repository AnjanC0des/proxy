package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/elazarl/goproxy"
)

func main() {
	println("Proxy started i think.")
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	proxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
		return nil, nil
	}
	proxy.Tr.DialContext = (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				return net.Dial(network, "8.8.8.8:53") // Using Google's DNS server
			},
		},
	}).DialContext
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
