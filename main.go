package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "net/http"
    "net/url"
    "strings"
    "time"
    "github.com/elazarl/goproxy"
)

// Create a custom condition type for YouTube
type youtubeCondition struct{}

func (y *youtubeCondition) HandleReq(req *http.Request, ctx *goproxy.ProxyCtx) bool {
    host := strings.ToLower(req.URL.Host)
    return strings.Contains(host, "youtube.com") || 
           strings.Contains(host, "youtu.be") ||
           strings.Contains(host, "ytimg.com") ||
           strings.Contains(host, "yt3.ggpht.com") ||
           strings.Contains(host, "googlevideo.com")
}

// Add HandleResp method to satisfy ReqCondition interface
func (y *youtubeCondition) HandleResp(resp *http.Response, ctx *goproxy.ProxyCtx) bool {
    return false // We don't need to handle responses
}

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
    proxy.OnRequest(&youtubeCondition{}).DoFunc(
        func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
            location, err := time.LoadLocation("Asia/Kolkata")
            if err != nil {
                fmt.Println("Error loading timezone:", err)
                return r, nil
            }

            currentHour := time.Now().In(location).Hour()
            fmt.Printf("YouTube request detected at hour: %d for URL: %s\n", currentHour, r.URL.String())

            if currentHour >= 9 && currentHour < 23 {
                return r, goproxy.NewResponse(r,
                    goproxy.ContentTypeHtml,
                    http.StatusForbidden,
                    `<html><body><h1>Access Blocked</h1><p>YouTube access is blocked during working hours (9:00 AM - 11:00 PM IST)</p></body></html>`)
            }
            return r, nil
        })

    // Start server with error checking
    server := &http.Server{
        Addr:    ":8080",
        Handler: proxy,
    }

    fmt.Println("Server is attempting to bind to port 8080...")
    err := server.ListenAndServe()
    if err != nil {
        log.Fatalf("Failed to start server: %v\n", err)
    }
}
