package redis_proxy

import (
	"encoding/base64"
	"flag"
	"fmt"
	"gitlab.sunborngame.com/base/log"
	"px/config"
	"px/framebase"
	"px/shared/asyn_mgr/asyn_msg"
	"px/shared/asyn_mgr/redis_proxy/client"
	"px/shared/asyn_mgr/redis_proxy/redis_inf"
	"px/utils"
	"reflect"
)

var (
	redisConf = flag.String("redis", "config/redis.xml", "redis config path")
)

const (
	LogTag = "[redis_proxy]"
)

type RedisProxyMgr struct {
	*asyn_msg.AsynBase
	client client.ClientInf
}

func CreateRedisProxy() *RedisProxyMgr {
	if err := config.LoadRedisConf(*redisConf); err != nil {
		panic(fmt.Errorf("load redis conf failed: %s", err.Error()))
	}
	var addrs = config.GetRedisCfg().Addrs()
	if len(addrs) == 0 {
		log.Panic("create redis proxy failed, addrs is empty")
	}

	var redisProxyMgr = &RedisProxyMgr{
		AsynBase: asyn_msg.NewAsynBase(),
	}
	if len(addrs) == 1 {
		redisProxyMgr.client = client.NewClient(addrs[0])
	} else {
		redisProxyMgr.client = client.NewClientCluster(addrs)
	}

	return redisProxyMgr
}

func (this *RedisProxyMgr) Start() {

}

func (this *RedisProxyMgr) GetModuleId() asyn_msg.AsynModuleId {
	return asyn_msg.RedisModuleId
}

func (this *RedisProxyMgr) ReqLen() int {
	return len(this.CallBacks)
}

func (this *RedisProxyMgr) Close() {
	this.client.Close()
}

func (this *RedisProxyMgr) Init() {
	go this.loop()
}

func (this *RedisProxyMgr) loop() {
	for {
		if !this.run() {
			break
		}
	}
}

func (this *RedisProxyMgr) run() bool {
	defer utils.Recover(framebase.SendWeChatMsg(), framebase.IsReleaseEnv())

	select {
	case req, ok := <-this.ReqChan.C():
		if !ok {
			return false
		}
		this.handleReq(req)
	}

	return true
}

func (this *RedisProxyMgr) handleReq(req asyn_msg.ReqInf) {
	log.Info("%s redis recv req:%v", LogTag, req)
	switch msg := req.(type) {
	case *ReqSet:
		this.handleSet(msg)
	case *ReqGet:
		this.handleGet(msg)
	case *ReqHSet:
		this.handleHSet(msg)
	case *ReqHGet:
		this.handleHGet(msg)
	case *ReqHDel:
		this.handleHDel(msg)
	case *ReqDel:
		this.handleDel(msg)
	case *ReqTtl:
		this.handleTtl(msg)
	default:
		log.Error("reqMsg err %v", req)
	}
}

func (this *RedisProxyMgr) handleSet(req *ReqSet) {
	var resp = &RespSet{}
	resp.SetFcId(req.GetFcId())
	defer this.RespChan.Put(resp)

	var redisData = this.encode(req.Value)
	err := this.client.Set(req.Key, redisData, req.Ttl)
	if err != nil {
		log.Error("%s set failed, Key: %s, Value: %s, error: %v", LogTag, req.Key, req.Value, err)
		resp.Err = err.Error()
	}
}

func (this *RedisProxyMgr) handleGet(req *ReqGet) {
	value, err := this.client.Get(req.Key)
	var resp = &RespGet{}
	if err != nil {
		resp.Err = err.Error()
	} else {
		var ret = this.decode(value)
		if ret == nil {
			log.Error("%s get failed, Key: %s, Value: %s, error: %v", LogTag, req.Key, value, err)
		}
		resp.Ret = ret
	}

	resp.SetFcId(req.GetFcId())
	this.RespChan.Put(resp)
}

func (this *RedisProxyMgr) handleHSet(req *ReqHSet) {
	var resp = &RespHSet{}
	resp.SetFcId(req.GetFcId())
	defer this.RespChan.Put(resp)

	var redisData = this.encode(req.Value)
	err := this.client.HSet(req.Key1, req.Key2, redisData)
	if err != nil {
		log.Error("%s hset failed, Key1: %s, key2: %s, Value: %s, error: %v", LogTag, req.Key1, req.Key2, req.Value, err)
		resp.Err = err.Error()
	}
}

func (this *RedisProxyMgr) handleHGet(req *ReqHGet) {
	var resp = &RespHGet{}
	resp.SetFcId(req.GetFcId())
	defer this.RespChan.Put(resp)

	value, err := this.client.HGet(req.Key1, req.Key2)
	if err != nil {
		resp.Err = err.Error()
	} else {
		var ret = this.decode(value)
		if ret == nil {
			log.Error("%s hget failed, Key1: %s, Key2: %s, value: %s, error: %v", LogTag, req.Key1, req.Key2, value, err)
		}
		resp.Ret = ret
	}
}

func (this *RedisProxyMgr) handleHDel(req *ReqHDel) {
	var resp = &RespHDel{}
	resp.SetFcId(req.GetFcId())
	defer this.RespChan.Put(resp)

	err := this.client.HDel(req.Key1, req.Key2...)
	if err != nil {
		log.Error("%s hdel failed, Key1: %s, key2: %v, error: %v", LogTag, req.Key1, req.Key2, err)
		resp.Err = err.Error()
	}
}

func (this *RedisProxyMgr) handleTtl(req *ReqTtl) {
	var resp = &RespTtl{}
	resp.SetFcId(req.GetFcId())
	defer this.RespChan.Put(resp)

	err := this.client.Ttl(req.Key, req.Ttl)
	if err != nil {
		log.Error("%s ttl failed, Key: %s, ttl: %v error: %v", LogTag, req.Key, req.Ttl, err)
		resp.Err = err.Error()
	}
}

func (this *RedisProxyMgr) handleDel(req *ReqDel) {
	var resp = &RespDel{}
	resp.SetFcId(req.GetFcId())
	defer this.RespChan.Put(resp)

	err := this.client.Del(req.Key)
	if err != nil {
		log.Error("%s del failed, Key: %s, error: %v", LogTag, req.Key, err)
		resp.Err = err.Error()
	}
}

func (this *RedisProxyMgr) encode(value interface{}) *redis_inf.RedisData {
	var tp = redis_inf.VTypeNone
	var tpName string
	switch v := value.(type) {
	case string:
		tp = redis_inf.VTypeString
	case int32, uint32:
		tp = redis_inf.VTypeInt32
	case int64, uint64:
		tp = redis_inf.VTypeInt64
	case []byte:
		tp = redis_inf.VTypeBytes
	case redis_inf.RedisDataInf:
		tp = redis_inf.VTypeData
		refTp := reflect.TypeOf(v).Elem()
		tpName = refTp.Name()
	}

	return &redis_inf.RedisData{
		Head: &redis_inf.RedisDataHead{
			Tp:     tp,
			TpName: tpName,
		},
		Body: value,
	}
}

func (this *RedisProxyMgr) decode(value string) any {
	var rs = &redis_inf.RedisData{
		Head: &redis_inf.RedisDataHead{},
	}
	err := rs.UnmarshalBinary([]byte(value))
	if err != nil {
		log.Error("%s redisData unmarshal failed, err:%v", LogTag, err)
		return nil
	}

	switch rs.Head.Tp {
	case redis_inf.VTypeString:
		return rs.Body
	case redis_inf.VTypeInt32:
		return int32(rs.Body.(float64))
	case redis_inf.VTypeUInt32:
		return uint32(rs.Body.(float64))
	case redis_inf.VTypeInt64:
		return int64(rs.Body.(float64))
	case redis_inf.VTypeUInt64:
		return uint64(rs.Body.(float64))
	case redis_inf.VTypeBytes:
		data, err := base64.StdEncoding.DecodeString(rs.Body.(string))
		if err != nil {
			log.Error("%s base64 decode failed, err:%v", LogTag, err)
			return nil
		}
		rs.Body = data
	case redis_inf.VTypeData:
		rs.Body = redis_inf.CreateMsg(rs.Head.TpName)
		err = rs.UnmarshalBinary([]byte(value))
		if err != nil {
			log.Error("%s redisData unmarshal failed, err:%v", LogTag, err)
			return nil
		}
		return rs.Body
	}

	return nil
}
