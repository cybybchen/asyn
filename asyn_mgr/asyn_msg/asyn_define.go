package asyn_msg

type AsynModuleId int32

const (
	AStarModuleId    AsynModuleId = 1
	DbClientModuleId AsynModuleId = 2
	DbPoolModuleId   AsynModuleId = 3
	HttpSvrModuleId  AsynModuleId = 4
	AoiModuleId      AsynModuleId = 5
	EtcdModuleId     AsynModuleId = 6
	DsaModuleId      AsynModuleId = 7
	RedisModuleId    AsynModuleId = 8
)

const (
	MessageChanCap = 1024
)
