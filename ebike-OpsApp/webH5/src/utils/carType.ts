import type { DeviceItem } from '@/api'

/**
 * 车辆状态映射：把 `/business/paas/device/list` 的原始状态码归类成大屏的统计口径。
 * `key` 为空的 `all` 表示不过滤。
 */
export interface CarTypeRule {
  name: string
  key: keyof DeviceItem | ''
  value: number | ''
  field: string
}

export const CAR_TYPES: readonly CarTypeRule[] = [
  { name: '所有车辆', key: '', value: '', field: 'all' },

  // 基本状态
  { name: '可使用', key: 'ridingState', value: 1, field: 'normalAvailable' },
  { name: '骑行中', key: 'ridingState', value: 2, field: 'normalRiding' },
  { name: '临停中', key: 'ridingState', value: 3, field: 'normalTempParking' },
  { name: '预约中', key: 'ridingState', value: 4, field: 'normalSubscribing' },
  { name: '运维中', key: 'ridingState', value: 5, field: 'operating' },

  // 运维中
  { name: '已报修', key: 'operationState', value: 5, field: 'operatingRepair' },
  { name: '低电量', key: 'operationState', value: 4, field: 'operatingLowbattery' },
  { name: '调度中', key: 'operationState', value: 2, field: 'operatingScheduling' },
  { name: '换电中', key: 'operationState', value: 3, field: 'operatingChangeBattery' },
  { name: '拖回中', key: 'operationState', value: 6, field: 'operatingRecall' },
  { name: '已下架', key: 'operationState', value: 1, field: 'operatingOffshelves' },

  // 告警状态
  { name: '高电压', key: 'alarmState', value: 1, field: 'alarmHeightBattery' },
  { name: '异常离线', key: 'alarmState', value: 7, field: 'alarmOffline' },
  { name: '电瓶移除', key: 'alarmState', value: 6, field: 'alarmPowerCut' },
  { name: '出服务区', key: 'alarmState', value: 3, field: 'alarmOutGfence' },
  { name: '异常移动', key: 'alarmState', value: 2, field: 'alarmMove' },
  { name: '车辆报失', key: 'alarmState', value: 9, field: 'alarmLost' },
  { name: '有单无程', key: 'alarmState', value: 8, field: 'alarmOrderWithoutGps' },
  { name: '超长订单', key: 'alarmState', value: 10, field: 'alarmTooLongOrder' },
  { name: '短时订单', key: 'alarmState', value: 11, field: 'alarmTooShortOrder' },
  { name: '开锁异常', key: 'alarmState', value: 12, field: 'alarmUnlockAbnormal' },
  { name: '站点外', key: 'alarmState', value: 5, field: 'alarmOutParkingZone' },
  { name: '禁停区', key: 'alarmState', value: 99, field: 'alarmNoParkingZone' },
  { name: '禁行区', key: 'alarmState', value: 5, field: 'alarmNoRidingZone' },
  { name: '头盔丢失', key: 'alarmState', value: 13, field: 'alarmHelmetLost' },
  { name: '头盔故障', key: 'alarmState', value: 14, field: 'alarmHelmetFault' },
]

export type DeviceBuckets = Record<string, DeviceItem[]>

/** 服务端对同一状态字段既可能返回标量也可能返回数组，两种都要能命中。 */
function matches(device: DeviceItem, rule: CarTypeRule): boolean {
  if (rule.field === 'all') return true
  const actual = device[rule.key as keyof DeviceItem]
  if (Array.isArray(actual)) return actual.includes(rule.value as number)
  return actual === rule.value
}

export function bucketDevices(devices: DeviceItem[]): DeviceBuckets {
  const buckets: DeviceBuckets = {}
  for (const rule of CAR_TYPES) {
    buckets[rule.field] = devices.filter((device) => matches(device, rule))
  }
  return buckets
}

/** 已下架（operationState 含 1）的车不计入任何运营统计。 */
export function excludeOffShelves(devices: DeviceItem[]): DeviceItem[] {
  return devices.filter((device) => !device.operationState?.includes(1))
}
