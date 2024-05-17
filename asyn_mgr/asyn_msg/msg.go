package asyn_msg

import (
	"px/common/message"
)

type (
	AsynCBPtr uintptr // 回调函数指针
	// 异步回调函数，如果内有嵌套回调，需传出内部回调指针，用来进行Pprof
	AsynCallback func(RespInf) AsynCBPtr

	AsynModInf interface {
		GetModuleId() AsynModuleId
		ReqLen() int
		Resp() <-chan RespInf
		Init()
		// 如果有嵌套，可以很方便的获取到最内层的回调
		SendReq(ReqInf, AsynCallback)
		HandleResp(inf RespInf) AsynCBPtr
		Close()
	}
	ReqInf interface {
		SetFcId(uint64)
		GetFcId() uint64
		GetModuleId() AsynModuleId
	}
	RespInf interface {
		SetFcId(uint64)
		GetFcId() uint64
		GetModuleId() AsynModuleId
	}
	RespMsgInf interface {
		GetMessage() message.Message
		GetUserIds() []uint64
		BroadcastAll() bool
	}

	ReqBase struct {
		fcId uint64
	}
	RespBase struct {
		fcId uint64
	}
	RespMsg struct {
		RespBase
		UserIds   []uint64
		Msg       message.Message
		Broadcast bool
		ModuleId  AsynModuleId
	}
)

func (this *ReqBase) SetFcId(fcId uint64) {
	this.fcId = fcId
}

func (this *ReqBase) GetFcId() uint64 {
	return this.fcId
}

func (this *RespBase) SetFcId(fcId uint64) {
	this.fcId = fcId
}

func (this *RespBase) GetFcId() uint64 {
	return this.fcId
}

func (this *RespMsg) GetUserIds() []uint64 {
	return this.UserIds
}

func (this *RespMsg) GetMessage() message.Message {
	return this.Msg
}

func (this *RespMsg) BroadcastAll() bool {
	return this.Broadcast
}

func (this *RespMsg) GetModuleId() AsynModuleId {
	return this.ModuleId
}
