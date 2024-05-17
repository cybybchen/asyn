package user_base

import "fmt"

type (
	UserBase struct {
		Account            string
		Id                 uint64
		Name               string
		Level              int32
		ServerId           int32
		Channel            string
		CreateTime         int64  //毫秒
		LastLoginTime      int64  //毫秒
		LastLogoutTime     int64  //毫秒
		LastDailyResetTime int64  //上次每日重置时间
		HashAccount        uint64 //account的hasn值，用于mod
		// 以下存入proto_db.UserBaseData
		*UserStatistics
		UserHead  *UserHead
		Signature string //签名

		simpleStr string
	}
	UserStatistics struct {
		TodayTotalOnlineTime         int64   // 今日累积在线时长
		UserTotalOnlineTime          int64   // 历史在线总时长
		LastCalcTodayTotalOnlineTime int64   // 上次计算今日累积在线时长的时间戳
		CumulativeLoginDays          int32   // 累计登录天数
		ShowAchievementMedals        []int32 // 展示的成就列表
		Exp                          int64   //经验
		ShowRoles                    []int32 // 展示的角色列表
	}
	UserHead struct {
		HeadId  int32
		FrameId int32
	}
)

func (this *UserBase) GetSimpleStr() string {
	if this.simpleStr == "" {
		this.RefreshSimpleStr()
	}
	return this.simpleStr
}

func (this *UserBase) RefreshSimpleStr() {
	this.simpleStr = fmt.Sprintf("(id:%d,name:%s,server:%d)", this.Id, this.Name, this.ServerId)
}
