package redis_inf

import (
	"encoding/json"
	"strconv"
)

const (
	UserDataKeyPrefix = "user_data:"
	HomeKey           = "home"
	GateKey           = "gate"
)

// 接口RedisDataInf，存入redis
type RedisDataUserLoginHome struct {
	UserId uint64
	HomeId int32
	GateId int32
}

func (this *RedisDataUserLoginHome) MarshalBinary() (data []byte, err error) {
	return json.Marshal(this)
}

func (this *RedisDataUserLoginHome) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, this)
}

func init() {
	RegisterMsgCreate(&RedisDataUserLoginHome{})
}

func GenUserDataRedisKey(userId uint64) string {
	return UserDataKeyPrefix + strconv.FormatUint(userId, 10)
}
