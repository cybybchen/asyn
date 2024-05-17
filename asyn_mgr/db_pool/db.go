package db_pool

import (
	"gitlab.sunborngame.com/base/log"
	"px/framebase"
	"px/shared/asyn_mgr/asyn_msg"
	"px/shared/asyn_mgr/db_pool/db"
	"px/utils"
	"px/utils/chanx"
)

type DB struct {
	dbInter  db.DBInter
	reqChan  *chanx.UnboundedChan[asyn_msg.ReqInf]
	respChan *chanx.UnboundedChan[asyn_msg.RespInf]
}

func newDB(cfg *db.MysqlCfg, respChan *chanx.UnboundedChan[asyn_msg.RespInf]) *DB {
	db, err := db.NewMysql(cfg)
	if err != nil {
		panic(err)
	}

	return &DB{
		dbInter:  db,
		reqChan:  chanx.NewUnboundedChan[asyn_msg.ReqInf](MessageChanCap),
		respChan: respChan,
	}
}

func (this *DB) init() {
	go this.loop()
}

func (this *DB) close() {
	this.dbInter.CloseLazy()
}

func (this *DB) loop() {
	for {
		if !this.run() {
			break
		}
	}
}

func (this *DB) run() bool {
	defer utils.Recover(framebase.SendWeChatMsg(), framebase.IsReleaseEnv())

	select {
	case req, ok := <-this.reqChan.C():
		if !ok {
			return false
		}
		this.handleReq(req)
	}

	return true
}

func (this *DB) handleReq(req asyn_msg.ReqInf) {
	switch msg := req.(type) {
	case *ReqDBOperator:
		this.handleReqDBOperator(msg)
	default:
		log.Error("reqMsg err %v", req)
	}
}

func (this *DB) handleReqDBOperator(req *ReqDBOperator) {
	var resp = &RespDBOperator{}
	resp.SetFcId(req.GetFcId())
	resp.Ret, resp.Err = this.dbInter.DBOperator(req.Sql, req.Op, req.Args)
	this.respChan.Put(resp)
}
