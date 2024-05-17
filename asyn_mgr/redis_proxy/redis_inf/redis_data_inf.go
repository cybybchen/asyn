package redis_inf

type RedisDataInf interface {
	MarshalBinary() (data []byte, err error)
	UnmarshalBinary(data []byte) error
}
