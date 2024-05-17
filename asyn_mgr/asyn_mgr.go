package asyn_mgr

import (
	"px/shared/asyn_mgr/asyn_msg"
	"px/utils/chanx"
	"time"

	"gitlab.sunborngame.com/base/log"
)

type AsynMgr struct {
	fcId        uint64
	respChan    *chanx.UnboundedChan[asyn_msg.RespInf]
	modules     map[asyn_msg.AsynModuleId]asyn_msg.AsynModInf
	respMsgChan *chanx.UnboundedChan[asyn_msg.RespMsgInf]
}

var asynMgr = &AsynMgr{
	modules:     make(map[asyn_msg.AsynModuleId]asyn_msg.AsynModInf),
	respChan:    chanx.NewUnboundedChan[asyn_msg.RespInf](MessageChanCap),
	respMsgChan: chanx.NewUnboundedChan[asyn_msg.RespMsgInf](MessageChanCap),
}

func GetAsynMgr() *AsynMgr {
	return asynMgr
}

func (this *AsynMgr) RegisterAsynModule(modules ...asyn_msg.AsynModInf) {
	for _, module := range modules {
		this.modules[module.GetModuleId()] = module
	}
}

func (this *AsynMgr) Init() {
	for _, module := range this.modules {
		log.Info("init modules %v", module.GetModuleId())
		module.Init()

		go this.loop(module)
	}
}

func (this *AsynMgr) Start() {

}

func (this *AsynMgr) OnClose() {
	for _, module := range this.modules {
		module.Close()
	}
}

func (this *AsynMgr) Resp() <-chan asyn_msg.RespInf {
	return this.respChan.C()
}

func (this *AsynMgr) RespMsg() <-chan asyn_msg.RespMsgInf {
	return this.respMsgChan.C()
}

func (this *AsynMgr) ReqLen() int {
	l := 0
	for _, module := range this.modules {
		l += module.ReqLen()
	}
	return l
}

func (this *AsynMgr) RespLen() int {
	return this.respChan.Len()
}

func (this *AsynMgr) loop(module asyn_msg.AsynModInf) {
	for {
		select {
		case msg, ok := <-module.Resp():
			if !ok {
				log.Error("modules closed, modules=%d", module.GetModuleId())
				return
			}

			switch respMsg := msg.(type) {
			case *asyn_msg.RespMsg:
				this.respMsgChan.Put(respMsg)
			default:
				this.respChan.Put(respMsg)
			}
		}
	}
}

func (this *AsynMgr) SendReq(req asyn_msg.ReqInf, cb asyn_msg.AsynCallback) {
	module, ok := this.modules[req.GetModuleId()]
	if !ok {
		log.Error("not found modules, req=%v", req)
		return
	}
	module.SendReq(req, cb)
}

func (this *AsynMgr) GetASynModule(moduleId asyn_msg.AsynModuleId) (asyn_msg.AsynModInf, bool) {
	module, ok := this.modules[moduleId]
	if !ok {
		log.Error("not found modules, moduleId=%d, all modules=%v", moduleId, this.modules)
		return nil, false
	}

	return module, true
}

func (this *AsynMgr) HandleResp(respInf asyn_msg.RespInf) asyn_msg.AsynCBPtr {
	start := time.Now().UnixMilli()
	module, ok := this.modules[respInf.GetModuleId()]
	if !ok {
		log.Error("modules not found, moduleId=%v", respInf.GetModuleId())
		return 0
	}
	cb := module.HandleResp(respInf)

	end := time.Now().UnixMilli()
	if end-start > 0 {
		log.Info("HandleResp:%d modules:%d cost:%dms", respInf.GetFcId(), int64(respInf.GetModuleId()), end-start)
	}
	return cb
}
