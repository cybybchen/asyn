package etcd

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"gitlab.sunborngame.com/base/log"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
	"px/config"
	"px/framebase"
	"px/shared/asyn_mgr/asyn_msg"
	"px/shared/asyn_mgr/etcd/etcd_data_inf"
	"px/utils"
	"reflect"
	"time"
)

var (
	etcdConf = flag.String("etcd", "config/dev_etcd.xml", "etcd config path")
)

const (
	LogTag    = "[etcd_mgr]"
	OpTimeout = 5 * time.Second
)

type EtcdMgr struct {
	*asyn_msg.AsynBase
	cli *clientv3.Client
}

func CreateEtcd() asyn_msg.AsynModInf {
	var etcdMgr = &EtcdMgr{
		AsynBase: asyn_msg.NewAsynBase(),
	}

	initConf()

	etcdMgr.initClientv3()

	return etcdMgr
}

func initConf() {
	err := config.LoadEtcdCfg(*etcdConf)
	if nil != err {
		panic(fmt.Errorf("LoadEtcdCfg err:%v", err))
		return
	}
}

func (this *EtcdMgr) Init() {
	go this.loop()
}

func (this *EtcdMgr) Close() {

}

func (this *EtcdMgr) loop() {
	for {
		if !this.run() {
			break
		}
	}
}

func (this *EtcdMgr) run() bool {
	defer utils.Recover(framebase.SendWeChatMsg(), framebase.IsReleaseEnv())

	select {
	case req, ok := <-this.ReqChan.C():
		if !ok {
			return false
		}
		this.handleReq(req)
	}

	return true
}

func (this *EtcdMgr) GetModuleId() asyn_msg.AsynModuleId {
	return asyn_msg.EtcdModuleId
}

func (this *EtcdMgr) ReqLen() int {
	return 0
}

func (this *EtcdMgr) initClientv3() {
	var dialTimeout = config.GetEtcdConfig().GetDialTimeout()
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   config.GetEtcdConfig().GetEndpoints(),
		DialTimeout: time.Second * time.Duration(dialTimeout),
	})
	if err != nil {
		log.Panic("[etcd_mgr] create err =%v", err)
	}
	this.cli = cli
}

func (this *EtcdMgr) handleReq(req asyn_msg.ReqInf) {
	log.Info("%s etcd recv req:%v", LogTag, req)
	switch msg := req.(type) {
	case *ReqWrite:
		this.handleWrite(msg)
	case *ReqWatch:
		go this.handleWatch(msg)
	default:
		log.Error("%s etcd reqMsg err %v", LogTag, req)
	}
}

func (this *EtcdMgr) handleWrite(req *ReqWrite) {
	var resp = &RespWrite{}
	resp.SetFcId(req.GetFcId())
	defer this.RespChan.Put(resp)

	etcdData, err := this.encode(req.Value)
	if err != nil {
		resp.Err = err.Error()
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), OpTimeout)
	defer cancel()

	// 写入数据
	_, err = this.cli.Put(ctx, req.Key, etcdData)
	if err != nil {
		log.Error("%s etcd put err:%v", LogTag, err)
	}
}

func (this *EtcdMgr) handleWatch(req *ReqWatch) {
	var options []clientv3.OpOption
	if req.WithPrefix {
		options = append(options, clientv3.WithPrefix())
	}
	res, err := this.cli.Get(context.TODO(), req.Key, options...)
	if err != nil {
		log.Error("[EtcdMgr] err : %s", err.Error())
		return
	}

	var resp = &RespWatch{
		Datas: make([]*EtcdOpData, 0, len(res.Kvs)),
	}
	resp.SetFcId(req.GetFcId())
	for _, kv := range res.Kvs {
		var value = this.decode(string(kv.Value))
		if value == nil {
			continue
		}
		resp.Datas = append(resp.Datas, &EtcdOpData{
			Key:   string(kv.Key),
			Value: value,
			Op:    mvccpb.PUT,
		})
	}
	if len(resp.Datas) != 0 {
		this.RespChan.Put(resp)
	}

	ch := this.cli.Watch(context.TODO(), req.Key, options...)
	for {
		select {
		case c := <-ch:
			resp.Datas = make([]*EtcdOpData, 0, len(c.Events))
			for _, e := range c.Events {
				var value = this.decode(string(e.Kv.Value))
				if value == nil {
					continue
				}
				switch e.Type {
				case clientv3.EventTypePut:
					resp.Datas = append(resp.Datas, &EtcdOpData{
						Op:    mvccpb.PUT,
						Key:   string(e.Kv.Key),
						Value: value,
					})
				case clientv3.EventTypeDelete:
					resp.Datas = append(resp.Datas, &EtcdOpData{
						Op:    mvccpb.DELETE,
						Key:   string(e.Kv.Key),
						Value: value,
					})
				}
			}
			if len(resp.Datas) > 0 {
				this.RespChan.Put(resp)
			}
		}
	}
}

func (this *EtcdMgr) HandleResp(resp asyn_msg.RespInf) asyn_msg.AsynCBPtr {
	switch resp.(type) {
	case *RespWrite:
		return this.AsynBase.HandleResp(resp)
	default:
		cb, ok := this.CallBacks[resp.GetFcId()]
		if !ok {
			log.Error("callback not found, fcid=%d, resp=%v", resp.GetFcId(), resp)
			return 0
		}

		x := cb(resp)
		if x == 0 {
			x = asyn_msg.AsynCBPtr(reflect.ValueOf(cb).Pointer())
		}

		return x
	}
}

func (this *EtcdMgr) encode(value interface{}) (string, error) {
	var tp = etcd_data_inf.VTypeNone
	var tpName string
	switch v := value.(type) {
	case string:
		tp = etcd_data_inf.VTypeString
	case int32, uint32:
		tp = etcd_data_inf.VTypeInt32
	case int64, uint64:
		tp = etcd_data_inf.VTypeInt64
	case []byte:
		tp = etcd_data_inf.VTypeBytes
	case etcd_data_inf.EtcdDataInf:
		tp = etcd_data_inf.VTypeData
		refTp := reflect.TypeOf(v).Elem()
		tpName = refTp.Name()
	}

	var etcdData = &etcd_data_inf.EtcdData{
		Head: &etcd_data_inf.EtcdDataHead{
			Tp:     tp,
			TpName: tpName,
		},
		Body: value,
	}

	ret, err := etcdData.MarshalBinary()
	if err != nil {
		log.Error("%s marshal err:%v", LogTag, err)
		return "", err
	}

	return string(ret), nil
}

func (this *EtcdMgr) decode(value string) any {
	var rs = &etcd_data_inf.EtcdData{
		Head: &etcd_data_inf.EtcdDataHead{},
	}
	err := rs.UnmarshalBinary([]byte(value))
	if err != nil {
		log.Error("%s etcdData unmarshal failed, err:%v", LogTag, err)
		return nil
	}

	switch rs.Head.Tp {
	case etcd_data_inf.VTypeString:
		return rs.Body
	case etcd_data_inf.VTypeInt32:
		return int32(rs.Body.(float64))
	case etcd_data_inf.VTypeUInt32:
		return uint32(rs.Body.(float64))
	case etcd_data_inf.VTypeInt64:
		return int64(rs.Body.(float64))
	case etcd_data_inf.VTypeUInt64:
		return uint64(rs.Body.(float64))
	case etcd_data_inf.VTypeBytes:
		data, err := base64.StdEncoding.DecodeString(rs.Body.(string))
		if err != nil {
			log.Error("%s base64 decode failed, err:%v", LogTag, err)
			return nil
		}
		rs.Body = data
	case etcd_data_inf.VTypeData:
		rs.Body = etcd_data_inf.CreateMsg(rs.Head.TpName)
		if rs.Body == nil {
			return nil
		}
		err = rs.UnmarshalBinary([]byte(value))
		if err != nil {
			log.Error("%s etcdData unmarshal failed, err:%v", LogTag, err)
			return nil
		}
		return rs.Body
	default:
		log.Error("unhandled default tp, rs:%v", rs)
	}

	return nil
}
