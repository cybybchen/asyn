package etcd_data_inf

import (
	"gitlab.sunborngame.com/base/log"
	"reflect"
)

var msgCreates = make(map[string]reflect.Type)

func RegisterMsgCreate(rsInf EtcdDataInf) {
	rsType := reflect.TypeOf(rsInf).Elem()
	_, ok := msgCreates[rsType.Name()]
	if ok {
		log.Panic("[etcd_mgr] msgCreate already registered, typeName:%s, rsInf:%v", rsType.Name(), rsInf)
	}

	msgCreates[rsType.Name()] = rsType
}

func CreateMsg(name string) EtcdDataInf {
	tp, ok := msgCreates[name]
	if !ok {
		log.Error("[etcd_mgr] get msg create failed, name:%s", name)
		return nil
	}

	etcdData, ok := reflect.New(tp).Interface().(EtcdDataInf)
	if !ok {
		log.Error("[etcd_mgr] get msg create failed, name:%s", name)
		return nil
	}

	return etcdData
}
