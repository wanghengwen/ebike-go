package constants

// Cmd mirrors Java CmdConstant (subset used by persistence and push).
const (
	CmdWild       = 0
	CmdPing       = 2
	CmdGPS1       = 3
	CmdAlarm      = 5
	CmdLogin      = 35
	CmdBMSInfo    = 66
	CmdFault      = 70
	CmdGPSV6      = 68
	CmdGPSPack    = 29
	CmdBikeGPS    = 81
	CmdBatteryBMS = 76
	CmdLogout     = 1001

	// CmdShuaka 刷卡中控上报（玉环项目定制，对应 ebike-device-worker 分支 feature/Bin201Decode-20240531）。
	CmdShuaka = 201
)

// Redis key prefixes — mirrors DeviceConstant.
const (
	RedisKeyDeviceEbike    = "device_ebike_"
	RedisKeyDeviceBike     = "device_bike_"
	RedisKeyDeviceBattery  = "device_battery_"
	RedisKeyDeviceCabinet  = "device_cabinet_"
	RedisKeyHashTenant     = "tenant:"
	RedisQueueChangeTenant = "deviceTenantIdQueue"
)
