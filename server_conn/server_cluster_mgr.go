package server_conn

import (
	"flag"
	"fmt"
	"gitlab.sunborngame.com/base/log"
	"px/common/message"
	"px/config"
	"px/define"
	"px/svr"
)

var (
	confLogic  = flag.String("logic", "config/dev_logic_cluster.xml", "logic cluster config path")
	confRouter = flag.String("router", "config/dev_router_cluster.xml", "router cluster config path")
)

const (
	defaultPipeSize = 10240

	defaultConnTimeout = 10
)

type ServerClusterMgr struct {
	logicCluster  svr.LogicClusterInf
	routerCluster svr.RouterClusterInf

	msgBus chan interface{}
}

var serverClusterMgr = &ServerClusterMgr{
	msgBus: make(chan interface{}, defaultPipeSize),
}

func GetServerClusterMgr() *ServerClusterMgr {
	return serverClusterMgr
}

// logic是服务器的用途
func (this *ServerClusterMgr) LoadLogicClusterOfServer(needLogicCluster bool, needRouterCluster bool) {
	if needLogicCluster {
		LoadLogicClusterCfg()
		this.logicCluster = svr.NewLogicClusterForServer()
		logicClusterServer, _ := this.logicCluster.(*svr.LogicClusterServer)
		logicClusterServer.InitLogicConnection(config.GetLogicClusterConfig().Logics())
	}
	if needRouterCluster {
		LoadRouterClusterCfg()
		this.routerCluster = svr.NewRouterClusterForServer()
		routerClusterServer, _ := this.routerCluster.(*svr.RouterClusterServer)
		routerClusterServer.InitRouterConnection(config.GetRouterClusterConfig().Routers())
	}

	this.start()
}

// logic是client的用途
func (this *ServerClusterMgr) LoadServerClusterOfClient(accSvrInf svr.AcceptorSvrInf, needLogicCluster bool, needRouterCluster bool) {
	if needLogicCluster {
		LoadLogicClusterCfg()
		this.logicCluster = svr.NewLogicClusterForClient(accSvrInf)
	}
	if needRouterCluster {
		LoadRouterClusterCfg()
		this.routerCluster = svr.NewRouterClusterForClient(accSvrInf)
	}

	this.start()
}

func (this *ServerClusterMgr) start() {
	if this.logicCluster != nil {
		go this.loopLogicCluster()
	}
	if this.routerCluster != nil {
		go this.loopRouterCluster()
	}
}

func (this *ServerClusterMgr) loopLogicCluster() {
	for {
		select {
		case msg, ok := <-this.logicCluster.GetMsgChan():
			if !ok {
				log.Error("logiccluster get msgchan err")
				continue
			}
			this.msgBus <- msg
		}
	}
}

func (this *ServerClusterMgr) loopRouterCluster() {
	for {
		select {
		case msg, ok := <-this.routerCluster.GetMsgChan():
			if !ok {
				log.Error("routercluster get msgchan err")
				continue
			}
			this.msgBus <- msg
		}
	}
}

func (this *ServerClusterMgr) GetMsgChan() <-chan interface{} {
	return this.msgBus
}

func LoadLogicClusterCfg() {
	err := config.LoadLogicClusterCfg(*confLogic)
	if nil != err {
		panic(fmt.Errorf("load_logic_cluster err:%v", err))
	}
	for serverId := int32(1); serverId <= int32(len(config.GetLogicClusterConfig().Logics())); serverId++ {
		var contain bool
		for _, cfg := range config.GetLogicClusterConfig().Logics() {
			if int32(cfg.ID) == serverId {
				contain = true
			}
		}
		if !contain {
			panic(fmt.Sprintf("logic server id must in series, need server:%d", serverId))
		}
	}
}

func LoadRouterClusterCfg() {
	err := config.LoadRouterClusterCfg(*confRouter)
	if nil != err {
		panic(fmt.Errorf("load_router_cluster err:%v", err))
	}
	for serverId := int32(1); serverId <= int32(len(config.GetRouterClusterConfig().Routers())); serverId++ {
		var contain bool
		for _, cfg := range config.GetRouterClusterConfig().Routers() {
			if int32(cfg.ID) == serverId {
				contain = true
			}
		}
		if !contain {
			panic(fmt.Sprintf("router server id must in series, need server:%d", serverId))
		}
	}
}

func (this *ServerClusterMgr) RecordServerSess(serverType int32, svrId int32, sessId uint64) {
	if serverType == define.ServerTypeLogic {
		logicClusterClient, ok := this.logicCluster.(*svr.LogicClusterClient)
		if !ok {
			log.Error("type change error, logicCluster is %v", this.logicCluster)
			return
		}
		logicClusterClient.RecordServerSess(svrId, sessId)
	} else if serverType == define.ServerTypeRouter {
		routerClusterClient, ok := this.routerCluster.(*svr.RouterClusterClient)
		if !ok {
			log.Error("type change error, routerCluster is %v", this.routerCluster)
			return
		}
		routerClusterClient.RecordServerSess(svrId, sessId)
	}
}

func (this *ServerClusterMgr) SendToLogic(svrId int32, msg message.Message) {
	this.logicCluster.SendToLogic(svrId, msg)
}

func (this *ServerClusterMgr) RemoveServerSess(sessId uint64) {
	logicClusterClient, ok := this.logicCluster.(*svr.LogicClusterClient)
	if ok {
		if logicClusterClient.RemoveLogicServerSess(sessId) {
			log.Warning("logic cluster session removed:%d", sessId)
		}
	}
	routerClusterClient, ok := this.routerCluster.(*svr.RouterClusterClient)
	if ok {
		if routerClusterClient.RemoveRouterServerSess(sessId) {
			log.Warning("router cluster session removed:%d", sessId)
		}
	}
}

func (this *ServerClusterMgr) GetLogicSvrIdByUserId(userId uint64) int32 {
	return this.logicCluster.GetLogicSvrIdByUserId(userId)
}

func (this *ServerClusterMgr) SendToHomeByUserId(userId uint64, msg message.Message) {
	this.routerCluster.SendToHomeByUserId(userId, msg)
}

func (this *ServerClusterMgr) SendToClient(userId uint64, msg message.Message) {
	this.routerCluster.SendToClient(userId, msg)
}

func (this *ServerClusterMgr) SendToHome(userId uint64, toHomeSvrId int32, msg message.Message) {
	this.routerCluster.SendToServer(msg, define.ServerTypeHome, toHomeSvrId, userId)
}

func (this *ServerClusterMgr) SendToGate(toGateSvrId int32, msg message.Message) {
	this.routerCluster.SendToServer(msg, define.ServerTypeGate, toGateSvrId, 0)
}

func (this *ServerClusterMgr) SendToAllLogin(msg message.Message) {
	this.routerCluster.SendToServer(msg, define.ServerTypeForLoginBroadcast, 0, 0)
}

func (this *ServerClusterMgr) SendToAllScene(msg message.Message) {
	this.routerCluster.SendToServer(msg, define.ServerTypeForHomeBroadcast, 0, 0)
}

func (this *ServerClusterMgr) SendToCenter(msg message.Message, toSvrType int32) {
	this.routerCluster.SendToServer(msg, toSvrType, 0, 0)
}

func (this *ServerClusterMgr) GetAllLogicSvrId() []int32 {
	return this.logicCluster.GetAllLogicSvrId()
}

func (this *ServerClusterMgr) BroadcastToLogin(msg message.Message) {
	this.routerCluster.SendToServer(msg, define.ServerTypeForLoginBroadcast, 0, 0)
}

func (this *ServerClusterMgr) BroadcastToGate(msg message.Message) {
	this.routerCluster.SendToServer(msg, define.ServerTypeForGateBroadcast, 0, 0)
}

func (this *ServerClusterMgr) BroadcastToRouter(msg message.Message) {
	this.routerCluster.BroadCast(msg)
}

func (this *ServerClusterMgr) SendToApi(msg message.Message) {
	this.routerCluster.SendToServer(msg, define.ServerTypeApi, 0, 0)
}

func (this *ServerClusterMgr) SendToBattleReport(msg message.Message) {
	this.routerCluster.SendToServer(msg, define.ServerTypeBattleReport, 0, 0)
}

func (this *ServerClusterMgr) SendToGM(msg message.Message) {
	this.routerCluster.SendToServer(msg, define.ServerTypeGM, 0, 0)
}

func (this *ServerClusterMgr) SendToAccount(msg message.Message) {
	this.routerCluster.SendToServer(msg, define.ServerTypeAccount, 0, 0)
}

func (this *ServerClusterMgr) SendToCDKey(msg message.Message) {
	this.routerCluster.SendToServer(msg, define.ServerTypeCDKey, 0, 0)
}

func (this *ServerClusterMgr) SendToLogin(msg message.Message) {
	this.routerCluster.SendToServer(msg, define.ServerTypeLogin, 0, 0)
}

func (this *ServerClusterMgr) SendToBattleManager(msg message.Message) {
	this.routerCluster.SendToServer(msg, define.ServerTypeBattleManager, 0, 0)
}

func (this *ServerClusterMgr) SendToDsa(dsaId int32, msg message.Message) {
	this.routerCluster.SendToServer(msg, define.ServerTypeDsa, dsaId, 0)
}

func (this *ServerClusterMgr) UserSendToFriend(userId uint64, msg message.Message) {
	this.routerCluster.SendToServer(msg, define.ServerTypeFriend, 0, userId)
}

func (this *ServerClusterMgr) UserSendToApi(userId uint64, toSvrId int32, msg message.Message) {
	this.routerCluster.SendToServer(msg, define.ServerTypeApi, toSvrId, userId)
}

func (this *ServerClusterMgr) UserSendToBattleReport(userId uint64, msg message.Message) {
	this.routerCluster.SendToServer(msg, define.ServerTypeBattleReport, 0, userId)
}

func (this *ServerClusterMgr) UserSendToLogic(userId uint64, msg message.Message) {
	this.logicCluster.UserSendToLogic(userId, msg)
}

func (this *ServerClusterMgr) SendToRandRouter(msg message.Message) {
	this.routerCluster.SendToRandRouter(msg)
}

func (this *ServerClusterMgr) GetLogicCount() int {
	return this.logicCluster.GetCount()
}

func (this *ServerClusterMgr) GetUserLogicClientIndex(userId uint64) uint64 {
	logicClusterServer, ok := this.logicCluster.(*svr.LogicClusterServer)
	if !ok {
		log.Error("type change error, logicCluster is %v", this.logicCluster)
		return 0
	}
	return logicClusterServer.GetUserLogicClientIndex(userId)
}

func (this *ServerClusterMgr) SendToServer(svrId, svrType int32, msg message.Message, userId uint64) {
	if svrType == define.ServerTypeLogic {
		this.logicCluster.SendToLogic(svrId, msg)
	} else {
		this.routerCluster.SendToServer(msg, svrType, svrId, userId)
	}
}
