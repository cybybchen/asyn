package db_op

import (
	"fmt"
	"px/proto/proto_db"
	"px/shared/asyn_mgr"
	"px/shared/asyn_mgr/asyn_msg"
	"px/shared/asyn_mgr/dbclient"
	"reflect"
)

func DoDbInsert(sql string, args []interface{}, cb func(string)) {
	var req = &dbclient.ReqDbQuery{
		Sql:  sql,
		Args: args,
		Op:   proto_db.DB_OPERATOR_DB_OP_INSERT,
	}
	asyn_mgr.GetAsynMgr().SendReq(req, func(inf asyn_msg.RespInf) asyn_msg.AsynCBPtr {
		resp, ok := inf.(*dbclient.RespDbQuery)
		if !ok {
			var err = fmt.Errorf("resp type error, resp=%v", inf)
			cb(err.Error())
			return 0
		}

		cb(resp.ErrMsg)

		return asyn_msg.AsynCBPtr(reflect.ValueOf(cb).Pointer())
	})
}

func DoDbSelect(sql string, args []interface{}, cb func(string, []*proto_db.DBData)) {
	var req = &dbclient.ReqDbQuery{
		Sql:  sql,
		Args: args,
		Op:   proto_db.DB_OPERATOR_DB_OP_SELECT,
	}
	asyn_mgr.GetAsynMgr().SendReq(req, func(inf asyn_msg.RespInf) asyn_msg.AsynCBPtr {
		resp, ok := inf.(*dbclient.RespDbQuery)
		if !ok {
			var err = fmt.Errorf("resp type error, resp=%v", inf)
			cb(err.Error(), nil)
			return 0
		}

		cb(resp.ErrMsg, resp.Data)

		return asyn_msg.AsynCBPtr(reflect.ValueOf(cb).Pointer())
	})
}

func DoDbUpdate(sql string, args []interface{}, cb func(string)) {
	var req = &dbclient.ReqDbQuery{
		Sql:  sql,
		Args: args,
		Op:   proto_db.DB_OPERATOR_DB_OP_UPDATE,
	}
	asyn_mgr.GetAsynMgr().SendReq(req, func(inf asyn_msg.RespInf) asyn_msg.AsynCBPtr {
		resp, ok := inf.(*dbclient.RespDbQuery)
		if !ok {
			var err = fmt.Errorf("resp type error, resp=%v", inf)
			cb(err.Error())
			return 0
		}

		cb(resp.ErrMsg)

		return asyn_msg.AsynCBPtr(reflect.ValueOf(cb).Pointer())
	})
}

func DoDbDelete(sql string, args []interface{}, cb func(string)) {
	var req = &dbclient.ReqDbQuery{
		Sql:  sql,
		Args: args,
		Op:   proto_db.DB_OPERATOR_DB_OP_DELETE,
	}
	asyn_mgr.GetAsynMgr().SendReq(req, func(inf asyn_msg.RespInf) asyn_msg.AsynCBPtr {
		resp, ok := inf.(*dbclient.RespDbQuery)
		if !ok {
			var err = fmt.Errorf("resp type error, resp=%v", inf)
			cb(err.Error())
			return 0
		}

		cb(resp.ErrMsg)

		return asyn_msg.AsynCBPtr(reflect.ValueOf(cb).Pointer())
	})
}

func DoDbReplace(sql string, args []interface{}, cb func(string)) {
	var req = &dbclient.ReqDbQuery{
		Sql:  sql,
		Args: args,
		Op:   proto_db.DB_OPERATOR_DB_OP_REPLACE,
	}
	asyn_mgr.GetAsynMgr().SendReq(req, func(inf asyn_msg.RespInf) asyn_msg.AsynCBPtr {
		resp, ok := inf.(*dbclient.RespDbQuery)
		if !ok {
			var err = fmt.Errorf("resp type error, resp=%v", inf)
			cb(err.Error())
			return 0
		}

		cb(resp.ErrMsg)

		return asyn_msg.AsynCBPtr(reflect.ValueOf(cb).Pointer())
	})
}

func DoDbQuery(sql string, args []interface{}, op proto_db.DB_OPERATOR, cb func(string, []*proto_db.DBData)) {
	var req = &dbclient.ReqDbQuery{
		Sql:  sql,
		Args: args,
		Op:   op,
	}
	asyn_mgr.GetAsynMgr().SendReq(req, func(inf asyn_msg.RespInf) asyn_msg.AsynCBPtr {
		resp, ok := inf.(*dbclient.RespDbQuery)
		if !ok {
			var err = fmt.Errorf("resp type error, resp=%v", inf)
			cb(err.Error(), nil)
			return 0
		}

		cb(resp.ErrMsg, resp.Data)

		return asyn_msg.AsynCBPtr(reflect.ValueOf(cb).Pointer())
	})
}

func DoDbMultiSelect(sqlMultiQuery *SqlMultiQuery, cb func(string, [][]*proto_db.DBData)) {
	var req = &dbclient.ReqDbMulQuery{
		Queries: make([]*dbclient.ReqDbQuery, 0, len(sqlMultiQuery.SqlQuery)),
	}
	for _, sqlQuery := range sqlMultiQuery.SqlQuery {
		req.Queries = append(req.Queries, &dbclient.ReqDbQuery{
			Sql:  sqlQuery.Sql,
			Args: sqlQuery.Args,
			Op:   proto_db.DB_OPERATOR_DB_OP_SELECT,
		})
	}
	asyn_mgr.GetAsynMgr().SendReq(req, func(inf asyn_msg.RespInf) asyn_msg.AsynCBPtr {
		resp, ok := inf.(*dbclient.RespDbMulQuery)
		if !ok {
			var err = fmt.Errorf("resp type error, resp=%v", inf)
			cb(err.Error(), nil)
			return 0
		}

		cb(resp.ErrMsg, resp.Data)

		return asyn_msg.AsynCBPtr(reflect.ValueOf(cb).Pointer())
	})
}

// 阻塞调用，慎重使用，主要用于服务器启动的时候使用，运行期间禁止使用
func DoDbBlockQuery(sql string, args []interface{}, op proto_db.DB_OPERATOR) (errMsg string, datas []*proto_db.DBData) {
	var req = &dbclient.ReqDbQuery{
		Sql:  sql,
		Args: args,
		Op:   op,
	}

	var ch = make(chan asyn_msg.RespInf)
	req.SetBlockCh(ch)
	asyn_mgr.GetAsynMgr().SendReq(req, nil)
	respInf := <-ch
	resp, ok := respInf.(*dbclient.RespDbQuery)
	if !ok {
		var err = fmt.Errorf("resp type error, resp=%v", respInf)
		errMsg = err.Error()
		return
	}
	errMsg = resp.ErrMsg
	datas = resp.Data
	return
}
