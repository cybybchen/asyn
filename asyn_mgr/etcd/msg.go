package etcd

import (
	"go.etcd.io/etcd/api/v3/mvccpb"
	"px/shared/asyn_mgr/asyn_msg"
)

type EtcdOpData struct {
	Op    mvccpb.Event_EventType
	Key   string
	Value any
}

type (
	ReqBase struct {
		asyn_msg.ReqBase
	}
	RespBase struct {
		asyn_msg.RespBase
	}
	ReqWatch struct {
		ReqBase
		Key        string
		WithPrefix bool
	}
	RespWatch struct {
		RespBase
		Datas []*EtcdOpData
	}
	ReqWrite struct {
		ReqBase
		Key   string
		Value any
	}
	RespWrite struct {
		RespBase
		Err string
	}
)

func (this *ReqBase) GetModuleId() asyn_msg.AsynModuleId {
	return asyn_msg.EtcdModuleId
}

func (this *RespBase) GetModuleId() asyn_msg.AsynModuleId {
	return asyn_msg.EtcdModuleId
}
