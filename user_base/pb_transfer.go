package user_base

import (
	"px/proto/proto_db"
	"px/proto/proto_object"
)

// UserBase funcs

func (this *UserBase) LoadBaseData(bd *proto_db.UserBaseData) {
	this.UserStatistics = NewUserStatisticsFromDb(bd.GetStatData())
	this.UserHead = NewUserHeadFromDb(bd.GetHead())
	this.Signature = bd.Signature
}

func (this *UserBase) PackUserBaseDbProto() *proto_db.UserBase {
	to := &proto_db.UserBase{
		Account:            this.Account,
		UserId:             this.Id,
		UserName:           this.Name,
		ServerId:           this.ServerId,
		Channel:            this.Channel,
		CreateTime:         this.CreateTime,
		LastLoginTime:      this.LastLoginTime,
		LastLogoutTime:     this.LastLogoutTime,
		LastDailyResetTime: this.LastDailyResetTime,
		BaseData:           this.PackUserBaseDataDbProto(),
		Level:              this.Level,
	}
	return to
}

func (this *UserBase) PackUserBaseDataDbProto() *proto_db.UserBaseData {
	return &proto_db.UserBaseData{
		Head:      this.UserHead.ToDbProto(),
		StatData:  this.UserStatistics.ToUserStatisticsDb(),
		Signature: this.Signature,
	}
}

func (this *UserBase) ToObjProto() *proto_object.UserBase {
	if this == nil {
		return &proto_object.UserBase{}
	}
	to := &proto_object.UserBase{
		Head:          this.UserHead.ToObjProto(),
		ServerId:      this.ServerId,
		Account:       this.Account,
		UserId:        this.Id,
		UserName:      this.Name,
		Level:         this.Level,
		StatisticData: this.ToUserStatisticsObj(),
		LastLoginTs:   this.LastLoginTime,
		IsOnline:      this.LastLoginTime > this.LastLogoutTime, // 底层没办法直接获取isonline，先以这种方式来获取
	}
	return to
}

// UserHead funcs

func NewUserHeadFromDb(data *proto_db.UserHead) *UserHead {
	to := &UserHead{
		HeadId:  data.GetHeadId(),
		FrameId: data.GetFrameId(),
	}
	return to
}

func (this *UserHead) ToDbProto() *proto_db.UserHead {
	if this == nil {
		return &proto_db.UserHead{}
	}
	to := &proto_db.UserHead{
		HeadId:  this.HeadId,
		FrameId: this.FrameId,
	}
	return to
}

func (this *UserHead) ToObjProto() *proto_object.UserHead {
	if this == nil {
		return &proto_object.UserHead{}
	}
	to := &proto_object.UserHead{
		HeadId:  this.HeadId,
		FrameId: this.FrameId,
	}
	return to
}

// UserStatistics funcs
func NewUserStatisticsFromDb(statData *proto_db.UserStatisticData) *UserStatistics {
	return &UserStatistics{
		TodayTotalOnlineTime:         statData.GetTodayTotalOnlineTime(),
		LastCalcTodayTotalOnlineTime: 0,
		CumulativeLoginDays:          statData.GetCumulativeLoginDays(),
		UserTotalOnlineTime:          statData.GetUserTotalOnlineTime(),
		Exp:                          statData.GetExp(),
		ShowAchievementMedals:        statData.GetShowAchievementMedals(),
		ShowRoles:                    statData.ShowRoles,
	}
}

func (this *UserStatistics) Clone() *UserStatistics {
	cop := *this
	return &cop
}

func (this *UserStatistics) ToUserStatisticsObj() *proto_object.UserStatisticData {
	return &proto_object.UserStatisticData{
		TodayTotalOnlineTime:   this.TodayTotalOnlineTime,
		CumulativeLoginDays:    this.CumulativeLoginDays,
		ShowAchievementMedalId: this.ShowAchievementMedals,
		Exp:                    this.Exp,
		ShowRoles:              this.ShowRoles,
	}
}

func (this *UserStatistics) ToUserStatisticsDb() *proto_db.UserStatisticData {
	return &proto_db.UserStatisticData{
		TodayTotalOnlineTime:  this.TodayTotalOnlineTime,
		CumulativeLoginDays:   this.CumulativeLoginDays,
		ShowAchievementMedals: this.ShowAchievementMedals,
		Exp:                   this.Exp,
		UserTotalOnlineTime:   this.UserTotalOnlineTime,
		ShowRoles:             this.ShowRoles,
	}
}
