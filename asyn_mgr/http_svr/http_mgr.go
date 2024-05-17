package http_svr

import (
	"fmt"
	"px/framebase"
	"px/shared/time_wheel"
	"px/utils"
	"px/utils/cbctx"
	"sync"

	"gitlab.sunborngame.com/base/log"
)

var (
	HttpAsyncMgr = CreateHttpSvrMgr()
)

type Req interface {
	ReqId() string
	Timeout() int64

	mgrIn()
	mgrOut()
	Wait() // 等待结束
}

type HttpReqBase struct {
	ReqId_   string
	Timeout_ int64 // ms
	wait     sync.WaitGroup
}

func (rbase *HttpReqBase) ReqId() string {
	return rbase.ReqId_
}

func (rbase *HttpReqBase) Timeout() int64 {
	return rbase.Timeout_
}

func (rbase *HttpReqBase) mgrIn() {
	rbase.wait.Add(1)
}

func (rbase *HttpReqBase) mgrOut() {
	rbase.wait.Done()
}

func (rbase *HttpReqBase) Wait() {
	rbase.wait.Wait()
}

type Ack interface {
	ReqId() string
	Error() error
}

type HttpAckBase struct {
	ReqId_ string
	Err_   error
}

func (rbase *HttpAckBase) ReqId() string {
	return rbase.ReqId_
}

func (fa *HttpAckBase) Error() error {
	return fa.Err_
}

type AsynCBPtr uintptr // 回调函数指针
type HttpCallback func(r Req, a Ack) AsynCBPtr
type ReqData struct {
	req    Req
	cb     HttpCallback
	timeId int64
}

type HttpSvrMgr struct {
	CallBack sync.Map // string -> Req, TimeId, callBack
	timer    *time_wheel.TimeWheelS
}

func CreateHttpSvrMgr() *HttpSvrMgr {
	httpSvr := &HttpSvrMgr{
		timer: time_wheel.NewTimeWheelSDefault(),
	}

	httpSvr.timer.Start()

	go httpSvr.loop()
	return httpSvr
}

func (this *HttpSvrMgr) loop() {
	defer utils.Recover(framebase.SendWeChatMsg(), framebase.IsReleaseEnv())

	for {
		select {
		case task, ok := <-this.timer.NotifyChannel.C():
			if !ok {
				log.Error("HttpSvrMgr Quit Timer")
				return
			}
			this.timer.TriggerTimerCb(task)
		}
	}

}

func (this *HttpSvrMgr) OnReq(req Req, cb HttpCallback) bool {
	info := ReqData{
		req: req,
		cb:  cb,
	}

	_, exsit := this.CallBack.LoadOrStore(req.ReqId(), &info)
	if exsit {
		cb(req, &HttpAckBase{
			ReqId_: req.ReqId(),
			Err_:   fmt.Errorf("%s Dup", req.ReqId()),
		})
		return false
	}

	if req.Timeout() > 0 {
		info.timeId = this.timer.AddOnceTimer(req.Timeout(), this.timerOut, req.ReqId())
	}

	req.mgrIn()

	return true
}

func (this *HttpSvrMgr) timerOut(now int64, ctx cbctx.Ctx) {
	if len(ctx) == 0 {
		log.Error("timeOut No Param")
		return
	}

	cst, ok := ctx[0].(cbctx.CtxString)
	if !ok {
		log.Error("timeOut Param not String:%t", ctx[0])
		return
	}

	reqId := string(cst)
	d, ok := this.CallBack.LoadAndDelete(reqId)
	if !ok {
		log.Error("ReqId:%s Load Null", reqId)
		return
	}

	data, ok := d.(*ReqData)
	if !ok {
		log.Error("ReqId:%s type:%t", reqId, d)
		return
	}

	data.cb(data.req, &HttpAckBase{
		ReqId_: reqId,
		Err_:   fmt.Errorf("%s Timeout", reqId),
	})

	data.req.mgrOut()
}

func (this *HttpSvrMgr) OnAck(ack Ack) {
	v, ok := this.CallBack.LoadAndDelete(ack.ReqId())
	if !ok {
		log.Warning("ReqId:%s OnAck Null", ack.ReqId())
		return
	}

	data, ok := v.(*ReqData)
	if !ok {
		log.Error("ReqId:%s OnAck type:%t", ack.ReqId(), v)
		return
	}

	if data.timeId > 0 {
		this.timer.RemoveTimer(data.timeId)
	}

	data.cb(data.req, ack)
	data.req.mgrOut()
}
