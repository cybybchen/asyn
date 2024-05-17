package etcd

import (
	"encoding/json"
	"px/shared/asyn_mgr/etcd/etcd_data_inf"
)

type TestEtcd struct {
	Key    int
	Value  string
	Value2 int
	Value3 []byte
}

func init() {
	etcd_data_inf.RegisterMsgCreate(&TestEtcd{})
}

func (this *TestEtcd) MarshalBinary() (data []byte, err error) {
	return json.Marshal(this)
}

func (this *TestEtcd) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, this)
}
