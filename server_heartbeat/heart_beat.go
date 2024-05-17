package server_heartbeat

import (
	"px/define"
	"px/framebase"
	"px/home/framework"
	"px/proto/proto_ssmsg"
	"px/shared/server_conn"
	"px/utils/cbctx"
)

type HeartBeat struct {
	homeIds []int32
	addr    string
}

var instance = &HeartBeat{}

func GetHeartBeat() *HeartBeat {
	return instance
}

func (this *HeartBeat) SetGateInfo(homeIds []int32, addr string) {
	this.homeIds, this.addr = homeIds, addr
}

func (this *HeartBeat) Start() {
	framebase.Timer.AddRepeatTimer(define.ServerHeartBeatTick, this.serverHeartBeat)
}

func (this *HeartBeat) serverHeartBeat(int64, cbctx.Ctx) {
	server_conn.GetServerClusterMgr().BroadcastToLogin(&proto_ssmsg.GC_SS_ServerHeartBeat{
		ServerId:     int32(framebase.GetServerId()),
		ServerOpenTs: framework.GetServerOpenTs(),
		ServerType:   framebase.ServerType,
		HomeIds:      this.homeIds,
		Addr:         this.addr,
	})
}
