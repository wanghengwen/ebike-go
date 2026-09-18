# Rider BLE 协议（P3）

对照 UniApp `ebike-UniApp/src/features/ble/`。本阶段**故意保留现网语义**，不修正 CRC / `resolveResult` 的怪比较。

## 帧

```
cmd(1) | datalen(1) | token(4)[+payload] | crc(1)
```

构造 CRC：`(cmd + Σdata + datalen) & 0xff`（`data` = token + payload）。

默认 token：`0A 0A 05 05`（十进制 `168428805`）。规划稿曾写成 `168496389`，那是 `0x0A0B0D05`，不是默认字节。

静音开/锁：`MUTE_PAYLOAD = [0, 0]`。语音在骑行上报成功后另发（P4）。

Token 超 32 位：TS `tokenToDataArray` 会写出超过 4 字节；Kotlin **显式取低 4 字节**，不抛异常。

## 黄金帧（默认 token）

| 动作 | hex |
| --- | --- |
| UNLOCK 静音 | `2c060a0a0505000050` |
| LOCK 静音 | `2b060a0a050500004f` |
| LOCK 临锁 voice7 vol100 | `2b060a0a05050764ba` |
| PLAY_VOICE START vol100 | `28060a0a05050364b3` |
| GET_DEVICE_INFO | `41040a0a050563` |
| GET_GPS | `32040a0a050554` |
| GET_LAST_BEACON | `42040a0a050564` |
| GET_RFID | `54040a0a050576` |

规划稿里临锁帧曾写成 `…645a`；按 TS 公式手算为 `…64ba`（`0x2b + Σdata + 6 = 0xba`）。

## `resolveAck`（`resolveResult`）

`retCode = parseInt(hex[4 .. 4+len*2], 16)`，比较 **`crc === cmd + length + retCode`（无 `& 0xff`）**。

`len > 1` 时 `retCode` 是多字节整数，比较几乎必失败，函数返回 `code: 0`。上层 `code === 0` 算成功 —— **与现网一致，不要改**。

`crcValid` 循环为 `i < len-1; i += 2`（含 CRC 字节），再比 `(sum & 0xff) === ((crc+crc) & 0xff)`。

## UUID / 扫描

| 项 | 值 |
| --- | --- |
| 扫描过滤 | `5841`（16-bit → `00005841-0000-1000-8000-00805F9B34FB`） |
| Service | `0783B03E-8535-B5A0-7140-A304D2495CB7` |
| Write RX | `…5CBA` |
| Notify TX | `…5CB8` |

广播匹配：`imei.indexOf(advertHex) === 3 && advertHex.length === 12`。

超时：搜索 / 连接 / 发送均为 60s（UniApp）。

## 指令码

`PLAY_VOICE=0x28` `LOCK=0x2b` `UNLOCK=0x2c` `GET_GPS=0x32` `GET_DEVICE_INFO=0x41` `GET_LAST_BEACON=0x42` `GET_RFID=0x54`

语音：`START=3` `LOCK=1` `TEMP_LOCK=7`，默认音量 100。

## Token API

`POST /client/paas/device/getBlueToothToken`，body 带 `imei`。失败回落默认 token。

## 真机未验

- 0x41 / 0x42 / 0x54 notify 字段偏移需要真车抓包补强（单测里 CRC 失败路径 + 合成 GPS 帧已覆盖解析骨架）。
- Android 12+ `BLUETOOTH_SCAN` / `CONNECT`、iOS `NSBluetoothAlwaysUsageDescription` 已在 Manifest / `iosApp/project.yml` 登记，未在真机弹窗走完。
- 厂商 BLE 栈（小米 / OPPO / vivo / 鸿蒙）：CCCD 写入延迟、首包 notify 丢失、`5841` 是否进广播，均属已知风险，本阶段不绕。
