package dbclient

import (
	"bytes"
	"px/define"
	"px/framebase"
	"px/proto/proto_db"
	"px/shared/asyn_mgr/asyn_msg"
	"px/shared/time_wheel"
	"px/utils"
	"px/utils/cbctx"

	"gitlab.sunborngame.com/base/log"
)

const (
	DbTick      = 5 * define.Second
	DbOpTimeout = 120 * define.Second
)

type DbMgr struct {
	*asyn_msg.AsynBase

	client *DBClient

	timer *time_wheel.TimeWheelS
}

func CreateDbMgr(reconn int, addr string) asyn_msg.AsynModInf {
	var dbMgr = &DbMgr{
		AsynBase: asyn_msg.NewAsynBase(),
		timer:    time_wheel.NewTimeWheelSDefault(),
	}
	dbMgr.client = newDbClient(reconn, addr)

	return dbMgr
}

func (this *DbMgr) GetModuleId() asyn_msg.AsynModuleId {
	return asyn_msg.DbClientModuleId
}

func (this *DbMgr) ReqLen() int {
	return len(this.CallBacks)
}

func (this *DbMgr) Init() {
	this.client.Start(defaultConnTimeout)
	this.timer.Start()

	go this.loop()

	this.timer.AddRepeatTimer(DbTick, this.tick)
}

func (this *DbMgr) Close() {

}

func (this *DbMgr) tick(now int64, _ cbctx.Ctx) {
	this.client.tick(now)
}

func (this *DbMgr) loop() {
	defer utils.Recover(framebase.SendWeChatMsg(), framebase.IsReleaseEnv())

	select { // first wait connect to dbproxy
	case resp := <-this.client.respChan.C():
		this.client.handleResp(resp)
	}
	for {
		if !this.run() {
			break
		}
	}
}

func (this *DbMgr) run() bool {
	defer utils.Recover(framebase.SendWeChatMsg(), framebase.IsReleaseEnv())

	for {
		select {
		case task, ok := <-this.timer.NotifyChannel.C():
			if !ok {
				return false
			}
			this.timer.TriggerTimerCb(task)
		case req, ok := <-this.ReqChan.C():
			if !ok {
				return false
			}
			this.handleReq(req)
		case resp := <-this.client.respChan.C():
			this.client.handleResp(resp)
		}
	}

	return true
}

func (this *DbMgr) handleReq(req asyn_msg.ReqInf) {
	log.Info("[DL_DBOperator] handleReq, fcId=%d", req.GetFcId())
	switch msg := req.(type) {
	case *ReqDbQuery:
		dbArgs, e := this.client.toDbArgs(msg.Args)
		if e != nil {
			log.Error("todbargs err, e=%v", e)
			var resp = &RespDbQuery{
				Data:   nil,
				ErrMsg: e.Error(),
			}
			resp.SetFcId(req.GetFcId())
			this.putResp(msg, resp)
			return
		}
		this.client.doDbQuery(msg.Sql, msg.Op, dbArgs, func(errMsg string, data []*proto_db.DBData) {
			var resp = &RespDbQuery{
				Data:   data,
				ErrMsg: errMsg,
			}
			resp.SetFcId(req.GetFcId())
			this.putResp(msg, resp)
		})
	case *ReqDbMulQuery:
		var callbackTimes = 0
		var resp = &RespDbMulQuery{}
		resp.SetFcId(req.GetFcId())
		var buff bytes.Buffer
		var endFunc = func() {
			callbackTimes++
			if callbackTimes == len(msg.Queries) {
				resp.ErrMsg = buff.String()
				this.putResp(msg, resp)
			}
		}
		for _, r := range msg.Queries {
			dbArgs, e := this.client.toDbArgs(r.Args)
			if e != nil {
				log.Error("todbargs err, e=%v", e)
				buff.WriteString("|| ")
				buff.WriteString(e.Error())
				endFunc()
				continue
			}
			this.client.doDbQuery(r.Sql, r.Op, dbArgs, func(errMsg string, data []*proto_db.DBData) {
				if errMsg != "" {
					buff.WriteString("|| ")
					buff.WriteString(errMsg)
				} else {
					resp.Data = append(resp.Data, data)
				}

				endFunc()
			})
		}
	}
}

func (this *DbMgr) putResp(req ReqInf, resp asyn_msg.RespInf) {
	var blockCh = req.GetBlockCh()
	if blockCh != nil {
		blockCh <- resp
	} else {
		this.RespChan.Put(resp)
	}
}
