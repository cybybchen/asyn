package redis_inf

import (
	"gitlab.sunborngame.com/base/log"
	"reflect"
)

var msgCreates = make(map[string]reflect.Type)

func RegisterMsgCreate(rsInf RedisDataInf) {
	rsType := reflect.TypeOf(rsInf).Elem()
	_, ok := msgCreates[rsType.Name()]
	if ok {
		log.Panic("[redis_proxy] msgCreate already registered, typeName:%s, rsInf:%v", rsType.Name(), rsInf)
	}

	msgCreates[rsType.Name()] = rsType
}

func CreateMsg(name string) RedisDataInf {
	tp, ok := msgCreates[name]
	if !ok {
		log.Error("[redis_proxy] get msg create failed, name:%s", name)
		return nil
	}

	return reflect.New(tp).Interface().(RedisDataInf)
}
