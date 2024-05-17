package client

import (
	"px/shared/asyn_mgr/redis_proxy/redis_inf"
	"time"
)

type ClientInf interface {
	Close()
	Set(key string, value *redis_inf.RedisData, ttl time.Duration) error
	Get(key string) (string, error)
	HSet(key1, key2 string, value *redis_inf.RedisData) error
	HGet(key1 string, key2 string) (string, error)
	HDel(key1 string, key2 ...string) error
	Ttl(key string, ttl time.Duration) error
	Del(key string) error
}
