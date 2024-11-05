package internal

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Cache interface {
	Has(string) bool
	Get(string) ([]byte, bool)
	GetInt(string) (int, bool)
	GetHeaders(string) (*http.Header, bool)
	Set(string, []byte) error
	SetInt(string, int) error
	SetHeaders(string, *http.Header) error
}

type ProxyObject struct {
	origin *url.URL
	cache  Cache
}

func NewProxyObject(cache Cache, origin *url.URL) *ProxyObject {
	return &ProxyObject{
		origin: origin,
		cache:  cache,
	}
}

func (p *ProxyObject) Start(host string, port int) {
	// Start the proxy server
	http.HandleFunc("/", p.handleRequest)
	log.Printf("Proxy server started on %s:%d, forward request to :%s\n", host, port, p.origin.Host)
	log.Fatalln(http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil))
}

func (p *ProxyObject) handleRequest(writer http.ResponseWriter, request *http.Request) {
	// Handle the request
	if isNotSafeMethod(request.Method) {
		fmt.Println("Method not safe")
		writer.Header().Set("X-Cache", "MISS")
		p.proxyRequest(writer, request, "")
		return
	}
	// generate a cache key based on the request
	cacheKey := generateCacheKey(request)
	isCached := p.cache.Has(cacheKey)

	if isCached {
		fmt.Println("Cache hit")
		p.responseFromCache(writer, cacheKey)
	} else {
		fmt.Println("Cache miss")
		p.proxyRequest(writer, request, cacheKey)
	}
}

func isNotSafeMethod(method string) bool {
	// Check if the method is not safe
	method = strings.ToUpper(method)
	return method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions
}

func generateCacheKey(request *http.Request) string {
	// Generate a cache key based on the request
	return request.Method + strings.TrimPrefix(request.URL.String(), "/")
}

func (p *ProxyObject) responseFromCache(writer http.ResponseWriter, key string) {
	// Get the response from the cache and write it to the writer
	data, _ := p.cache.Get(key)
	headers, _ := p.cache.GetHeaders(key)
	for key, values := range *headers {
		writer.Header()[key] = values
	}
	writer.Header().Set("X-Cache", "HIT")
	status, ok := p.cache.GetInt(key)
	if !ok {
		writer.WriteHeader(status)
	}
	if data != nil {
		_, _ = writer.Write(data)
	}
}

func (p *ProxyObject) proxyRequest(writer http.ResponseWriter, request *http.Request, key string) {
	// Forward the request to the origin server
	resp, err := p.getResponseFromOrigin(request)
	if err != nil {
		writer.WriteHeader(http.StatusBadGateway)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	for key, values := range resp.Header {
		writer.Header()[key] = values
	}
	err = p.cache.SetInt(key, resp.StatusCode)
	if err != nil {
		fmt.Printf("error setting cache: %s\n", err)
		return
	}
	err = p.cache.SetHeaders(key, &resp.Header)
	if err != nil {
		fmt.Printf("error setting cache: %s\n", err)
		return
	}
	err = p.cache.Set(key, respBody)
	if err != nil {
		fmt.Printf("error setting cache: %s\n", err)
		return
	}
	writer.Header().Set("X-Cache", "MISS")
	writer.WriteHeader(resp.StatusCode)
	_, _ = writer.Write(respBody)
}

func (p *ProxyObject) getResponseFromOrigin(request *http.Request) (*http.Response, error) {
	// Get the response from the origin server
	originURL := *p.origin
	originURL.Path = request.URL.Path
	originURL.RawQuery = request.URL.RawQuery
	originReq, err := http.NewRequest(request.Method, originURL.String(), request.Body)
	if err != nil {
		return nil, err
	}
	originReq.Header = request.Header.Clone()
	client := http.Client{}
	return client.Do(originReq)

}
