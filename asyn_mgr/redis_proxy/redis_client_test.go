package redis_proxy

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"px/shared/asyn_mgr/redis_proxy/redis_inf"
	"reflect"
	"testing"
)

var addr = "192.168.18.163:6379"

func TestRedisClient(t *testing.T) {
	var client = redis.NewClient(&redis.Options{
		Addr: addr,
	})

	defer client.Close()

	err := client.Set(context.Background(), "44", &TestRedis{
		Key:    1,
		Value:  string([]byte{1, 2}),
		Value2: 22,
		Value3: []byte{1, 2},
	}, 0).Err()
	if err != nil {
		t.Error(err)
	}

	value, _ := client.Get(context.Background(), "44").Result()
	var rt = &TestRedis{}
	rt.UnmarshalBinary([]byte(value))
	fmt.Println(rt)

	client.HSet(context.Background(), "11", "22", &TestRedis{
		Key:    1,
		Value:  "11",
		Value2: 22,
	})

	var testRedis = &TestRedis{
		Key:    1,
		Value:  "11",
		Value2: 22,
	}
	tp1 := reflect.TypeOf(testRedis)
	tp := tp1.Elem()
	fmt.Println(tp.Name())
	a := reflect.New(tp)
	b := a.Interface().(redis_inf.RedisDataInf)
	fmt.Println(b)
	tp.Kind()

	rt = &TestRedis{
		Key:    1,
		Value:  string([]byte{1, 2}),
		Value2: 22,
		Value3: []byte{1, 2},
	}
	err = client.Set(context.Background(), "5", &redis_inf.RedisData{
		Head: &redis_inf.RedisDataHead{
			Tp:     redis_inf.VTypeData,
			TpName: tp.Name()},
		Body: rt}, 0).Err()
	if err != nil {
		t.Error(err)
	}

	value, _ = client.Get(context.Background(), "5").Result()
	var rs = &redis_inf.RedisData{
		Head: &redis_inf.RedisDataHead{},
	}
	err = rs.UnmarshalBinary([]byte(value))
	if err != nil {
		t.Error(err)
	}
	fmt.Println(rs)

	rs.Body = &TestRedis{}
	err = rs.UnmarshalBinary([]byte(value))
	if err != nil {
		t.Error(err)
	}
	fmt.Println(rs)
}

func TestRedisString(t *testing.T) {
	var client = redis.NewClient(&redis.Options{
		Addr: addr,
	})
	defer client.Close()

	err := client.Set(context.Background(), "8", &redis_inf.RedisData{
		Head: &redis_inf.RedisDataHead{
			Tp:     redis_inf.VTypeData,
			TpName: "",
		}, Body: []byte{1, 2, 3}}, 0).Err()
	if err != nil {
		t.Error(err)
	}

	value, _ := client.Get(context.Background(), "8").Bytes()
	var rs = &redis_inf.RedisData{
		Head: &redis_inf.RedisDataHead{},
	}
	err = rs.UnmarshalBinary(value)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(rs)
	var body = []byte(rs.Body.(string))
	fmt.Println(body)

	var rd = &redis_inf.RedisData{
		Head: &redis_inf.RedisDataHead{
			Tp: redis_inf.VTypeData,
		},
		Body: []byte{1, 2, 3},
	}
	fmt.Println(rd)
	j, _ := json.Marshal(rd)
	fmt.Println(string(j))

	var rd1 = &redis_inf.RedisData{
		Head: &redis_inf.RedisDataHead{},
	}
	json.Unmarshal(j, rd1)
	fmt.Println(rd1)
	fmt.Println(rd1.Body)

	// 将Base64编码的字符串解码为[]byte类型
	data, err := base64.StdEncoding.DecodeString(rd1.Body.(string))
	if err != nil {
		fmt.Println("解码失败:", err)
		return
	}

	fmt.Println("解码后的[]byte数据:")
	fmt.Printf("%v\n", data)
}

func TestInt(t *testing.T) {
	var client = redis.NewClient(&redis.Options{
		Addr: addr,
	})
	defer client.Close()

	err := client.Set(context.Background(), "9", &redis_inf.RedisData{
		Head: &redis_inf.RedisDataHead{
			Tp:     redis_inf.VTypeData,
			TpName: "",
		}, Body: 9}, 0).Err()
	if err != nil {
		t.Error(err)
	}

	value, _ := client.Get(context.Background(), "9").Bytes()
	var rs = &redis_inf.RedisData{
		Head: &redis_inf.RedisDataHead{},
	}
	err = rs.UnmarshalBinary(value)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(rs)
	rs.Body = int32(rs.Body.(float64))
}
