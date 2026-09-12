import { describe, expect, it } from 'vitest'
import { PermissionCode } from '@/stores/permission'
import {
  classify,
  isCarId,
  isImei,
  isPhone,
  requiredCode,
  scopeHint,
} from './searchKind'

describe('isPhone', () => {
  it('接受大陆 11 位号段', () => {
    expect(isPhone('13800138000')).toBe(true)
    expect(isPhone('19912345678')).toBe(true)
  })

  it('接受香港 8 位', () => {
    expect(isPhone('51234567')).toBe(true)
    expect(isPhone('91234567')).toBe(true)
  })

  it('拒绝明显不是手机号的串', () => {
    expect(isPhone('12345')).toBe(false)
    expect(isPhone('abcdefghijk')).toBe(false)
    expect(isPhone('12800138000')).toBe(false)
  })
})

describe('isImei', () => {
  it('只认 15 位纯数字', () => {
    expect(isImei('860123456789012')).toBe(true)
    expect(isImei('86012345678901')).toBe(false)
    expect(isImei('8601234567890123')).toBe(false)
    expect(isImei('86012345678901a')).toBe(false)
  })
})

describe('isCarId', () => {
  it('认 9 位纯数字', () => {
    expect(isCarId('123456789')).toBe(true)
  })

  it('认 08 开头的 7/8 位（修了遗留 matches 死代码）', () => {
    expect(isCarId('0812345')).toBe(true)
    expect(isCarId('08123456')).toBe(true)
    expect(isCarId('081234')).toBe(false)
    // 9 位纯数字本身就合法，跟是否 08 开头无关
    expect(isCarId('081234567')).toBe(true)
    expect(isCarId('0812345678')).toBe(false)
  })

  it('拒绝非数字', () => {
    expect(isCarId('08ABCDE')).toBe(false)
  })
})

describe('classify', () => {
  it('优先级：手机号 → IMEI → 车号 → 姓名', () => {
    expect(classify('13800138000')).toBe('phone')
    expect(classify('860123456789012')).toBe('imei')
    expect(classify('123456789')).toBe('carId')
    expect(classify('0812345')).toBe('carId')
    expect(classify('张三')).toBe('name')
  })

  it('忽略空白', () => {
    expect(classify(' 13800138000 ')).toBe('phone')
  })

  it('15 位数字不会被当成手机号或车号', () => {
    expect(classify('860123456789012')).toBe('imei')
  })
})

describe('requiredCode', () => {
  it('人员类走 121202，车辆类走 121201', () => {
    expect(requiredCode('phone')).toBe(PermissionCode.OrderQueryPersonal)
    expect(requiredCode('name')).toBe(PermissionCode.OrderQueryPersonal)
    expect(requiredCode('carId')).toBe(PermissionCode.OrderQueryVehicle)
    expect(requiredCode('imei')).toBe(PermissionCode.OrderQueryVehicle)
  })
})

describe('scopeHint', () => {
  it('按权限组合给 placeholder', () => {
    expect(scopeHint(true, true)).toBe('请输入手机号、姓名、车辆号、设备号')
    expect(scopeHint(true, false)).toBe('请输入手机号、姓名')
    expect(scopeHint(false, true)).toBe('请输入车辆号、设备号')
    expect(scopeHint(false, false)).toBe('')
  })
})
