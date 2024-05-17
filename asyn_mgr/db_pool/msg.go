package db_pool

import (
	"px/proto/proto_db"
	"px/shared/asyn_mgr/asyn_msg"
)

type (
	ReqBase struct {
		asyn_msg.ReqBase
	}
	RespBase struct {
		asyn_msg.RespBase
	}
	ReqDBOperator struct {
		ReqBase
		Sql  string
		Op   proto_db.DB_OPERATOR
		Args []*proto_db.DBArgs
	}
	RespDBOperator struct {
		RespBase
		Ret []map[string][]byte
		Err error
	}
)

func (this *ReqBase) GetModuleId() asyn_msg.AsynModuleId {
	return asyn_msg.DbPoolModuleId
}

func (this *RespBase) GetModuleId() asyn_msg.AsynModuleId {
	return asyn_msg.DbPoolModuleId
}
