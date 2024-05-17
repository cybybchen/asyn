package asyn_aoi

import (
	"gitlab.sunborngame.com/base/log"
	"twe/define"
	"twe/framebase"
	"twe/shared/asyn_mgr/asyn_msg"
	"twe/utils"
)

var AysnAoiPools map[define.AsynAoiTyp]IAsynAoi

type AoiMgr struct {
	*asyn_msg.AsynBase
}

type SysCreateFunc func() IAsynAoi

func RegAoiCreateFc(anysAoiTyp define.AsynAoiTyp, fc IAsynAoi) {
	if AysnAoiPools == nil {
		AysnAoiPools = map[define.AsynAoiTyp]IAsynAoi{}
	}
	AysnAoiPools[anysAoiTyp] = fc
}

func (this *AoiMgr) ReqLen() int {
	return 0
}

func (this *AoiMgr) Init() {
	go this.reqLoop()
}

func (this *AoiMgr) Close() {

}

var aoiMgr = CreateAoiMgr()

func CreateAoiMgr() *AoiMgr {
	return &AoiMgr{
		AsynBase: asyn_msg.NewAsynBase(),
	}
}

func GetMgr() *AoiMgr {
	return aoiMgr
}

func (this *AoiMgr) GetModuleId() define.AsynModuleId {
	return define.AoiModuleId
}

func (this *AoiMgr) reqLoop() {
	for {
		if !this.run() {
			break
		}
	}
}

func (this *AoiMgr) run() bool {
	defer utils.Recover(framebase.SendWeChatMsg(), framebase.IsReleaseEnv())

	select {
	case req, ok := <-this.ReqChan.C():
		if !ok {
			log.Error("astar get req chan msg error")
			return false
		}
		this.handleReq(req)
	}

	return true
}

func (this *AoiMgr) handleReq(req asyn_msg.ReqInf) {
	aoiReq, ok := req.(ReqAoiInf)
	if !ok {
		log.Error("AoiMgr req type err, req=%v", req)
		return
	}
	aoiPool := AysnAoiPools[define.AsynAoiTyp(aoiReq.GetAsynAoiTyp())]
	if aoiPool == nil {
		log.Error("aoiPool len = 0, type=%d", aoiReq.GetAsynAoiTyp())
		return
	}
	switch aoiReq.GetOperateTyp() {
	case define.AsynAoiOperateTyp_AddCamera:
		aoiPool.AddCamera(aoiReq.GetGuid(), aoiReq.GetBound())
	case define.AsynAoiOperateTyp_UpdateCamera:
		aoiPool.UpdateUserCamera(aoiReq.GetGuid(), aoiReq.GetBound())
	case define.AsynAoiOperateTyp_RemoveCamera:
		aoiPool.RemoveCamera(aoiReq.GetGuid())
	case define.AsynAoiOperateTyp_AddModel:
		aoiPool.AddModel(aoiReq)
	case define.AsynAoiOperateTyp_UpdateModel:
		aoiPool.UpdateModel(aoiReq)
	case define.AsynAoiOperateTyp_RemoveModel:
		aoiPool.RemoveModel(aoiReq.GetGuid())
	case define.AsynAoiOperateTyp_BroadcastModel:
		aoiPool.UpdateAndBroadcastModel(aoiReq)
	case define.AsynAoiOperateTyp_Log:
		aoiPool.Log(aoiReq.GetAsynAoiTyp())
	default:
		log.Error("reqMsg err %v", req)
	}

}
