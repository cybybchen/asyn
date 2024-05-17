package db_pool

import (
	"px/framebase"
	"px/proto/proto_db"
	"px/shared/asyn_mgr/asyn_msg"
	"px/shared/asyn_mgr/db_pool/db"
	"px/utils"
	"strconv"

	"gitlab.sunborngame.com/base/log"
)

const (
	PoolMinSize = 1
	PoolMaxSize = 32
)

type DbPoolMgr struct {
	*asyn_msg.AsynBase

	dbs []*DB
}

func CreateDbPoolMgr(poolSize int, cfg *db.MysqlCfg) asyn_msg.AsynModInf {
	//做个池子大小限制
	if poolSize > PoolMaxSize {
		poolSize = PoolMaxSize
	}
	if poolSize <= 0 {
		poolSize = PoolMinSize
	}

	var dbPoolMgr = &DbPoolMgr{
		AsynBase: asyn_msg.NewAsynBase(),
	}
	dbPoolMgr.initDbs(poolSize, cfg)

	return dbPoolMgr
}

func (this *DbPoolMgr) initDbs(poolSize int, cfg *db.MysqlCfg) {
	for i := 0; i < poolSize; i++ {
		db := newDB(cfg, this.RespChan)
		this.dbs = append(this.dbs, db)
	}
}

func (this *DbPoolMgr) GetModuleId() asyn_msg.AsynModuleId {
	return asyn_msg.DbPoolModuleId
}

func (this *DbPoolMgr) ReqLen() int {
	return len(this.CallBacks)
}

func (this *DbPoolMgr) Init() {
	for _, db := range this.dbs {
		db.init()
	}
	go this.loop()
}

func (this *DbPoolMgr) Close() {
	for _, db := range this.dbs {
		db.close()
	}
}

func (this *DbPoolMgr) loop() {
	for {
		if !this.run() {
			break
		}
	}
}

func (this *DbPoolMgr) run() bool {
	defer utils.Recover(framebase.SendWeChatMsg(), framebase.IsReleaseEnv())

	select {
	case req, ok := <-this.ReqChan.C():
		if !ok {
			return false
		}
		this.handleReq(req)
	}

	return true
}

func (this *DbPoolMgr) handleReq(req asyn_msg.ReqInf) {
	switch msg := req.(type) {
	case *ReqDBOperator:
		this.handleReqDBOperator(msg)
	default:
		log.Error("reqMsg err %v", req)
	}
}

func (this *DbPoolMgr) handleReqDBOperator(req *ReqDBOperator) {
	if len(req.Args) == 0 {
		db := this.getDB(0)
		db.reqChan.Put(req)
	} else {
		argId, e := this.parseReqArgs(req.Args[0])
		if e != nil {
			log.Error("get db err, e=%v", e)
			var resp = &RespDBOperator{
				Err: e,
			}
			resp.SetFcId(req.GetFcId())
			this.RespChan.Put(resp)
			return
		}
		db := this.getDB(argId)
		db.reqChan.Put(req)
	}
}

func (this *DbPoolMgr) parseReqArgs(arg *proto_db.DBArgs) (int, error) {
	switch arg.GetArgsType() {
	case proto_db.DB_ARGS_TYPE_D_A_T_INT:
		dbArg, e := strconv.Atoi(string(arg.GetArgs()))
		if e != nil {
			log.Error("parseReqArgs, e=%v", e)
			return 0, e
		}
		return dbArg, nil
	case proto_db.DB_ARGS_TYPE_D_A_T_BIGINT:
		dbArg, e := strconv.Atoi(string(arg.GetArgs()))
		if e != nil {
			log.Error("parseReqArgs, e=%v", e)
			return 0, e
		}
		return dbArg, nil
	}

	return 0, nil
}

func (this *DbPoolMgr) getDB(argId int) *DB {
	if argId == 0 {
		return this.dbs[utils.RandInt(len(this.dbs))]
	}
	if argId < 0 {
		argId = -argId
	}
	return this.dbs[argId%len(this.dbs)]
}
