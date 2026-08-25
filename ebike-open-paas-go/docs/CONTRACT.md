# 契约对照与落地说明

配套：`xiaoan-paas-api.md`（小安原文）、本服务源码。

## 包络

对外一律 Xiaoan 形：

```json
{"success": true, "data": ...}
{"success": false, "error": {"errormessage": "UNAUTHORIZED_ERROR", "promot": "..."}}
```

指令成功时 `success=true` 且 `data.code` 为设备结果码（0/108/109/132/136/1001…）。

内部下游仍是平台 `Result{success,code,msg,data}`；翻译只发生在本服务边界。

## 鉴权

1. Header `xc-access-token` 与 body/query `agentId` 必须同时出现
2. `open.agents`（Nacos 热更新）映射 `agentId → tenantId + token`
3. 设备归属：`device_ebike_{imei}.tenantId` 必须等于 agent 映射的 `tenantId`（anvelink 导入即可，不要求绑车辆编号）
4. 限流：Redis 滑动窗口 `openPaas_rateLimit_{agentId}`，默认 30000 / 300s

## 回调

| 存储 | `openPaas_callback_{agentId}_{event}` = SET(url) |
| 注册 | `POST /ebike/v1/callback` |
| 查询 | `GET /ebike/v1/callback?event=` |
| 注销 | `DELETE /ebike/v1/callback?event=&url=` |
| 入站 | `POST /internal/event/report`（consume Feign） |
| 白名单 | `GET /internal/callback/subscribedTenants` |

分发：有界队列 + worker 池 + 重试；GPS (0,0) 过滤；队列满则丢弃并计数。

## 接口判决摘要

| # | 路径 | 实现 |
|---|---|---|
| 1 | allDevices | anvelink-console `/device/getList`（consoleAuth 自动登录/续期） |
| 2 | deviceInfo | paas detail |
| 3 | api/device | paas deviceInfo (c34) |
| 4 | GPSPoints | worker getTrajectory type=0，秒→毫秒 |
| 5 | lock | paas lock（locked→acc） |
| 6 | acc | paas lock |
| 7 | defend | paas defend |
| 8 | backWheel | paas rearWheelLock（locked 与 sw 极性相反） |
| 9 | batteryCompartment | paas batteryCompartment |
| 10 | deviceVoice | paas voice + voiceIndexMap |
| 11-13 | callback | Redis CRUD |
| 14 | bluetooth | paas bluetooth |
| 15 | reboot | paas restart |
| 16 | batteryPowerSwitch | 明确错误 |
| 17 | mc | openapi /set_limit_speed |
| 18 | currentBmsInfo | transmission c=41 |
| 19 | address | detail + map/regeo |
| 20 | lbs2gps | 明确错误 |
| 21 | batteryInfo | detail 部分字段 |
| 22 | bmsInfo | detail SN |
| 23 | sms | 明确错误 |
