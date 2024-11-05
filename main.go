package main

import (
	"flag"
	"github.com/joiller/cache-proxy/internal"
	"net/url"
)

func main() {
	origin := flag.String("origin", "", "origin")
	port := flag.Int("port", 8080, "port")
	clearCache := flag.Bool("clear-cache", false, "clear cache")
	flag.Parse()
	cache := internal.NewCacheObject()

	if *clearCache {
		// clear cache
		cache.ClearCache()
		return
	}

	u, _ := url.ParseRequestURI(*origin)
	proxy := internal.NewProxyObject(cache, u)
	proxy.Start("localhost", *port)
}
