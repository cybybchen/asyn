package redis_proxy

import (
	"px/shared/asyn_mgr/asyn_msg"
	"time"
)

type ReqBase struct {
	asyn_msg.ReqBase
	Ttl time.Duration
}

type RespBase struct {
	asyn_msg.RespBase
}

type ReqSet struct {
	ReqBase
	Key   string
	Value any
}

type RespSet struct {
	RespBase
	Err string
}

type ReqGet struct {
	ReqBase
	Key string
}

type RespGet struct {
	RespBase
	Err string
	Ret any
}

type (
	ReqHSet struct {
		ReqBase
		Key1  string
		Key2  string
		Value any
	}
	RespHSet struct {
		RespBase
		Err string
	}
	ReqHGet struct {
		ReqBase
		Key1 string
		Key2 string
	}
	RespHGet struct {
		RespBase
		Ret any
		Err string
	}
	ReqHDel struct {
		ReqBase
		Key1 string
		Key2 []string
	}
	RespHDel struct {
		RespBase
		Err string
	}
	ReqTtl struct {
		ReqBase
		Key string
	}
	RespTtl struct {
		RespBase
		Err string
	}
	ReqDel struct {
		ReqBase
		Key string
	}
	RespDel struct {
		RespBase
		Err string
	}
)

func (this *ReqBase) GetModuleId() asyn_msg.AsynModuleId {
	return asyn_msg.RedisModuleId
}

func (this *RespBase) GetModuleId() asyn_msg.AsynModuleId {
	return asyn_msg.RedisModuleId
}
