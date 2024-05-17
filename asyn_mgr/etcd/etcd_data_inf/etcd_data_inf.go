package etcd_data_inf

type EtcdDataInf interface {
	MarshalBinary() (data []byte, err error)
	UnmarshalBinary(data []byte) error
}
