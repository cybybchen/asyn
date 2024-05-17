package asyn_aoi

import (
	"twe/define"
	"twe/shared/asyn_mgr/asyn_msg"
	"twe/utils/lmath"
	"twe/utils/quadTree"
)

type ReqAoiInf interface {
	asyn_msg.ReqInf
	GetAsynAoiTyp() int32
	GetGuid() uint64
	GetOperateTyp() define.AsynAoiOperateTyp
	GetBound() quadTree.Bounds
}

type ReqAoiBase struct {
	asyn_msg.ReqBase
	typ        int32
	guid       uint64
	bound      quadTree.Bounds
	pos        lmath.Vector3
	operateTyp int32
}

func (this *ReqAoiBase) GetAsynAoiTyp() int32 {
	return this.typ
}

func (this *ReqAoiBase) SetAsynAoiTyp(asynAoiTyp int32) {
	this.typ = asynAoiTyp
}

func (this *ReqAoiBase) GetGuid() uint64 {
	return this.guid
}

func (this *ReqAoiBase) SetGuid(guid uint64) {
	this.guid = guid
}

func (this *ReqAoiBase) GetBound() quadTree.Bounds {
	return this.bound
}

func (this *ReqAoiBase) GetPos() lmath.Vector3 {
	return this.pos
}

func (this *ReqAoiBase) InitReqAoiBase(typ int32, bound quadTree.Bounds, guid uint64, pos lmath.Vector3) {
	this.typ = typ
	this.bound = bound
	this.guid = guid
	this.pos = pos
}

func (this *ReqAoiBase) GetModuleId() define.AsynModuleId {
	return define.AoiModuleId
}

type ReqAddCamera struct {
	ReqAoiBase
}

func (this *ReqAddCamera) GetOperateTyp() define.AsynAoiOperateTyp {
	return define.AsynAoiOperateTyp_AddCamera
}

type ReqUpdateCamera struct {
	ReqAoiBase
}

func (this *ReqUpdateCamera) GetOperateTyp() define.AsynAoiOperateTyp {
	return define.AsynAoiOperateTyp_UpdateCamera
}

type ReqRemoveCamera struct {
	ReqAoiBase
}

func (this *ReqRemoveCamera) GetOperateTyp() define.AsynAoiOperateTyp {
	return define.AsynAoiOperateTyp_RemoveCamera
}

type ReqAddAsynAoiModel struct {
	ReqAoiBase
}

func (this *ReqAddAsynAoiModel) GetOperateTyp() define.AsynAoiOperateTyp {
	return define.AsynAoiOperateTyp_AddModel
}

type ReqUpdateAsynAoiModel struct {
	ReqAoiBase
}

func (this *ReqUpdateAsynAoiModel) GetOperateTyp() define.AsynAoiOperateTyp {
	return define.AsynAoiOperateTyp_UpdateModel
}

type ReqRemoveAsynAoiModel struct {
	ReqAoiBase
}

func (this *ReqRemoveAsynAoiModel) GetOperateTyp() define.AsynAoiOperateTyp {
	return define.AsynAoiOperateTyp_RemoveModel
}

type UpdateAndBroadcastAsynAoiModel struct {
	ReqAoiBase
}

func (this *UpdateAndBroadcastAsynAoiModel) GetOperateTyp() define.AsynAoiOperateTyp {
	return define.AsynAoiOperateTyp_BroadcastModel
}

type ReqLog struct {
	ReqAoiBase
}

func (this *ReqLog) GetOperateTyp() define.AsynAoiOperateTyp {
	return define.AsynAoiOperateTyp_Log
}
