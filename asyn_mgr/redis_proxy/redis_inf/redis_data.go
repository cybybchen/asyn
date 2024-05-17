package redis_inf

import "encoding/json"

type RedisValueType int32

const (
	VTypeNone RedisValueType = iota
	VTypeInt32
	VTypeUInt32
	VTypeInt64
	VTypeUInt64
	VTypeString
	VTypeBytes
	VTypeData
)

type (
	RedisData struct {
		Head *RedisDataHead
		Body interface{}
	}
)

type RedisDataHead struct {
	Tp     RedisValueType
	TpName string
}

func (this *RedisData) MarshalBinary() (data []byte, err error) {
	return json.Marshal(this)
}

func (this *RedisData) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, this)
}
