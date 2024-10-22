package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "net/http"
    "net/url"
    "time"
    "github.com/elazarl/goproxy"
)

func main() {
    fmt.Println("Proxy server starting on port 8080...")
    proxy := goproxy.NewProxyHttpServer()
    proxy.Verbose = true

    // Log all requests
    proxy.OnRequest().DoFunc(
        func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
            fmt.Printf("Received request for: %s\n", r.URL.String())
            return r, nil
        })

    // Configure proxy transport
    proxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
        return nil, nil
    }

    proxy.Tr.DialContext = (&net.Dialer{
        Timeout:   30 * time.Second,
        KeepAlive: 30 * time.Second,
        Resolver: &net.Resolver{
            PreferGo: true,
            Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
                fmt.Printf("Attempting to dial: %s\n", address)
                return net.Dial(network, address)
            },
        },
    }).DialContext

    // Block YouTube during specific hours
    proxy.OnRequest(goproxy.DstHostIs("www.youtube.com")).DoFunc(
        func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
            currentHour := time.Now().Hour()
            fmt.Printf("YouTube request detected at hour: %d\n", currentHour)
            if currentHour >= 9 && currentHour <= 23 {
                return r, goproxy.NewResponse(r, 
                    goproxy.ContentTypeText, 
                    http.StatusForbidden, 
                    "Access to this site is blocked during working hours.")
            }
            return r, nil
        })

    // Start server with error checking
    server := &http.Server{
        Addr:    "[::]:8080", // Listen on both IPv4 and IPv6
        Handler: proxy,
    }

    // Try to start the server
    fmt.Println("Server is attempting to bind to port 8080...")
    err := server.ListenAndServe()
    if err != nil {
        log.Printf("Failed to start server: %v\n", err)
        return
    }
}
