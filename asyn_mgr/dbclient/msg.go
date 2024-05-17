package dbclient

import (
	"px/proto/proto_db"
	"px/shared/asyn_mgr/asyn_msg"
)

type (
	ReqInf interface {
		asyn_msg.ReqInf
		SetBlockCh(chan asyn_msg.RespInf)
		GetBlockCh() chan asyn_msg.RespInf
	}
	ReqBase struct {
		asyn_msg.ReqBase
		ch chan asyn_msg.RespInf
	}
	RespBase struct {
		asyn_msg.RespBase
	}
	ReqDbQuery struct {
		ReqBase
		Sql  string
		Args []interface{}
		Op   proto_db.DB_OPERATOR
	}
	RespDbQuery struct {
		RespBase
		ErrMsg string
		Data   []*proto_db.DBData
	}
	ReqDbMulQuery struct {
		ReqBase
		Queries []*ReqDbQuery
	}
	RespDbMulQuery struct {
		RespBase
		ErrMsg string
		Data   [][]*proto_db.DBData
	}
)

func (this *ReqBase) SetBlockCh(ch chan asyn_msg.RespInf) {
	this.ch = ch
}

func (this *ReqBase) GetBlockCh() chan asyn_msg.RespInf {
	return this.ch
}

func (this *ReqBase) GetModuleId() asyn_msg.AsynModuleId {
	return asyn_msg.DbClientModuleId
}

func (this *RespBase) GetModuleId() asyn_msg.AsynModuleId {
	return asyn_msg.DbClientModuleId
}
