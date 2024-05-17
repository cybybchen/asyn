package etcd_data_inf

import "encoding/json"

type EtcdValueType int32

const (
	VTypeNone EtcdValueType = iota
	VTypeInt32
	VTypeUInt32
	VTypeInt64
	VTypeUInt64
	VTypeString
	VTypeBytes
	VTypeData
)

type (
	EtcdData struct {
		Head *EtcdDataHead
		Body any
	}
)

type EtcdDataHead struct {
	Tp     EtcdValueType
	TpName string
}

func (this *EtcdData) MarshalBinary() (data []byte, err error) {
	return json.Marshal(this)
}

func (this *EtcdData) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, this)
}
