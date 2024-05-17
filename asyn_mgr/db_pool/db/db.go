package db

import (
	"errors"
	"px/proto/proto_db"
)

var (
	NoUserError      = errors.New(`user not found`)
	UserMarshalError = errors.New(`user unmarshal error`)
)

type DBInter interface {
	LoadAllUsers() (dbUsers []*proto_db.UserData, err error)
	//LoadAllWorldUnits() (dbUnits []*proto_db.DBUnit, err error)
	//SaveOrDeleteWorldUnit(save bool, data *proto_db.DBUnit) (err error)
	LoadUser(id uint64) (name string, data *proto_db.UserData, err error)
	LoadGMail(opSeq uint64) (out [][]byte, seq uint64, err error)
	InsertOrUpdateUser(insert bool, uid uint64, name string, data *proto_db.UserData) (err error)
	LogData(data []byte) (err error)
	SavingCenterData(id int, data []byte) (err error)
	LoadingCenterData(id int) (data []byte, err error)
	CloseImmediately()
	CloseLazy()
	DBOperator(string, proto_db.DB_OPERATOR, []*proto_db.DBArgs) (ret []map[string][]byte, err error)
}
