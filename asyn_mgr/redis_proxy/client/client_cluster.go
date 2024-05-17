package client

import (
	"context"
	"github.com/redis/go-redis/v9"
	"gitlab.sunborngame.com/base/log"
	"px/shared/asyn_mgr/redis_proxy/redis_inf"
	"time"
)

const (
	ClientClusterLogTag = "[client_cluster]"
)

type ClientCluster struct {
	rClient *redis.ClusterClient
}

func NewClientCluster(addrs []string) *ClientCluster {
	rClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: addrs,
	})

	_, err := rClient.Ping(context.Background()).Result()
	if err != nil {
		log.Panic("connect redis err:%v", err)
	}
	log.Info("connect redis success")

	return &ClientCluster{rClient: rClient}
}

func (this *ClientCluster) Close() {

}

func (this *ClientCluster) Set(key string, value *redis_inf.RedisData, ttl time.Duration) error {
	// 设置值
	err := this.rClient.Set(context.Background(), key, value, ttl).Err()
	if err != nil {
		log.Error("%v set key=%s, value=%s, err:", ClientClusterLogTag, key, value, err)
		return err
	}

	return nil
}

func (this *ClientCluster) Get(key string) (string, error) {
	value, err := this.rClient.Get(context.Background(), key).Result()
	if err != nil {
		log.Error("%v redis set err:%v", ClientLogTag, err)
		return "", err
	}

	return value, nil
}

func (this *ClientCluster) HSet(key1 string, key2 string, value *redis_inf.RedisData) error {
	err := this.rClient.HSet(context.Background(), key1, key2, value).Err()
	if err != nil {
		log.Error("%v redis hset err:%v", ClientLogTag, err)
		return err
	}

	return nil
}

func (this *ClientCluster) HGet(key1 string, key2 string) (string, error) {
	value, err := this.rClient.HGet(context.Background(), key1, key2).Result()
	if err != nil {
		log.Error("%v redis hget err:%v", ClientLogTag, err)
		return "", err
	}

	return value, nil
}

func (this *ClientCluster) HDel(key1 string, key2 ...string) error {
	err := this.rClient.HDel(context.Background(), key1, key2...).Err()
	if err != nil {
		log.Error("%v redis hdel err:%v", ClientLogTag, err)
	}
	return err
}

func (this *ClientCluster) Ttl(key string, ttl time.Duration) error {
	err := this.rClient.Expire(context.Background(), key, ttl).Err()
	if err != nil {
		log.Error("%v redis ttl err:%v", ClientLogTag, err)
	}

	return err
}

func (this *ClientCluster) Del(key string) error {
	err := this.rClient.Del(context.Background(), key).Err()
	if err != nil {
		log.Error("%v redis del err:%v", ClientLogTag, err)
	}
	return err
}
