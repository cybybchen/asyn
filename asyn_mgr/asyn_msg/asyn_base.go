package asyn_msg

import (
	"gitlab.sunborngame.com/base/log"
	"math"
	"px/utils/chanx"
	"reflect"
)

type AsynBase struct {
	FcId uint64

	CallBacks map[uint64]AsynCallback

	ReqChan  *chanx.UnboundedChan[ReqInf]
	RespChan *chanx.UnboundedChan[RespInf]
}

func NewAsynBase() *AsynBase {
	return &AsynBase{
		ReqChan:   chanx.NewUnboundedChan[ReqInf](MessageChanCap),
		RespChan:  chanx.NewUnboundedChan[RespInf](MessageChanCap),
		CallBacks: make(map[uint64]AsynCallback),
	}
}

func (this *AsynBase) nextFcId() uint64 {
	if this.FcId == math.MaxUint64 {
		this.FcId = 0
	}
	this.FcId++
	return this.FcId
}

func (this *AsynBase) SendReq(req ReqInf, cb AsynCallback) {
	req.SetFcId(this.nextFcId())
	if cb != nil {
		this.CallBacks[req.GetFcId()] = cb
	}

	this.ReqChan.Put(req)
}

func (this *AsynBase) Resp() <-chan RespInf {
	return this.RespChan.C()
}

func (this *AsynBase) HandleResp(resp RespInf) AsynCBPtr {
	cb, ok := this.CallBacks[resp.GetFcId()]
	if !ok {
		log.Error("callback not found, fcid=%d, resp=%v", resp.GetFcId(), resp)
		return 0
	}

	x := cb(resp)
	if x == 0 {
		x = AsynCBPtr(reflect.ValueOf(cb).Pointer())
	}

	this.DeleteCallBack(resp)
	return x
}

func (this *AsynBase) DeleteCallBack(resp RespInf) {
	delete(this.CallBacks, resp.GetFcId())
}
