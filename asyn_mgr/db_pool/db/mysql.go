package db

import (
	"database/sql"
	"google.golang.org/protobuf/proto"
	"px/common/db"
	"px/proto/proto_db"
	"strconv"
	"sync/atomic"
)

type (
	MysqlCfg struct {
		Addr     string
		User     string
		Passwd   string
		Database string
	}
	Mysql struct {
		DbImp   *db.DB
		working int32
	}
)

const (
	userMod uint64 = 16
)

func NewMysql(cfg *MysqlCfg) (*Mysql, error) {
	dbClient, err := db.Connect(cfg.User, cfg.Passwd, cfg.Addr, cfg.Database)
	if nil != err {
		return nil, err
	}
	err = dbClient.Ping()
	if err != nil {
		return nil, err
	}
	return &Mysql{DbImp: dbClient}, nil
}

func (m *Mysql) LoadUser(id uint64) (name string, data *proto_db.UserData, err error) {
	atomic.AddInt32(&m.working, 1)
	defer atomic.AddInt32(&m.working, -1)
	tbName := userTable(id)
	rows, e := m.DbImp.Client.Query("SELECT name, data FROM "+tbName+" WHERE id=?", id)
	if nil != e {
		err = e
		return
	}
	var dbEncodeData []byte

	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&name, &dbEncodeData)
		if nil != err {
			return
		}
	} else {
		err = NoUserError
		return
	}

	dbData := &proto_db.UserData{}

	err = proto.Unmarshal(dbEncodeData, dbData)
	if err != nil {
		err = UserMarshalError
		return
	}
	return name, dbData, nil
}

func (m *Mysql) LoadUserByName(name string) (data *proto_db.UserData, err error) {
	atomic.AddInt32(&m.working, 1)
	defer atomic.AddInt32(&m.working, -1)
	tbName := userTable(0)
	rows, e := m.DbImp.Client.Query("SELECT name, data FROM "+tbName+" WHERE name=?", name)
	if nil != e {
		err = e
		return
	}
	var dbEncodeData []byte

	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&name, &dbEncodeData)
		if nil != err {
			return
		}
	} else {
		err = NoUserError
		return
	}

	dbData := &proto_db.UserData{}

	err = proto.Unmarshal(dbEncodeData, dbData)
	if err != nil {
		err = UserMarshalError
		return
	}
	return dbData, nil
}

func (m *Mysql) LoadAllUsers() ([]*proto_db.UserData, error) {
	atomic.AddInt32(&m.working, 1)
	defer atomic.AddInt32(&m.working, -1)
	tbName := userTable(0)
	rows, e := m.DbImp.Client.Query("SELECT id, name, data, UserHeroData, TroopPlans, underground_city, breed, nursery, school FROM " + tbName + "")
	if nil != e {
		return nil, e
	}

	defer rows.Close()

	var users []*proto_db.UserData
	for rows.Next() {
		dbUser, err := parseDbUserFromDbData(rows)
		if err != nil {
			panic(err)
		}
		users = append(users, dbUser)
	}

	return users, nil
}

func (m *Mysql) LoadGMail(opSeq uint64) (out [][]byte, seq uint64, err error) {
	rows, e := m.DbImp.Client.Query("SELECT opseq, data FROM NC_GMAIL_DATA WHERE opseq > ?", opSeq)
	if nil != e {
		err = e
		return
	}
	defer rows.Close()
	for rows.Next() {
		var (
			tmpSeq uint64
			by     []byte
		)
		err = rows.Scan(&tmpSeq, &by)
		if nil != err {
			return
		}
		if opSeq < tmpSeq {
			opSeq = tmpSeq
		}
		out = append(out, by)
	}
	return out, opSeq, nil
}

func (m *Mysql) InsertOrUpdateUser(insert bool, uid uint64, name string, data *proto_db.UserData) (err error) {
	//atomic.AddInt32(&m.working, 1)
	//defer atomic.AddInt32(&m.working, -1)
	//var (
	//	baseData        []byte
	//	hero            []byte
	//	troop           []byte
	//	undergroundCity []byte
	//	breed           []byte
	//	nursery         []byte
	//	school          []byte
	//)
	//if data != nil {
	//	baseData, err = proto.Marshal(data.GetStatisticData())
	//	if data.GetStatisticData() != nil && err != nil {
	//		return
	//	}
	//	hero, err = proto.Marshal(data.GetHero())
	//	if data.GetHero() != nil && err != nil {
	//		return
	//	}
	//	troop, err = proto.Marshal(data.GetTroops())
	//	if data.GetTroops() != nil && err != nil {
	//		return
	//	}
	//	undergroundCity, err = proto.Marshal(data.GetCity())
	//	if data.GetCity() != nil && err != nil {
	//		return
	//	}
	//}
	//tbName := userTable(uid)
	//if insert {
	//	_, err = m.DbImp.Client.Exec("INSERT INTO "+tbName+
	//		" (id, name, data, UserHeroData, TroopPlans, underground_city, breed, nursery, school)"+
	//		" VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
	//		uid,
	//		name,
	//		baseData,
	//		hero,
	//		troop,
	//		undergroundCity,
	//		breed,
	//		nursery,
	//		school,
	//	)
	//} else {
	//	_, err = m.DbImp.Client.Exec("UPDATE "+tbName+
	//		" SET name=?, data=?, UserHeroData=?, TroopPlans=?, underground_city=?, breed=?, nursery=?, school=? WHERE id=?",
	//		name,
	//		baseData,
	//		hero,
	//		troop,
	//		undergroundCity,
	//		breed,
	//		nursery,
	//		school,
	//		uid)
	//}
	return
}

func (m *Mysql) SavingCenterData(id int, data []byte) (err error) {
	_, err = m.DbImp.Client.Exec("INSERT INTO center (id, Data) VALUES(?, ?) ON DUPLICATE KEY UPDATE Data=?",
		id, data, data)
	return
}

func (m *Mysql) LoadingCenterData(id int) (data []byte, err error) {
	rows, e := m.DbImp.Client.Query("SELECT id, data FROM center WHERE id=?", id)
	if nil != e {
		err = e
		return
	}
	defer rows.Close()
	if rows.Next() {
		var (
			uid uint64
		)
		err = rows.Scan(&uid, &data)
		if nil != err {
			return
		}
	}
	return
}

func (m *Mysql) LogData(data []byte) (err error) {
	atomic.AddInt32(&m.working, 1)
	defer atomic.AddInt32(&m.working, -1)
	_, err = m.DbImp.Client.Exec("INSERT INTO data_record(data) VALUES(?)",
		data)
	return
}

func (m *Mysql) CloseImmediately() {
	m.DbImp.Close()
}

func (m *Mysql) CloseLazy() {
	for atomic.LoadInt32(&m.working) != 0 {
		continue
	}
	m.DbImp.Close()
}

func extractParams(args []*proto_db.DBArgs) ([]interface{}, error) {
	var dbArgs []interface{}
	for _, arg := range args {
		switch arg.GetArgsType() {
		case proto_db.DB_ARGS_TYPE_D_A_T_INT:
			dbArg, e := strconv.Atoi(string(arg.GetArgs()))
			if e != nil {
				return nil, e
			}
			dbArgs = append(dbArgs, dbArg)
		case proto_db.DB_ARGS_TYPE_D_A_T_BIGINT:
			dbArg, e := strconv.Atoi(string(arg.GetArgs()))
			if e != nil {
				return nil, e
			}
			dbArgs = append(dbArgs, dbArg)
		case proto_db.DB_ARGS_TYPE_D_A_T_VARCHAR:
			dbArg := string(arg.GetArgs())
			dbArgs = append(dbArgs, dbArg)
		case proto_db.DB_ARGS_TYPE_D_A_T_BLOB:
			dbArgs = append(dbArgs, arg.GetArgs())
		}
	}
	return dbArgs, nil
}

func (m *Mysql) DBOperator(sql string, op proto_db.DB_OPERATOR, args []*proto_db.DBArgs) (ret []map[string][]byte, err error) {
	atomic.AddInt32(&m.working, 1)
	defer atomic.AddInt32(&m.working, -1)
	var dbArgs []interface{}
	dbArgs, err = extractParams(args)
	if err != nil {
		return nil, err
	}
	if op == proto_db.DB_OPERATOR_DB_OP_SELECT {
		rows, e := m.DbImp.Query(sql, dbArgs...)
		if nil != e {
			return nil, e
		}

		defer rows.Close()

		ret, err = Rows2maps(rows)
	} else {
		_, err = m.DbImp.Exec(sql, dbArgs...)
	}

	return
}

func userTable(uid uint64) string {
	//mod := uid % userMod
	//return "user_" + strconv.Itoa(int(mod))
	return "user"
}

func worldUnitTable(uid uint64) string {
	//mod := uid % userMod
	//return "user_" + strconv.Itoa(int(mod))
	return "worldUnit"
}

func parseDbUserFromDbData(rows *sql.Rows) (dbUser *proto_db.UserData, err error) {
	//dbUser = &proto_db.UserData{
	//	BaseData: &proto_db.UserDbData{},
	//	Hero: &proto_db.UserHero{},
	//	Troop: &proto_db.TroopPlans{},
	//	City: &proto_db.UserCity{},
	//	BreedRoom:&proto_db.BreedRoom{},
	//	NurseryCenter: &proto_db.NurseryCenter{},
	//	School: &proto_db.School{},
	//}
	//var (
	//	baseData    []byte
	//	heroData    []byte
	//	troopData   []byte
	//	cityData    []byte
	//	breedData   []byte
	//	nurseryData []byte
	//	schoolData  []byte
	//)
	//
	//err = rows.Scan(&dbUser.Uid,
	//	&dbUser.Name,
	//	&baseData,
	//	&heroData,
	//	&troopData,
	//	&cityData,
	//	&breedData,
	//	&nurseryData,
	//	&schoolData)
	//if err != nil {
	//	err = fmt.Errorf(fmt.Sprintf("rows scan error %v", err))
	//	return
	//}
	//err = proto.Unmarshal(baseData, dbUser.BaseData)
	//if err != nil {
	//	err = fmt.Errorf(fmt.Sprintf("Unmarshal error %v", err))
	//	return
	//}
	//err = proto.Unmarshal(heroData, dbUser.Hero)
	//if err != nil {
	//	err = fmt.Errorf(fmt.Sprintf("Unmarshal error %v", err))
	//	return
	//}
	//err = proto.Unmarshal(troopData, dbUser.Troop)
	//if err != nil {
	//	err = fmt.Errorf(fmt.Sprintf("Unmarshal error %v", err))
	//	return
	//}
	//err = proto.Unmarshal(cityData, dbUser.City)
	//if err != nil {
	//	err = fmt.Errorf(fmt.Sprintf("Unmarshal error %v", err))
	//	return
	//}
	//err = proto.Unmarshal(breedData, dbUser.BreedRoom)
	//if err != nil {
	//	err = fmt.Errorf(fmt.Sprintf("Unmarshal error %v", err))
	//	return
	//}
	//err = proto.Unmarshal(nurseryData, dbUser.NurseryCenter)
	//if err != nil {
	//	err = fmt.Errorf(fmt.Sprintf("Unmarshal error %v", err))
	//	return
	//}
	//err = proto.Unmarshal(schoolData, dbUser.School)
	//if err != nil {
	//	err = fmt.Errorf(fmt.Sprintf("Unmarshal error %v", err))
	//	return
	//}
	return
}

//func parseDbUnitFromDbData(rows *sql.Rows) (dbUnit *proto_db.DBUnit, err error) {
//	dbUnit = &proto_db.DBUnit{}
//	var (
//		unitData []byte
//	)
//
//	err = rows.Scan(&dbUnit.Guid, &unitData)
//	if err != nil {
//		err = fmt.Errorf(fmt.Sprintf("rows scan error %v", err))
//		return
//	}
//	err = proto.Unmarshal(unitData, dbUnit)
//	if err != nil {
//		err = fmt.Errorf(fmt.Sprintf("Unmarshal error %v", err))
//		return
//	}
//	return
//}
//
//func (m *Mysql) LoadAllWorldUnits() (dbUnits []*proto_db.DBUnit, err error) {
//	atomic.AddInt32(&m.working, 1)
//	defer atomic.AddInt32(&m.working, -1)
//	tbName := worldUnitTable(0)
//	rows, e := m.DbImp.Client.Query("SELECT guid, data FROM " + tbName + "")
//	if nil != e {
//		return nil, e
//	}
//
//	defer rows.Close()
//
//	dbUnits = []*proto_db.DBUnit{}
//	for rows.Next() {
//		dbUnit, err := parseDbUnitFromDbData(rows)
//		if err != nil {
//			panic(err)
//		}
//		dbUnits = append(dbUnits, dbUnit)
//	}
//
//	return
//}

//func (m *Mysql) SaveOrDeleteWorldUnit(save bool, data *proto_db.DBUnit) (err error) {
//	atomic.AddInt32(&m.working, 1)
//	defer atomic.AddInt32(&m.working, -1)
//
//	tbName := worldUnitTable(0)
//	if save {
//		unitData, e := proto.Marshal(data)
//		if data != nil && e != nil {
//			return e
//		}
//
//		_, err = m.DbImp.Client.Exec("REPLACE INTO "+tbName+
//			" (guid, data)"+
//			" VALUES (?, ?)",
//			data.Guid,
//			unitData,
//		)
//	} else {
//		_, err = m.DbImp.Client.Exec("delete from "+tbName+
//			" WHERE guid = ?",
//			data.Guid)
//	}
//	return
//}
