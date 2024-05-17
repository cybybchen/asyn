package etcd

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"strings"
	"testing"
	"time"
)

func TestFuncEtcd(t *testing.T) {
	var key = "/cyb/cyb"
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"192.168.18.136:2379"},
		DialTimeout: time.Second * 5,
	})

	kv := clientv3.NewKV(cli)
	_, err = kv.Put(context.TODO(), key, "123")
	if err != nil {
		return
	}
	res, err := kv.Get(context.TODO(), key)
	for _, kv := range res.Kvs {
		fmt.Println(kv)
	}
	ch := cli.Watch(context.TODO(), key, clientv3.WithPrefix())
	for {
		select {
		case c := <-ch:
			for _, e := range c.Events {
				switch e.Type {
				case clientv3.EventTypePut:
					fmt.Println(string(e.Kv.Key))
					fmt.Println(string(e.Kv.Value))
				case clientv3.EventTypeDelete:
					ret := strings.Replace(string(e.Kv.Key), key, "", -1)
					fmt.Println(ret)
					//keyArray := strings.Split(string(e.Kv.Key), "/")
					//if len(keyArray) <= 0 {
					//	log.Error("[EventTypeDelete] key Split err : %s", err.Error())
					//	return
					//}
					//nodeId, _ := strconv.Atoi(keyArray[len(keyArray)-1])
					//fmt.Println(uint16(nodeId))
				}
			}
		}
	}
}
