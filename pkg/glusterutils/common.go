package glusterutils

import (
	"sync"
	"time"

	"github.com/gluster/glusterd2/pkg/restclient"
)

func initRESTClient(config *Config) (*restclient.Client, error) {
	client, err := restclient.New(
		config.Glusterd2Endpoint,
		config.Glusterd2User,
		config.Glusterd2Secret,
		config.Glusterd2Cacert,
		config.Glusterd2Insecure,
	)
	if err != nil {
		return nil, err
	}
	client.SetTimeout(time.Duration(config.Timeout) * time.Second)
	return client, nil
}

func setDefaultConfig(config *Config) {
	if config.Timeout == 0 {
		config.Timeout = 30
	}
	if config.GlusterMgmt == "" {
		config.GlusterMgmt = "glusterd"
	}
	if config.GlusterCmd == "" {
		config.GlusterCmd = "gluster"
	}
	if config.Glusterd2Endpoint == "" {
		config.Glusterd2Endpoint = "http://localhost:24007"
	}
}

type cache struct {
	key       string
	data      interface{}
	ttl       float64
	createdAt time.Time
}

func (c *cache) isExpired() bool {
	duration := time.Since(c.createdAt)
	return duration.Seconds() >= c.ttl
}

var caches = make(map[string]cache)
var cacheLock = &sync.Mutex{}

func cacheSet(key string, data interface{}, ttl int) {
	caches[key] = cache{
		key:       key,
		data:      data,
		ttl:       float64(ttl),
		createdAt: time.Now(),
	}
}

func cacheGet(key string) interface{} {
	c, exists := caches[key]
	if !exists {
		return nil
	}
	if c.isExpired() {
		return nil
	}
	return c.data
}

func cacheCleanup() {
	var toDel []string
	for k, v := range caches {
		if v.isExpired() {
			toDel = append(toDel, k)
		}
	}

	for _, k := range toDel {
		delete(caches, k)
	}
}

// CachedOutput caches the result of given function
func CachedOutput(key string, fn func() (interface{}, error), ttl int) (interface{}, error) {
	cacheLock.Lock()
	defer cacheLock.Unlock()
	cacheCleanup()
	data := cacheGet(key)
	if data == nil {
		newdata, err := fn()
		if err != nil {
			return nil, err
		}
		cacheSet(key, newdata, ttl)
		return newdata, nil
	}
	return data, nil
}
