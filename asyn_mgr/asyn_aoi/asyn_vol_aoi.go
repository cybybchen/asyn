package asyn_aoi

import (
	"twe/define"
	"twe/proto/proto_csmsg"
	"twe/proto/proto_object"
	"twe/shared/asyn_mgr/asyn_msg"
	"twe/utils"
	"twe/utils/lmath"
	"twe/utils/quadTree"
)

func init() {
	RegAoiCreateFc(define.AsynVoltageAoi, CreateAsynVoltageAoi())
}

type VoltageViewModel struct {
	LastTime   int64 //创建或者上次更换占领者的时间
	State      int32 //	State             proto_object.VOLTAGE_STATE
	CfgId      int32 //配置id
	EntityType int32 //实体类型

	BelongType     int32  //归属类型
	BelongToUserId uint64 //归属玩家
	BelongToGuild  uint64 //归属联盟
}

type ReqAddAynsAoiVoltage struct {
	ReqAddAsynAoiModel
	VoltageView VoltageViewModel
}

type ReqUpdateAynsAoiVoltage struct {
	ReqUpdateAsynAoiModel
	VoltageView VoltageViewModel
}

type ReqRemoveAynsAoiVoltage struct {
	ReqRemoveAsynAoiModel
}

type ReqUpdateAndBroadcastAsynAoiVoltage struct {
	UpdateAndBroadcastAsynAoiModel
	VoltageView VoltageViewModel
}

type AsynVoltageModel struct {
	AsynEntityModelBase
	volViewModel VoltageViewModel
}

func createVoltageModel(guid uint64, bound quadTree.Bounds, pos lmath.Vector3, voltageViewModel VoltageViewModel) *AsynVoltageModel {
	model := &AsynVoltageModel{}
	model.InitAsynEntityModelBase(guid, bound, pos)
	model.volViewModel = voltageViewModel
	return model
}

type AsynVoltageAoi struct {
	AsynAoiBase
}

func (view *AsynVoltageAoi) GetAsynAoiTyp() int32 {
	return int32(define.AsynVoltageAoi)
}

func CreateAsynVoltageAoi() *AsynVoltageAoi {
	view := &AsynVoltageAoi{}
	view.InitAsynAoiBase(view.NotifyVoltageEnter, view.NotifyVoltageLeave)
	return view
}

func (view *AsynVoltageAoi) AddModel(model ReqAoiInf) {
	addModelInfo, ok := model.(*ReqAddAynsAoiVoltage)
	if !ok {
		return
	}

	voltageModel := createVoltageModel(addModelInfo.GetGuid(), addModelInfo.GetBound(), addModelInfo.GetPos(), addModelInfo.VoltageView)
	view.addAsynAoiEntityModel(voltageModel)
}

func (view *AsynVoltageAoi) UpdateModel(model ReqAoiInf) {
	updateModelInfo, ok := model.(*ReqUpdateAynsAoiVoltage)
	if !ok {
		return
	}

	voltageModel := createVoltageModel(updateModelInfo.GetGuid(), updateModelInfo.GetBound(), updateModelInfo.GetPos(), updateModelInfo.VoltageView)
	view.updateAsynAoiEntityModel(voltageModel)
}

func (view *AsynVoltageAoi) NotifyVoltageEnter(entities utils.GuidMap, cameras utils.GuidMap) {
	if len(entities) <= 0 || len(cameras) <= 0 {
		return
	}

	msgSC := &proto_csmsg.SC_NotifyAddVoltageModel{}
	for guid := range entities {
		msgSC.VoltageModels = append(msgSC.VoltageModels, view.PackVoltageModel(guid))
	}

	var userIds = make([]uint64, 0, len(cameras))
	for cameraGuid := range cameras {
		userIds = append(userIds, cameraGuid)
	}
	GetMgr().RespChan.Put(&asyn_msg.RespMsg{
		UserIds: userIds,
		Msg:     msgSC,
	})
}

func (view *AsynVoltageAoi) NotifyVoltageLeave(entities utils.GuidMap, cameras utils.GuidMap) {
	if len(entities) <= 0 || len(cameras) <= 0 {
		return
	}

	msgSC := &proto_csmsg.SC_NotifyRemoveVoltageModel{}
	for guid := range entities {
		msgSC.VoltageModelIds = append(msgSC.VoltageModelIds, guid)
	}

	var userIds = make([]uint64, 0, len(cameras))
	for cameraGuid := range cameras {
		userIds = append(userIds, cameraGuid)
	}
	GetMgr().RespChan.Put(&asyn_msg.RespMsg{
		UserIds: userIds,
		Msg:     msgSC,
	})
}

func (view *AsynVoltageAoi) UpdateAndBroadcastModel(model ReqAoiInf) {
	broadcastModelInfo, ok := model.(*ReqUpdateAndBroadcastAsynAoiVoltage)
	if !ok {
		return
	}

	voltageModel := view.GetAysnEntityModel(model.GetGuid())
	if voltageModel == nil {
		return
	}

	voltageModelInfo := voltageModel.(*AsynVoltageModel)
	voltageModelInfo.volViewModel = broadcastModelInfo.VoltageView

	msgSC := &proto_csmsg.SC_NotifyUpdateVoltageModel{}
	msgSC.VoltageModels = append(msgSC.VoltageModels, view.PackVoltageModel(model.GetGuid()))
	var userIds = make([]uint64, 0, len(voltageModel.GetSeeMeCameras()))
	for userId, _ := range voltageModel.GetSeeMeCameras() {
		userIds = append(userIds, userId)
	}
	GetMgr().RespChan.Put(&asyn_msg.RespMsg{
		UserIds: userIds,
		Msg:     msgSC,
	})
}

func (view *AsynVoltageAoi) PackVoltageModel(guid uint64) *proto_object.VoltageModel {
	asynEntityModel := view.GetAysnEntityModel(guid)
	if asynEntityModel == nil {
		return nil
	}

	voltageModel := asynEntityModel.(*AsynVoltageModel)
	model := &proto_object.VoltageModel{
		Guid:       voltageModel.GetGuid(),
		EntityType: voltageModel.volViewModel.EntityType,
		Pos:        utils.ToProtoVector2(voltageModel.AsynEntityModelBase.pos),
		TemplateId: voltageModel.volViewModel.CfgId,
		BuildTime:  voltageModel.volViewModel.LastTime,
		State:      voltageModel.volViewModel.State,
		Owner: &proto_object.EntityOwnerInfo{
			Belong:    voltageModel.volViewModel.BelongType,
			UserId:    voltageModel.volViewModel.BelongToUserId,
			UserGuild: voltageModel.volViewModel.BelongToGuild,
		},
	}

	return model
}
