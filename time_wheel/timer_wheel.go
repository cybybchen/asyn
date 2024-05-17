package time_wheel

import (
	"fmt"
	"px/framebase/guid"
	"px/utils"
	"px/utils/cbctx"
	"px/utils/chanx"
	"reflect"
	"runtime"
	"runtime/debug"
	"time"

	"gitlab.sunborngame.com/base/log"
)

const (
	TimeWheelStep = 50 * time.Millisecond
	TimeWheelSlot = 1000
	TaskChanCap   = 10240
)

type ChanTask struct {
	Task    *TaskS // if not nil, add timer. else del timer by TimerID
	TimerID int64
}

// CallbackFuncS 延时任务回调函数
type CallbackFuncS func(int64, cbctx.Ctx)

// TimeWheelS 时间轮
type TimeWheelS struct {
	timeStep      time.Duration //时间步长
	ticker        *time.Ticker
	logTick       *time.Ticker
	tasks         *utils.SortedSet[int64, *TaskS] //时间轮槽
	taskChannel   *chanx.UnboundedChan[*ChanTask] // 任务channel
	stopChannel   chan bool                       // 停止定时器channel
	NotifyChannel *chanx.UnboundedChan[*TaskS]

	lastTickTime int64
}

type TaskS struct {
	TimerID int64
	Data    cbctx.Ctx // 回调函数参数

	delay    time.Duration // 延迟时间
	repeated bool
	interval int64 //每次间隔时间
	exeTime  int64

	defaultCb CallbackFuncS
}

func (t *TaskS) Less(other utils.SortedSetData[int64]) bool {
	data := other.(*TaskS)
	return t.exeTime < data.exeTime
}

func (t *TaskS) Key() int64 {
	return t.TimerID
}

func NewTaskS(delay time.Duration, repeated bool, data cbctx.Ctx, interval int64, defaultCb CallbackFuncS) *TaskS {
	return &TaskS{
		TimerID:   int64(guid.GetGuidMgr().GenerateNewGuid()),
		delay:     delay,
		Data:      data,
		repeated:  repeated,
		interval:  interval,
		defaultCb: defaultCb,
	}
}

func (t *TaskS) GetCallBack() CallbackFuncS {
	return t.defaultCb
}

func (t *TaskS) TriggerDefaultCb() {
	var start = utils.NowUnixMilli()
	if !t.repeated && start-t.exeTime > utils.Millisecond*2 {
		log.Error("TriggerDefaultCb timer trigger diff > than 2 sec, now is %d, exeTime is %d, diff is %d, timeId=%d, callback=%v", start, t.exeTime, start-t.exeTime, t.TimerID, runtime.FuncForPC(reflect.ValueOf(t.defaultCb).Pointer()).Name())
	}
	t.defaultCb(utils.NowUnixMilli(), t.Data)
	var end = utils.NowUnixMilli()
	var cost = end - start
	if cost > 50 {
		log.Info("timetick exc cost time > 1 ms, cost=%v, callback=%v", cost, runtime.FuncForPC(reflect.ValueOf(t.defaultCb).Pointer()).Name())
	}
}

func (t *TaskS) updateNextDelay() {
	if t.interval > 0 {
		t.delay = time.Duration(t.interval) * time.Millisecond
		t.exeTime = utils.NowUnixMilli() + t.interval
		t.interval = 0
	}
}

func NewTimeWheelSDefault() *TimeWheelS {
	return NewTimeWheelS(TimeWheelStep)
}

// NewTimeWheelS 创建时间轮
func NewTimeWheelS(timeStep time.Duration) *TimeWheelS {
	if timeStep <= 0 {
		return nil
	}
	tw := &TimeWheelS{
		timeStep:      timeStep,
		taskChannel:   chanx.NewUnboundedChan[*ChanTask](TaskChanCap),
		stopChannel:   make(chan bool, 1024),
		NotifyChannel: chanx.NewUnboundedChan[*TaskS](TaskChanCap),
	}
	tw.initTasks()

	return tw
}

func (tw *TimeWheelS) initTasks() {
	tw.tasks = utils.NewSortedSet[int64, *TaskS]()
}

func (tw *TimeWheelS) Start() {
	tw.ticker = time.NewTicker(tw.timeStep)
	tw.logTick = time.NewTicker(30 * time.Second)
	tw.lastTickTime = utils.NowUnixMilli()
	go tw.start()
}

func (tw *TimeWheelS) Stop() {
	tw.stopChannel <- true
}

func (m *TimeWheelS) AddOnceTimer(milli int64, cb CallbackFuncS, data ...interface{}) int64 {
	return m.addTimer(time.Duration(milli)*time.Millisecond, false, data, 0, cb)
}

func (m *TimeWheelS) AddRepeatTimer(milli int64, cb CallbackFuncS, data ...interface{}) int64 {
	return m.addTimer(time.Duration(milli)*time.Millisecond, true, data, 0, cb)
}

func (m *TimeWheelS) AddRepeatTimerByInterval(milli int64, interval int64, cb CallbackFuncS, data ...interface{}) int64 {
	return m.addTimer(time.Duration(milli)*time.Millisecond, true, data, interval, cb)
}

func (m *TimeWheelS) AddOnceTimerWithExpireTimestamp(expireTs int64, cb CallbackFuncS, data ...interface{}) int64 {
	remainMilli := expireTs - utils.NowUnixMilli()
	if remainMilli <= 0 {
		// 可能执行时当前时间已经超过expireTs而导致addTimer直接返回
		remainMilli = 1
	}
	return m.addTimer(time.Duration(remainMilli)*time.Millisecond, false, data, 0, cb)
}

// addTimer 添加定时器
// 主线程执行
func (tw *TimeWheelS) addTimer(delay time.Duration, repeated bool,
	datas []interface{}, interval int64, defaultCb CallbackFuncS) int64 {
	if delay < 0 {
		log.Error("addTimer err delay < 0 %s", string(debug.Stack()))
		return 0
	}
	iData := make(cbctx.Ctx, 0, len(datas))
	for _, data := range datas {
		switch v := data.(type) {
		case int:
			iData = append(iData, cbctx.CtxInt(v))
		case uint32:
			iData = append(iData, cbctx.CtxInt(v))
		case int64:
			iData = append(iData, cbctx.CtxInt(v))
		case int32:
			iData = append(iData, cbctx.CtxInt(v))
		case uint64:
			iData = append(iData, cbctx.CtxInt(v))
		case time.Duration:
			iData = append(iData, cbctx.CtxInt(v))
		case string:
			iData = append(iData, cbctx.CtxString(v))
		case bool:
			iData = append(iData, cbctx.CtxBool(v))
		default:
			iData = append(iData, cbctx.NewCtxObj(data))
		}
	}
	t := NewTaskS(delay, repeated, iData, interval, defaultCb)
	tw.taskChannel.Put(&ChanTask{
		Task: t,
	})
	log.Debugc(func() string {
		return fmt.Sprintf("addTimer id:[%d] name: [%v]", t.TimerID, runtime.FuncForPC(reflect.ValueOf(t.defaultCb).Pointer()).Name())
	})
	return t.TimerID
}

// RemoveTimer 删除定时器 key为添加定时器时传递的定时器唯一标识
func (tw *TimeWheelS) RemoveTimer(timerID int64) {
	tw.taskChannel.Put(&ChanTask{TimerID: timerID})
}

func (tw *TimeWheelS) start() {
	for {
		select {
		case key := <-tw.taskChannel.C():
			if key.Task != nil {
				tw.addTaskS(key.Task)
			} else if key.TimerID > 0 {
				tw.removeTaskS(key.TimerID)
			}
			continue
		default:
		}
		select {
		case now := <-tw.ticker.C:
			tw.tickHandler(utils.NanoToMilliSeconds(now.UnixNano()))
		case now := <-tw.logTick.C:
			log.Info("TimeWheelS current task count=%d, now=%d", tw.tasks.Len(), utils.NanoToMilliSeconds(now.UnixNano()))
		case <-tw.stopChannel:
			tw.ticker.Stop()
			return
		}
	}
}

func (tw *TimeWheelS) tickHandler(now int64) {
	tw.scanAndNotify(tw.tasks, now)
}

// 没有可继续执行的任务了，返回false
func (tw *TimeWheelS) scanAndNotify(l *utils.SortedSet[int64, *TaskS], now int64) {
	for {
		taskS, ok := l.Front()
		if !ok {
			break
		}
		if taskS.exeTime > now {
			break
		}
		tw.NotifyChannel.Put(taskS)
		tw.tasks.Remove(taskS)
		if taskS.repeated {
			taskS.updateNextDelay()
			tw.addTaskS(taskS)
		}
	}
}

func (tw *TimeWheelS) addTaskS(taskInf any) {
	taskS, ok := taskInf.(*TaskS)
	if !ok {
		log.Error("taskinf type err, taskInfo=%v", taskInf)
		return
	}

	taskS.exeTime = taskS.delay.Milliseconds() + utils.NowUnixMilli()
	tw.tasks.Push(taskS)
}

func (tw *TimeWheelS) removeTaskS(timerID int64) {
	log.Debug("removeTaskS remove timer task %d", timerID)
	tw.tasks.RemoveByKey(timerID)
}

// TriggerTimerCb 主线程执行
func (tw *TimeWheelS) TriggerTimerCb(task *TaskS) {
	task.TriggerDefaultCb()
}
