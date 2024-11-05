package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

type CacheObject struct {
	folder string
}

func NewCacheObject() *CacheObject {
	c := &CacheObject{
		folder: "./tmp/cache",
	}
	c.createFolder()
	return c
}

func (c *CacheObject) createFolder() {
	err := os.MkdirAll(c.folder, os.ModePerm)
	if err != nil {
		log.Fatalf("Error creating cache folder: %s", err)
	}
}

func (c *CacheObject) path(key string) string {
	return c.folder + "/" + key
}

func (c *CacheObject) Has(key string) bool {
	path := c.path(key)
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (c *CacheObject) Get(key string) ([]byte, bool) {
	path := c.path(key)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []byte{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return []byte{}, false
	}
	return data, true
}

func (c *CacheObject) GetInt(key string) (int, bool) {
	data, ok := c.Get(key + "-int")
	if !ok {
		return 0, false
	}
	atoi, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, false
	}
	return atoi, true
}

func (c *CacheObject) GetHeaders(key string) (*http.Header, bool) {
	data, ok := c.Get(key + "-headers")
	if !ok {
		return nil, false
	}
	var headers http.Header
	err := json.Unmarshal(data, &headers)
	if err != nil {
		return nil, false
	}
	return &headers, true
}

func (c *CacheObject) Set(key string, value []byte) error {
	fmt.Println("Setting cache for key:", key)
	path := c.path(key)
	fmt.Println("Path:", path)
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating cache file: %s", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	_, err = file.Write(value)
	if err != nil {
		return fmt.Errorf("error writing to cache file: %s", err)
	}
	return nil
}

func (c *CacheObject) SetInt(key string, value int) error {
	return c.Set(key+"-int", []byte(strconv.Itoa(value)))
}

func (c *CacheObject) SetHeaders(key string, headers *http.Header) error {
	data, err := json.Marshal(headers)
	if err != nil {
		return fmt.Errorf("error marshalling headers: %s", err)
	}
	return c.Set(key+"-headers", data)
}

func (c *CacheObject) ClearCache() {
	// clear cache
	fmt.Println("Clearing cache")
	err := os.RemoveAll(c.folder)
	if err != nil {
		log.Fatalf("Error clearing cache: %s", err)
	}
	err = os.MkdirAll(c.folder, os.ModePerm)
	if err != nil {
		log.Fatalf("Error creating cache folder: %s", err)
	}
}
