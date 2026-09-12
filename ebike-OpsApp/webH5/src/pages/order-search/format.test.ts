import { describe, expect, it } from 'vitest'
import { OrderPayState } from '@/api'
import {
  distance,
  duration,
  isSettled,
  isoDateTime,
  orDash,
  payStateLabel,
  ridingStateLabel,
  yuan,
} from './format'

describe('yuan', () => {
  it('分转元，保留两位', () => {
    expect(yuan(0)).toBe('0.00')
    expect(yuan(1)).toBe('0.01')
    expect(yuan(100)).toBe('1.00')
    expect(yuan(12345)).toBe('123.45')
  })

  it('空值出 --', () => {
    expect(yuan(null)).toBe('--')
    expect(yuan(undefined)).toBe('--')
  })
})

describe('duration', () => {
  it('位数超过 10 当毫秒再换算成秒', () => {
    // 11 位：36610000000 ms → 36610000 秒
    expect(duration(36_610_000_000)).toBe('423天17小时26分40秒')
    expect(duration('36610000000')).toBe('423天17小时26分40秒')
  })

  it('不超过 10 位当秒', () => {
    // 3661 秒 = 1 小时 1 分 1 秒（短时长走这条）
    expect(duration(3661)).toBe('1小时1分1秒')
    expect(duration(61)).toBe('1分1秒')
    expect(duration(86400)).toBe('1天')
  })

  it('不满一分钟也要显示秒，避免空白', () => {
    expect(duration(0)).toBe('0秒')
    expect(duration(5)).toBe('5秒')
  })

  it('空值与非法值出 --', () => {
    expect(duration(null)).toBe('--')
    expect(duration('')).toBe('--')
    expect(duration('abc')).toBe('--')
    expect(duration(-1)).toBe('--')
  })
})

describe('distance', () => {
  it('不足 1 公里按米，否则按公里', () => {
    expect(distance(0)).toBe('0米')
    expect(distance(999)).toBe('999米')
    expect(distance(1000)).toBe('1.00公里')
    expect(distance(1530)).toBe('1.53公里')
  })

  it('空值出 --', () => {
    expect(distance(null)).toBe('--')
    expect(distance(undefined)).toBe('--')
  })
})

describe('payStateLabel / isSettled', () => {
  it('映射 OrderIzPayEnum', () => {
    expect(payStateLabel(OrderPayState.Riding)).toBe('骑行中')
    expect(payStateLabel(OrderPayState.Frozen)).toBe('支付中')
    expect(payStateLabel(OrderPayState.ToPay)).toBe('待支付')
    expect(payStateLabel(OrderPayState.Paid)).toBe('已支付')
    expect(payStateLabel(99)).toBe('--')
    expect(payStateLabel(null)).toBe('--')
  })

  it('只有已支付算结清', () => {
    expect(isSettled(OrderPayState.Paid)).toBe(true)
    expect(isSettled(OrderPayState.ToPay)).toBe(false)
    expect(isSettled(null)).toBe(false)
  })
})

describe('ridingStateLabel', () => {
  it('映射用户 ridingState', () => {
    expect(ridingStateLabel(6)).toBe('骑行中')
    expect(ridingStateLabel(7)).toBe('待支付')
    expect(ridingStateLabel(0)).toBe('--')
    expect(ridingStateLabel(null)).toBe('--')
  })
})

describe('isoDateTime / orDash', () => {
  it('ISO 的 T 换成空格', () => {
    expect(isoDateTime('2024-01-02T03:04:05')).toBe('2024-01-02 03:04:05')
    expect(isoDateTime('')).toBe('--')
    expect(isoDateTime(null)).toBe('--')
  })

  it('空串和 - 都当无值', () => {
    expect(orDash('hello')).toBe('hello')
    expect(orDash('')).toBe('--')
    expect(orDash('-')).toBe('--')
    expect(orDash('  -  ')).toBe('--')
    expect(orDash(null)).toBe('--')
  })
})
