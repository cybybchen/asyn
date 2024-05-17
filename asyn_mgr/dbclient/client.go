package dbclient

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"px/common/message"
	"px/common/net-engine/trans-db"
	"px/define"
	"px/proto/proto_db"
	"px/svr"
	"px/utils"
	"px/utils/chanx"
	"reflect"
	"strconv"
	"time"

	"gitlab.sunborngame.com/base/log"
	"gitlab.sunborngame.com/base/sunnet"
	"google.golang.org/protobuf/proto"
)

type (
	DbReq struct {
		Op   proto_db.DB_OPERATOR
		Sql  string
		Args []interface{}
	}

	DbRet struct {
		isFinish bool
		fcid     uint64
		ErrMsg   string
		Datas    []*proto_db.DBData
	}

	DbFc struct {
		fcId        uint64
		cb          func(string, []*proto_db.DBData)
		expiredTime int64
	}

	DBClient struct {
		c            *svr.DbClientSvr
		DbMessage    *message.MsgMgr
		fcId         uint64
		mulQ2OriginQ map[uint64]uint64
		mulQWaits    map[uint64][]*DbRet
		mulQFcs      map[uint64]func(uint64, string, []*proto_db.DBData)
		fcs          *utils.SortedSet[uint64, *DbFc]

		respChan *chanx.UnboundedChan[any]
	}
)

const (
	defaultConnTimeout = 10
)

func (this *DbFc) Key() uint64 {
	return this.fcId
}

func (this *DbFc) Less(other utils.SortedSetData[uint64]) bool {
	o2 := other.(*DbFc)
	return this.expiredTime < o2.expiredTime
}

func (this *DbFc) callBack(errMsg string, datas []*proto_db.DBData) {
	this.cb(errMsg, datas)
}

func newDbClient(reconn int, addr string) *DBClient {
	d := &DBClient{
		fcs:          utils.NewSortedSet[uint64, *DbFc](),
		mulQ2OriginQ: make(map[uint64]uint64),
		mulQWaits:    make(map[uint64][]*DbRet),
		mulQFcs:      make(map[uint64]func(uint64, string, []*proto_db.DBData)),
		respChan:     chanx.NewUnboundedChan[any](MessageChanCap),
	}
	d.c = d.newDbClientSvr(reconn, addr)
	d.DbMessage = NewDbMessage(d)

	return d
}

func (c *DBClient) newDbClientSvr(reconn int, addr string) *svr.DbClientSvr {
	var dbSvr = &svr.DbClientSvr{}
	reconnDura := time.Duration(reconn) * time.Second
	dbSvr.SetConn(sunnet.NewTcpConnector(
		addr,
		trans_db.NewTransmitter(binary.BigEndian),
		c,
		reconnDura,
	))

	return dbSvr
}

func NewDbMessage(c *DBClient) *message.MsgMgr {
	mgr := message.NewMsgMgr(define.MsgIdBeginDB, define.MsgIdEndDB, int64(define.ClientMsgExecuteTimeMax))
	mgr.Register(uint32(proto_db.MSG_ID_MSG_DL_DBOperator), c)
	return mgr
}

func (c *DBClient) Start(timeoutSeconds int) {
	if timeoutSeconds <= 0 {
		timeoutSeconds = defaultConnTimeout
	}
	c.c.StartWithCtx(timeoutSeconds)
}

func (c *DBClient) SendEvent(msg interface{}) {
	c.respChan.Put(msg)
}

func (c *DBClient) handleResp(msg interface{}) {
	switch ty := msg.(type) {
	case *sunnet.EventNormal:
		pk, ok := ty.M.(*trans_db.Packet)
		if !ok {
			return
		}
		_ = c.DbMessage.Dispatcher(message.ClientID{GateSessId: ty.S.Sid(), ClientSessId: ty.S.Sid()}, uint32(pk.T), pk.V)
	case *sunnet.EventSessionAdd:
		log.Info("dbproxy connected:%d", ty.S.Sid())
		c.c.Cancel()
	case *sunnet.EventSessionClosed:
		log.Error("dbproxy close:%d err:%s", ty.S.Sid(), ty.E.Error())
	}
}

func (c *DBClient) Name() string { return "MSG_ID_MSG_DL_LoadDBData" }

func (c *DBClient) Execute(id message.ClientID, data message.Messager) error {
	msg, ok := data.(*proto_db.DL_DBOperator)
	if !ok {
		return errors.New("Marsh error" + reflect.TypeOf(data).Name())
	}

	c.HandleDBCall(msg)

	return nil
}

func (c *DBClient) nextFcId() uint64 {
	if c.fcId == math.MaxUint64 {
		c.fcId = 0
	}
	c.fcId++
	return c.fcId
}

func (c *DBClient) doDbQuery(sql string, op proto_db.DB_OPERATOR, args []interface{}, fc func(string, []*proto_db.DBData)) uint64 {
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
			//目前mysql用的是mediumblob，大小不能超过（1<<24）-1
			var bys = arg.([]byte)
			if !CheckMediumBlob(bys) {
				var e = fmt.Errorf("CheckMediumBlob failed, blob length=%v", len(bys))
				log.Error(e.Error())
				fc(e.Error(), nil)
				return 0
			}
			dbArg.ArgsType = proto_db.DB_ARGS_TYPE_D_A_T_BLOB
			dbArg.Args = bys
		default:
			log.Error("doDbQuery not found type %v", arg)
		}
		dbArgs = append(dbArgs, dbArg)
	}
	var msg = &proto_db.LD_DBOperator{
		FcId: c.nextFcId(),
		Sql:  sql,
		Op:   op,
		Args: dbArgs,
	}
	c.fcs.Push(&DbFc{
		fcId:        msg.FcId,
		cb:          fc,
		expiredTime: utils.NowUnixMilli() + DbOpTimeout,
	})
	c.c.Send(msg)
	return msg.FcId
}

func (c *DBClient) doMulDbQuery(sql string, op proto_db.DB_OPERATOR, args []interface{}, fc func(uint64, string, []*proto_db.DBData)) uint64 {
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
	var msg = &proto_db.LD_DBOperator{
		FcId: c.nextFcId(),
		Sql:  sql,
		Op:   op,
		Args: dbArgs,
	}
	c.mulQFcs[msg.FcId] = fc

	c.c.Send(msg)
	return msg.FcId
}

func (c *DBClient) DoSelect(sql string, args []interface{}, fc func(string, []*proto_db.DBData)) uint64 {
	return c.doDbQuery(sql, proto_db.DB_OPERATOR_DB_OP_SELECT, args, fc)
}

func (c *DBClient) DoUpdate(sql string, args []interface{}, fc func(string, []*proto_db.DBData)) uint64 {
	return c.doDbQuery(sql, proto_db.DB_OPERATOR_DB_OP_UPDATE, args, fc)
}

func (c *DBClient) DoInsert(sql string, args []interface{}, fc func(string, []*proto_db.DBData)) uint64 {
	return c.doDbQuery(sql, proto_db.DB_OPERATOR_DB_OP_INSERT, args, fc)
}

func (c *DBClient) DoDelete(sql string, args []interface{}, fc func(string, []*proto_db.DBData)) uint64 {
	return c.doDbQuery(sql, proto_db.DB_OPERATOR_DB_OP_DELETE, args, fc)
}

func (c *DBClient) DoReplace(sql string, args []interface{}, fc func(string, []*proto_db.DBData)) uint64 {
	return c.doDbQuery(sql, proto_db.DB_OPERATOR_DB_OP_REPLACE, args, fc)
}

func (c *DBClient) DoMulQuery(reqs []DbReq, fc func(ret []*DbRet)) {
	fcid := c.nextFcId()
	for _, req := range reqs {
		dbArgs, err := c.toDbArgs(req.Args)
		if err != nil {
			c.mulQWaits[fcid] = append(c.mulQWaits[fcid], &DbRet{
				isFinish: true,
				ErrMsg:   fmt.Sprintf("marshal arg err %s", err),
			})
			continue
		}
		qfcid := c.doMulDbQuery(req.Sql, proto_db.DB_OPERATOR(req.Op), dbArgs, func(fcid uint64, s string, data []*proto_db.DBData) {
			originFcId, ok := c.mulQ2OriginQ[fcid]
			if !ok {
				return
			}
			if len(c.mulQWaits[originFcId]) == 0 {
				return
			}
			for _, r := range c.mulQWaits[originFcId] {
				if r.fcid != fcid {
					continue
				}
				r.isFinish = true
				r.ErrMsg = s
				r.Datas = data
				delete(c.mulQ2OriginQ, fcid)
				break
			}
			for _, r := range c.mulQWaits[originFcId] {
				if !r.isFinish {
					return
				}
			}
			defer delete(c.mulQWaits, originFcId)
			fc(c.mulQWaits[originFcId])
		})
		c.mulQ2OriginQ[qfcid] = fcid
		c.mulQWaits[fcid] = append(c.mulQWaits[fcid], &DbRet{
			fcid:     qfcid,
			isFinish: false,
		})
	}
}

func (c *DBClient) tick(now int64) {
	for {
		dbFc, ok := c.fcs.Front()
		if !ok {
			break
		}
		if dbFc.expiredTime > now {
			break
		}

		dbFc.callBack(fmt.Sprintf("mysql timeout fcid:%d", dbFc.fcId), nil)
		c.fcs.Remove(dbFc)
	}
}

func (c *DBClient) HandleDBCall(msg *proto_db.DL_DBOperator) error {
	log.Info("[DL_DBOperator] handleResp, fcId=%d", msg.GetFcId())
	dbFc, ok := c.fcs.GetByKey(msg.FcId)
	if ok {
		dbFc.callBack(msg.ErrMsg, msg.Data)
		c.fcs.RemoveByKey(msg.FcId)
	} else {
		log.Warning("DL_DBOperator fcid:%d null", msg.FcId)
	}

	mulQFc, ok := c.mulQFcs[msg.FcId]
	if ok && mulQFc != nil {
		mulQFc(msg.FcId, msg.ErrMsg, msg.Data)
		delete(c.mulQFcs, msg.FcId)
	}

	return nil
}

func (c *DBClient) GetClientSvr() *svr.DbClientSvr {
	return c.c
}

func (c *DBClient) toDbArgs(args []interface{}) ([]interface{}, error) {
	dbargs := make([]interface{}, 0, len(args))
	for _, arg := range args {
		msg, ok := arg.(proto.Message)
		if !ok {
			dbargs = append(dbargs, arg)
			continue
		}
		pb, err := proto.Marshal(msg)
		if err != nil {
			log.Error("OnDbQuery arg marshal err %s", msg.ProtoReflect().Type())
			return nil, err
		}
		dbargs = append(dbargs, pb)
	}
	return dbargs, nil
}
