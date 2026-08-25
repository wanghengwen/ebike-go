package event

// Xiaoan notify codes (callback event=3). Only the ones we can actually source
// from saas_0 are named here; see docs for the full 0..56 table.
const (
	NotifyAutoDefend     = 0  // 车辆已自动设防
	NotifyLogin          = 3  // 设备登录网关服务器
	NotifyLostConnection = 4  // 设备与网关服务器失去连接
	NotifyMoveAlarm      = 5  // 锁车/设防状态下移动报警
	NotifyBatteryRemoved = 6  // 电瓶移除报警
	NotifyAccOn          = 7  // 电门开启
	NotifyAccOff         = 8  // 电门关闭
	NotifySeatUnlocked   = 11 // 鞍座(电池仓)已开锁
	NotifySeatLocked     = 12 // 鞍座(电池仓)已上锁
	NotifyFenceExit      = 17 // 出地理围栏
	NotifyFenceEnter     = 18 // 入地理围栏
	NotifyBatteryRestore = 23 // 电瓶恢复连接
	NotifyOverloadOn     = 55 // 超载事件触发
	NotifyOverloadOff    = 56 // 超载事件解除
)

// alarmToNotify maps our internal AlarmTypeEnum onto Xiaoan notify codes.
//
// Only codes with an unambiguous counterpart are listed. An unmapped alarm is
// dropped rather than forwarded as-is: the two numbering schemes overlap
// (internal 13 is "auto lock", Xiaoan 13 is "auto defend for shared devices"),
// so passing an unknown code through would deliver a plausible-looking but wrong
// event to third parties.
var alarmToNotify = map[int]int{
	3:  NotifyMoveAlarm,      // 非法移动
	6:  NotifyBatteryRemoved, // 电源断开
	7:  NotifyAccOn,          // 电门开
	8:  NotifyAccOff,         // 电门关
	9:  NotifySeatUnlocked,   // 电池仓开
	10: NotifySeatLocked,     // 电池仓关
	11: NotifyBatteryRestore, // 电源接通
	13: NotifyAutoDefend,     // 自动落锁
	55: NotifyOverloadOn,
	56: NotifyOverloadOff,
}

// NotifyForAlarm translates an internal alarm type. ok is false when the alarm
// has no Xiaoan equivalent.
func NotifyForAlarm(alarmType int) (notify int, ok bool) {
	n, ok := alarmToNotify[alarmType]
	return n, ok
}
