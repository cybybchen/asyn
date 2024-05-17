package db_pool

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"px/proto/proto_db"
	"px/shared/asyn_mgr/asyn_msg"
	"px/shared/asyn_mgr/db_pool/db"
	"px/utils/chanx"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gitlab.sunborngame.com/base/log"

	_ "flag"

	"google.golang.org/protobuf/proto"
)

type Config struct {
	DBAddr      string `json:"db_url"`
	DBUser      string `json:"db_user"`
	DBPassword  string `json:"db_passwd"`
	DBDatabase  string `json:"db_dbname"`
	DBTableName string `json:"db_table"`

	DBCount     int32 `json:"conn_cnt"`
	DataCount   int32 `json:"data_cnt"`
	IndexStart  int64 `json:"index_begin"`
	IndexEnd    int64 `json:"index_end"`
	IndexOffset int64 `json:"index_offset"`
	DataLen     int32 `json:"date_len"`

	Actions []string `json:"actions"`
}

var (
	sRespChan     = chanx.NewUnboundedChan[asyn_msg.RespInf](10240)
	sDBs          []*DB
	sConfig       = Config{}
	sProcessCount = int32(0)
)

const (
	ACTION_INSERT = "insert"
	ACTION_UPDATE = "update"
	ACTION_UPBASE = "upbase"
	ACTION_DELETE = "delete"
	ACTION_QUERY  = "query"

	server_id = 400
)

func monitor(t int64, wg *sync.WaitGroup, a string) {
	i := int32(0)
	for i < sConfig.DataCount {
		c := <-sRespChan.C()
		resp := c.(*RespDBOperator)
		if resp.Err != nil {
			log.Error("%s resp:%d err:%s", a, i, resp.Err.Error())
		}
		i++
		atomic.AddInt32(&sProcessCount, -1)
		if i%10000 == 0 {
			log.Info("%s Cnt:%d Cost:%d", a, i, time.Now().UnixMilli()-t)
		}
	}

	log.Info("%s Cnt:%d Cost:%d", a, i, time.Now().UnixMilli()-t)
	wg.Done()
}

func testInsert(t *testing.T) {
	log.Info("Action insert run start....")
	// name := "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
	nameBytes := make([]byte, sConfig.DataLen)
	for i := 0; i < int(sConfig.DataLen)-1; i++ {
		nameBytes[i] = '0'
	}
	nameBytes[sConfig.DataLen-1] = '9'
	name := string(nameBytes)

	baseDataProto := &proto_db.UserBaseData{
		Guild: &proto_db.UserGuildSimple{
			Name: name[0:70],
		},
	}
	baseData, err := proto.Marshal(baseDataProto)
	if err != nil {
		t.Errorf("baseData err:%s", err.Error())
		return
	}

	dataProto := &proto_db.UserData{
		//Pay: &proto_db.Pay{
		//	Orders: []*proto_db.PayOrder{
		//		{
		//			PayId: name,
		//		},
		//	},
		//},
	}
	data, err := proto.Marshal(dataProto)
	if err != nil {
		t.Errorf("data err:%s", err.Error())
		return
	}

	args := []interface{}{
		uint64(0), baseData, data, "", int32(1), int64(2),
		int64(3), int64(4), "", "", int32(server_id),
	}

	var dbArgs []*proto_db.DBArgs
	for _, arg := range args {
		dbArg := &proto_db.DBArgs{}
		switch arg.(type) {
		case int32:
			dbArg.ArgsType = proto_db.DB_ARGS_TYPE_D_A_T_INT
			dbArg.Args = []byte(strconv.FormatInt(int64(arg.(int32)), 10))
		case uint32:
			dbArg.ArgsType = proto_db.DB_ARGS_TYPE_D_A_T_INT
			dbArg.Args = []byte(strconv.FormatInt(int64(arg.(uint32)), 10))
		case int64:
			dbArg.ArgsType = proto_db.DB_ARGS_TYPE_D_A_T_BIGINT
			dbArg.Args = []byte(strconv.FormatInt(arg.(int64), 10))
		case uint64:
			dbArg.ArgsType = proto_db.DB_ARGS_TYPE_D_A_T_BIGINT
			dbArg.Args = []byte(strconv.FormatInt(int64(arg.(uint64)), 10))
		case string:
			dbArg.ArgsType = proto_db.DB_ARGS_TYPE_D_A_T_VARCHAR
			dbArg.Args = []byte(arg.(string))
		case []byte:
			dbArg.ArgsType = proto_db.DB_ARGS_TYPE_D_A_T_BLOB
			dbArg.Args = arg.([]byte)
		default:
			log.Error("doDbQuery not found type %v", arg)
		}
		dbArgs = append(dbArgs, dbArg)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	startTime := time.Now().UnixMilli()
	go monitor(startTime, wg, ACTION_INSERT)

	index := 0
	for i := 0; i < int(sConfig.DataCount); i++ {
		cnt := atomic.LoadInt32(&sProcessCount)
		if cnt > sConfig.DBCount*50 {
			time.Sleep(time.Millisecond * 500)
			log.Info("too much %d sleep 500ms to %d", cnt, atomic.LoadInt32(&sProcessCount))
		}
		uid := int64(i*int(sConfig.IndexOffset))%int64(sConfig.IndexEnd-sConfig.IndexStart) + int64(sConfig.IndexStart)
		sqlAs := make([]*proto_db.DBArgs, len(dbArgs))
		sqlAs[0] = &proto_db.DBArgs{
			ArgsType: proto_db.DB_ARGS_TYPE_D_A_T_BIGINT,
			Args:     []byte(strconv.FormatInt(uid, 10)),
		}
		for k := 1; k < len(dbArgs); k++ {
			sqlAs[k] = dbArgs[k]
		}
		req := &ReqDBOperator{
			Sql:  fmt.Sprintf("insert into %s (id,baseData,data,name,level,create_time,last_login_time,last_logout_time,account_id,channel,serverId) values (?,?,?,?,?,?,?,?,?,?,?)", sConfig.DBTableName),
			Op:   proto_db.DB_OPERATOR_DB_OP_INSERT,
			Args: sqlAs,
		}
		req.SetFcId(uint64(uid))
		sDBs[index].reqChan.Put(req)
		index = (index + 1) % int(sConfig.DBCount)
		atomic.AddInt32(&sProcessCount, 1)
	}

	log.Info("Action insert main end")
	wg.Wait()
	log.Info("Action insert main wait Ok")

	// for i := 0; i < int(sConfig.DBCount); i++ {
	// 	dbs[i].close()
	// }
	// log.Info("db close Ok")
	// log.Flush()
	// log.Close()
}

func testQuery(t *testing.T) {
	log.Info("Action query run start....")
	wg := &sync.WaitGroup{}
	wg.Add(1)
	startTime := time.Now().UnixMilli()
	go monitor(startTime, wg, ACTION_QUERY)

	index := 0
	for i := 0; i < int(sConfig.DataCount); i++ {
		cnt := atomic.LoadInt32(&sProcessCount)
		if cnt > sConfig.DBCount*50 {
			time.Sleep(time.Millisecond * 500)
			log.Info("too much %d sleep 500ms to %d", cnt, atomic.LoadInt32(&sProcessCount))
		}
		uid := int64(i*int(sConfig.IndexOffset)+int(sConfig.IndexStart)) % int64(sConfig.IndexEnd)
		sqlAs := make([]*proto_db.DBArgs, 1)
		sqlAs[0] = &proto_db.DBArgs{
			ArgsType: proto_db.DB_ARGS_TYPE_D_A_T_BIGINT,
			Args:     []byte(strconv.FormatInt(uid, 10)),
		}
		req := &ReqDBOperator{
			Sql:  fmt.Sprintf("select data from %s where id=?", sConfig.DBTableName),
			Op:   proto_db.DB_OPERATOR_DB_OP_SELECT,
			Args: sqlAs,
		}
		req.SetFcId(uint64(uid))
		sDBs[index].reqChan.Put(req)
		index = (index + 1) % int(sConfig.DBCount)
		atomic.AddInt32(&sProcessCount, 1)
	}

	log.Info("Action query main end")
	wg.Wait()
	log.Info("Action query main wait Ok")

	// for i := 0; i < int(sConfig.DBCount); i++ {
	// 	dbs[i].close()
	// }
	// log.Info("db close Ok")
	// log.Flush()
	// log.Close()
}

func testUpdate(t *testing.T) {
	log.Info("Action update run start....")
	// name := "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
	nameBytes := make([]byte, sConfig.DataLen)
	for i := 0; i < int(sConfig.DataLen)-1; i++ {
		nameBytes[i] = '1'
	}
	nameBytes[sConfig.DataLen-1] = '8'
	name := string(nameBytes)

	baseDataProto := &proto_db.UserBaseData{
		Guild: &proto_db.UserGuildSimple{
			Name: name,
		},
	}
	baseData, err := proto.Marshal(baseDataProto)
	if err != nil {
		t.Errorf("baseData err:%s", err.Error())
		return
	}

	dataProto := &proto_db.UserData{}
	data, err := proto.Marshal(dataProto)
	if err != nil {
		t.Errorf("data err:%s", err.Error())
		return
	}

	args := []interface{}{
		baseData, data, "", int32(1), int64(2),
		int64(3), int64(4), "", "", int32(server_id),
		uint64(0),
	}

	var dbArgs []*proto_db.DBArgs
	for _, arg := range args {
		dbArg := &proto_db.DBArgs{}
		switch arg.(type) {
		case int32:
			dbArg.ArgsType = proto_db.DB_ARGS_TYPE_D_A_T_INT
			dbArg.Args = []byte(strconv.FormatInt(int64(arg.(int32)), 10))
		case uint32:
			dbArg.ArgsType = proto_db.DB_ARGS_TYPE_D_A_T_INT
			dbArg.Args = []byte(strconv.FormatInt(int64(arg.(uint32)), 10))
		case int64:
			dbArg.ArgsType = proto_db.DB_ARGS_TYPE_D_A_T_BIGINT
			dbArg.Args = []byte(strconv.FormatInt(arg.(int64), 10))
		case uint64:
			dbArg.ArgsType = proto_db.DB_ARGS_TYPE_D_A_T_BIGINT
			dbArg.Args = []byte(strconv.FormatInt(int64(arg.(uint64)), 10))
		case string:
			dbArg.ArgsType = proto_db.DB_ARGS_TYPE_D_A_T_VARCHAR
			dbArg.Args = []byte(arg.(string))
		case []byte:
			dbArg.ArgsType = proto_db.DB_ARGS_TYPE_D_A_T_BLOB
			dbArg.Args = arg.([]byte)
		default:
			log.Error("doDbQuery not found type %v", arg)
		}
		dbArgs = append(dbArgs, dbArg)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	startTime := time.Now().UnixMilli()
	go monitor(startTime, wg, ACTION_UPDATE)

	index := 0
	for i := 0; i < int(sConfig.DataCount); i++ {
		cnt := atomic.LoadInt32(&sProcessCount)
		if cnt > sConfig.DBCount*50 {
			time.Sleep(time.Millisecond * 500)
			log.Info("too much %d sleep 500ms to %d", cnt, atomic.LoadInt32(&sProcessCount))
		}
		uid := int64(i*int(sConfig.IndexOffset)+int(sConfig.IndexStart)) % int64(sConfig.IndexEnd)
		sqlAs := make([]*proto_db.DBArgs, len(dbArgs))

		for k := 0; k < len(dbArgs)-1; k++ {
			sqlAs[k] = dbArgs[k]
		}
		sqlAs[len(dbArgs)-1] = &proto_db.DBArgs{
			ArgsType: proto_db.DB_ARGS_TYPE_D_A_T_BIGINT,
			Args:     []byte(strconv.FormatInt(uid, 10)),
		}

		req := &ReqDBOperator{
			Sql:  fmt.Sprintf("update %s set baseData=?,data=?,name=?,level=?,create_time=?,last_login_time=?,last_logout_time=?,account_id=?,channel=?,serverId=? where id=?", sConfig.DBTableName),
			Op:   proto_db.DB_OPERATOR_DB_OP_UPDATE,
			Args: sqlAs,
		}
		req.SetFcId(uint64(uid))
		sDBs[index].reqChan.Put(req)
		index = (index + 1) % int(sConfig.DBCount)
		atomic.AddInt32(&sProcessCount, 1)
	}

	log.Info("Action update main end")
	wg.Wait()
	log.Info("Action update main wait Ok")

	// for i := 0; i < int(sConfig.DBCount); i++ {
	// 	dbs[i].close()
	// }
	// log.Info("db close Ok")
	// log.Flush()
	// log.Close()
}

func testUpdateBase(t *testing.T) {
	log.Info("Action update Base run start....")
	// name := "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
	nameBytes := make([]byte, sConfig.DataLen)
	for i := 0; i < int(sConfig.DataLen)-1; i++ {
		nameBytes[i] = '1'
	}
	nameBytes[sConfig.DataLen-1] = '8'
	name := string(nameBytes)

	baseDataProto := &proto_db.UserBaseData{
		Guild: &proto_db.UserGuildSimple{
			Name: name,
		},
	}

	baseData, err := proto.Marshal(baseDataProto)
	if err != nil {
		t.Errorf("baseData err:%s", err.Error())
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	startTime := time.Now().UnixMilli()
	go monitor(startTime, wg, ACTION_UPDATE)

	index := 0
	for i := 0; i < int(sConfig.DataCount); i++ {
		cnt := atomic.LoadInt32(&sProcessCount)
		if cnt > sConfig.DBCount*50 {
			time.Sleep(time.Millisecond * 500)
			log.Info("too much %d sleep 500ms to %d", cnt, atomic.LoadInt32(&sProcessCount))
		}
		uid := int64(i*int(sConfig.IndexOffset)+int(sConfig.IndexStart)) % int64(sConfig.IndexEnd)
		sqlAs := make([]*proto_db.DBArgs, 2)

		sqlAs[0] = &proto_db.DBArgs{
			ArgsType: proto_db.DB_ARGS_TYPE_D_A_T_BLOB,
			Args:     baseData,
		}
		sqlAs[1] = &proto_db.DBArgs{
			ArgsType: proto_db.DB_ARGS_TYPE_D_A_T_BIGINT,
			Args:     []byte(strconv.FormatInt(uid, 10)),
		}

		req := &ReqDBOperator{
			Sql:  fmt.Sprintf("update %s set baseData=? where id=?", sConfig.DBTableName),
			Op:   proto_db.DB_OPERATOR_DB_OP_UPDATE,
			Args: sqlAs,
		}
		req.SetFcId(uint64(uid))
		sDBs[index].reqChan.Put(req)
		index = (index + 1) % int(sConfig.DBCount)
		atomic.AddInt32(&sProcessCount, 1)
	}

	log.Info("Action update Base main end")
	wg.Wait()
	log.Info("Action update Base main wait Ok")

	// for i := 0; i < int(sConfig.DBCount); i++ {
	// 	dbs[i].close()
	// }
	// log.Info("db close Ok")
	// log.Flush()
	// log.Close()
}

func testDelete(t *testing.T) {
	log.Info("Action delete run start....")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	startTime := time.Now().UnixMilli()
	go monitor(startTime, wg, ACTION_DELETE)

	index := 0
	for i := 0; i < int(sConfig.DataCount); i++ {
		cnt := atomic.LoadInt32(&sProcessCount)
		if cnt > sConfig.DBCount*50 {
			time.Sleep(time.Millisecond * 500)
			log.Info("too much %d sleep 500ms to %d", cnt, atomic.LoadInt32(&sProcessCount))
		}
		uid := int64(i*int(sConfig.IndexOffset)+int(sConfig.IndexStart)) % int64(sConfig.IndexEnd)
		sqlAs := make([]*proto_db.DBArgs, 1)
		sqlAs[0] = &proto_db.DBArgs{
			ArgsType: proto_db.DB_ARGS_TYPE_D_A_T_BIGINT,
			Args:     []byte(strconv.FormatInt(uid, 10)),
		}

		req := &ReqDBOperator{
			Sql:  fmt.Sprintf("delete from %s where id=?", sConfig.DBTableName),
			Op:   proto_db.DB_OPERATOR_DB_OP_DELETE,
			Args: sqlAs,
		}
		req.SetFcId(uint64(uid))
		sDBs[index].reqChan.Put(req)
		index = (index + 1) % int(sConfig.DBCount)
		atomic.AddInt32(&sProcessCount, 1)
	}

	log.Info("Action delete main end")
	wg.Wait()
	log.Info("Action delete main wait Ok")

	// for i := 0; i < int(sConfig.DBCount); i++ {
	// 	dbs[i].close()
	// }
	// log.Info("db close Ok")
	// log.Flush()
	// log.Close()
}

func TestDB(t *testing.T) {
	log.LoadConfiguration("log_home.xml", nil)
	readFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Error("config err:%s", err.Error())
		return
	}
	err = json.Unmarshal(readFile, &sConfig)
	if err != nil {
		log.Error("json config err:%s", err.Error())
		return
	}
	if sConfig.IndexOffset == 0 {
		sConfig.IndexOffset = 1
	}

	log.Info("config load OK %v", sConfig)
	sDBs = make([]*DB, sConfig.DBCount)

	dbCnf := db.MysqlCfg{
		Addr:     sConfig.DBAddr,
		User:     sConfig.DBUser,
		Passwd:   sConfig.DBPassword,
		Database: sConfig.DBDatabase,
	}

	for i := 0; i < int(sConfig.DBCount); i++ {
		db := newDB(&dbCnf, sRespChan)
		db.init()
		sDBs[i] = db
	}

	log.Info("TestDB db cnt:%d init OK", sConfig.DBCount)
	for _, a := range sConfig.Actions {
		switch a {
		case ACTION_DELETE:
			testDelete(t)
		case ACTION_INSERT:
			testInsert(t)
		case ACTION_UPDATE:
			testUpdate(t)
		case ACTION_QUERY:
			testQuery(t)
		case ACTION_UPBASE:
			testUpdateBase(t)
		}
	}
	log.Info("TestDB closing...")

	for i := 0; i < int(sConfig.DBCount); i++ {
		sDBs[i].close()
	}
	log.Info("db close Ok")
	log.Flush()
	log.Close()
}

func BytesToInt64(buf []byte) int64 {
	number, err := strconv.ParseInt(string(buf), 10, 64)
	if err != nil {
		log.Error("err is %s", err.Error())
		return math.MaxInt64
	}
	return number
}

func TestGetProp(t *testing.T) {
	log.LoadConfiguration("log_home.xml", nil)
	readFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Error("config err:%s", err.Error())
		return
	}
	err = json.Unmarshal(readFile, &sConfig)
	if err != nil {
		log.Error("json config err:%s", err.Error())
		return
	}

	log.Info("config load OK %v", sConfig)
	dbCnf := db.MysqlCfg{
		Addr:     sConfig.DBAddr,
		User:     sConfig.DBUser,
		Passwd:   sConfig.DBPassword,
		Database: sConfig.DBDatabase,
	}

	db := newDB(&dbCnf, sRespChan)
	db.init()

	log.Info("db init ok")

	req := &ReqDBOperator{
		Sql:  fmt.Sprintf("select * from %s", sConfig.DBTableName),
		Op:   proto_db.DB_OPERATOR_DB_OP_SELECT,
		Args: nil,
	}
	db.reqChan.Put(req)

	c := <-sRespChan.C()
	resp := c.(*RespDBOperator)
	if resp.Err != nil {
		log.Error("err:%s", resp.Err.Error())
	} else {
		var userIds []uint64
		for _, arrs := range resp.Ret {
			loginTime := BytesToInt64(arrs["last_logout_time"])
			databytes := arrs["data"]
			var usr proto_db.UserData
			proto.Unmarshal(databytes, &usr)
			//ua := usr.GetUserActivity().GetSignInActivities()
			//if ua == nil {
			//	continue
			//}
			//a := ua[1]
			//if a == nil {
			//	continue
			//}
			//if a.SignInTimes > 1 {
			//	continue
			//}
			//if !a.AcceptTodayReward {
			//	continue
			//}
			//if usr.GetTask().MainTaskRewardChapter < 6 {
			//	continue
			//}
			if loginTime < 1688054400000 {
				continue
			}
			userIds = append(userIds, uint64(BytesToInt64(arrs["id"])))
		}
		log.Info("ddddddddddd: %v", userIds)
	}

	log.Info("TestDB closing...")

	db.close()
	log.Info("db close Ok")
	log.Flush()
	log.Close()
}
