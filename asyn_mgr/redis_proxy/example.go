package redis_proxy

import (
	"encoding/json"
	"px/shared/asyn_mgr/redis_proxy/redis_inf"
)

type TestRedis struct {
	Key    int
	Value  string
	Value2 int
	Value3 []byte
}

func init() {
	redis_inf.RegisterMsgCreate(&TestRedis{})
}

func (this *TestRedis) MarshalBinary() (data []byte, err error) {
	return json.Marshal(this)
}

func (this *TestRedis) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, this)
}
