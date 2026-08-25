# 小安科技电单车 PaaS 服务接口

> 来源：[https://paas.doc.xiaoantech.com/](https://paas.doc.xiaoantech.com/)  
> 文档形态：Postman Documenter（collection `RW8FF6Se`），按官方集合转写  
> 当前版本：**v1.1.13**  
> 基地址：`https://api.xiaoantech.com`  
> 说明：示例 JSON 保留原文写法（如个别 `succcess` 拼写）；测试 `agentId` / `xc-access-token` 来自官方文档公开测试用例，不可用于生产。

适用于搭载小安盒子的共享电单车、智能电动车控制与数据通信
 
## 当前版本 v1.1.13
 
## 术语解释
 

**agent：** 每一个使用PaaS平台的个人和团体
**agentId：** agent的标识
**xc-access-token：** 访问接口的唯一凭证 
 
## 测试用例
 

**测试agentId：**87
**测试xc-access-token:**525511E04347F12AE6C5D74D705DD84A 
 
## 错误码
 
### 业务错误
 

接口调用时出现suc=false的业务错误
 
### 业务错误对照表
                                  
| ERROR_TYPE | 含义 |
| --- | --- |
| INPUT_PARAMS_MISS | 入参缺失 |
| SEND_CMD_ERROR | 发送命令失败 |
| DEVICE_INFO_IS_NULL | 设备信息为空 |
| UNAUTHORIZED_ERROR | TOKEN认证失败 |
| IMEI_ILLEGAL | imei错误 |
| BIZ_ERROR | 其他业务错误 |
 
### 设备通信错误
 

与设备进行通信的接口，在调用成功suc=true后会返回这次通信结果的code。
100~110：网关服务器异常情况返回
110~  : 设备异常情况返回
1000~ : 业务异常情况返回 
 
### code含义对照表
                                                                                      
| code | 含义 |
| --- | --- |
| 0 | 操作成功 |
| 100 | 服务器内部错误 |
| 101 | 请求无IMEI |
| 102 | 无请求内容 |
| 103 | 无请求内容 |
| 104 | 请求URL错误 |
| 105 | 请求范围过大 |
| 106 | 接入服务器无响应 |
| 107 | 接入服务器不在线 |
| 108 | 设备无响应 |
| 109 | 设备离线 |
| 110 | 设备内存错误 |
| 111 | 设备不支持该命令 |
| 113 | 电池类型未设置 |
| 114 | 参数不合法 |
| 115 | 参数缺失 |
| 132 | 不执行锁车命令。(正在骑行) |
| 136 | 未扫描到道钉 |
| 1001 | 设备不属于当前agent |
 
## 注意事项
 

接口全部使用HTTPS协议。 每一个请求的header都必须有xc-access-token字段。
各个接口所列字段的透出与否(不涉及字段变更)和接入设备型号有关，调用方可根据接口实际返回字段确认。
 

实时接口访问频率阈值：30k/5分钟千台车(在线) 
 
## 版本修订记录
                                                                                                
| 版本号 | log | 更新时间 |
| --- | --- | --- |
| v1.1.13 | 开锁车相关接口增加道钉停车功能 | 2020-08-06 |
| v1.1.12 | 回调数据类型增加BMS数据上报 | 2020-06-10 |
| v1.1.11 | 增加车辆短信协议控制接口 | 2019-12-12 |
| v1.1.10 | 增加获取电池BMS信息的实时和缓存接口(需设备支持) | 2019-11-16 |
| v1.1.9 | 实时接口增加字段支持播放特定位铃声，并能控制音量 | 2019-10-18 |
| v1.1.8 | 增加设防接口 |  |
| v1.1.7 | gps callback过滤无定位(0,0)坐标的发送 |  |
| v1.1.6 | 增加控制器限速接口，增加基站数据解析接口 |  |
| v1.1.5 | 增加注册事件类型5，透出设备UART/485 buffer数据 |  |
| v1.1.4 | 增加获取设备最新定位文字位置的接口 |  |
| v1.1.3 | 缓存接口透出平滑后电压值 |  |
| v1.1.2 | 增加缓存更新时机和事件上报时机描述 |  |
| v1.1.1 | 增加GPS 状态位sw 描述 |  |
| v1.1.0 | 增加事件回调相关接口 |  |
| v1.0.2 | 增加外设锁控制接口描述 |  |
| v1.0.1 | 测试参数变更 |  |
| v1.0.0 | 创建 |  |
 
## 接口描述


---

## 接口目录

| # | 接口 | 方法 | 路径 |
| --- | --- | --- | --- |
| 1 | 获取所有设备列表 | `GET` | `/ebike/v1/allDevices` |
| 2 | 获取设备信息 | `GET` | `/ebike/v1/deviceInfo` |
| 3 | 获取设备信息[实时] | `POST` | `/ebike/api/device` |
| 4 | 获取设备轨迹 | `GET` | `/ebike/v1/GPSPoints` |
| 5 | 开车/锁车[实时] | `POST` | `/ebike/v1/lock` |
| 6 | 电门控制[实时] | `POST` | `/ebike/v1/acc` |
| 7 | 设防控制[实时] | `POST` | `/ebike/v1/defend` |
| 8 | 后轮锁控制[实时] | `POST` | `/ebike/v1/backWheel` |
| 9 | 电池仓锁(鞍座锁)控制[实时] | `POST` | `/ebike/v1/batteryCompartment` |
| 10 | 播放语音[实时] | `POST` | `/ebike/v1/deviceVoice` |
| 11 | 注册事件回调链接 | `POST` | `/ebike/v1/callback` |
| 12 | 获取事件回调链接 | `GET` | `/ebike/v1/callback` |
| 13 | 注销事件回调链接 | `DELETE` | `/ebike/v1/callback` |
| 14 | 配置蓝牙参数[实时] | `POST` | `/ebike/v1/bluetooth` |
| 15 | 重启设备[实时] | `POST` | `/ebike/v1/reboot` |
| 16 | 电池放电开关控制[实时] | `POST` | `/ebike/v1/batteryPowerSwitch` |
| 17 | 设置控制器限速百分比[实时] | `POST` | `/ebike/v1/mc` |
| 18 | 获取车辆电池BMS信息[实时] | `GET` | `/ebike/v1/currentBmsInfo` |
| 19 | 获取设备最新定位地址信息 | `GET` | `/ebike/v1/address` |
| 20 | 基站定位数据转GPS定位 | `POST` | `/ebike/v1/lbs2gps` |
| 21 | 获取服务端缓存的电池信息 | `GET` | `/ebike/v1/batteryInfo` |
| 22 | 获取车辆电池BMS信息 | `GET` | `/ebike/v1/bmsInfo` |
| 23 | 短信接口[实时] | `POST` | `/ebike/v1/sms` |

---


## 1. 获取所有设备列表

- **方法**：`GET`
- **URL**：`https://api.xiaoantech.com/ebike/v1/allDevices`

### Query 参数

| 参数 | 示例 | 说明 |
| --- | --- | --- |
| `agentId` | `87` | agent 标识 |
| `pageSize` | `10` | 分页大小 |
| `pageNumber` | `0` | 页码，从 0 起 |

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |

### 说明

获取所有设备列表 
 
#### 返回值说明
 

imei:设备的15位唯一标识
 
#### 正确返回体
 

```
{
    succcess:true,
    data:{
        total:146,
        devices:[{
            imei:868183030457282
        }]
    }
}

```

#### 错误返回体
 

```
{
    succcess:false,
    error:{
        errormessage:"",
        promot:"request params missing!"
    }
}

```

---


## 2. 获取设备信息

- **方法**：`GET`
- **URL**：`https://api.xiaoantech.com/ebike/v1/deviceInfo`

### Query 参数

| 参数 | 示例 | 说明 |
| --- | --- | --- |
| `agentId` | `87` | agent 标识 |
| `imei` | `865532040744076` | 15 位设备号 |

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |

### 说明

服务端缓存最近一次上报的设备信息，用于查询加速。
 
#### 缓存数据刷新时机
 

设备有三种机制进行数据同步，即心跳包上报，GPS包上报，接受查询或设置命令时。
 
#### 同步机制
 

同步机制 | 同步周期 | 相关数据
--- | ---
心跳包 | 150s | 电压、信号强度
GPS包 | 5s/运动状态下，10min/静止状态下 | 经纬度，外设锁状态，设备状态
命令 | 立刻 | 命令涉及的数据
 

对于设备在线状态，由于设备与网关服务器的保活机制，设备登录网关服务器时立即判定为设备上线，而在设备与网关服务器失去连接5分钟后判定为离线。
 
#### 正确返回体
 

```
{
    succcess:true,
    data:{
        imei: "868183030457282",
        isOnline: 1,    
        gsmSignal: 23,   
        defend: 1,
        acc: 0,
        lat: 30.51040856769523,
        lng: 114.43096840580687,
        batteryLock: 1,
        backWheelLock: 1,
        objectType: 41,
        voltage: 49240,
        version: 131088,
        isMoving: 0,
        imsi: "880041107000000"
    }
}

```

各个字段的含义：
imei：设备号
isOnline: 车辆在线状态，1：在线，0：离线
gsmSignal：信号强度
defend：车辆设防状态，1：已设防，0：已撤防
acc：电门状态，1：电门开，0：电门关
lat：纬度，gcj02格式
lng：经度，gcj02格式
batteryLock：电池锁，1：上锁状态，0：解锁状态
backWheelLock：后轮锁，1：上锁状态，0：解锁状态
objectType：设备型号
voltage：平滑后的车辆电池电压，单位：mV
version：设备固件版本号
isMoving：是否在移动，1：移动，0：静止
imsi: 卡imsi
 

lastGPSTimestamp: 最后一次定位时间 时间戳
 

lastLoginTimestamp：最近一次登录时间 时间戳
 
#### 错误返回体
 

```
{
    succcess:true,
    error:{
        errormessage:"",
        promot:""
    }
}

```

---


## 3. 获取设备信息[实时]

- **方法**：`POST`
- **URL**：`https://api.xiaoantech.com/ebike/api/device`

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |

### 请求体

```json
{
  "agentId": 252,
  "agentToken": "C936941DA37396DF9C341D7FE446707D",
  "imei": "865209034340316",
  "cmd": {
    "c": 34
  }
}
```

### 说明

与设备通信获取设备当前信息，接口响应时间较长，用于准确获取设备实时数据和状态的场景，如判断开车锁车状态，电池仓锁状态等。
 
#### 参数说明
 

imei:设备的15位唯一标识
 
#### 正确返回体
 

```
{
    success: true,
    data: {
        code: 0,
        result: {
            autolock: {
                sw: 1,
                period: 5
            },
            battery: {
                percent: 0,
                type: 48,
                voltage: 240
            },
            voltageMv: 45663,
            GSMSignal: 23,
            gsm: 31,
            acc: 0,
            defend: 1,
            isBeep: 0,
            audioType: 7,
            pkeType: 1,
            gps: {
                timestamp: 1527930367,
                lat: 30.512766,
                lng: 114.425682,
                speed: 1,
                course: 87
            },
            wheelLock: 1,
          seatLock: 0,
          mode: 0,
          isWheelSpan: 0,
          powerState: 1,
        }
    }
}

```

各个字段的含义(字段是否存在与设备类型相关)：
code：查询信息的结果，详情参考错误码
autolock：自动设防状态，其中sw：1自动设防开启，sw：0自动设防关闭，period：自动设防时间
battery：电池信息，其中percent：剩余电量百分比，type：电池类型，voltage：电压值，单位：mv
voltageMv: 电压值，单位：mV
GSMSignal：信号强度
gsm: 信号强度
defend：设备设防状态，1：已设防，0：已撤防
acc：电门状态，1：电门开，0：电门关
isBeep：勿扰模式状态，1：开启，0：关闭
wheelLock： 后轮锁状态，1：已上锁，0：未上锁
seatLock: 后座锁状态，1：已上锁，0：未上锁
isWheelSpan: 后轮转动状态，1：转动中，0：静止
mode:设备工作模式
powerState: 电池连接状态，1：连接，0：未连接
audioType：当前语音类型
pkeType：PKE状态，1：开启，0：关闭
gps：gps定位信息,wgs84格式。若模块无法通过GPS定位，则不透出该字段，取而代之透出lbs或cell字段。timestamp(时间戳，单位：秒)，lat（纬度），lng（经度），speed（km/h），course（度°。其中正北为0，顺时针计算）。
cell:基站信息，若 GPS 已定位，该字段省略。mcc(mobile country code), mnc(mobile network code), lac(local area code), ci(cell id)。
lbs:通过基站定位得出的经纬度定位信息。wgs84格式。 
 
#### 错误返回体
 

```
{
  success:true,
  error:{
    errormessage:"",
    promot:""
  }
}

```

---


## 4. 获取设备轨迹

- **方法**：`GET`
- **URL**：`https://api.xiaoantech.com/ebike/v1/GPSPoints`

### Query 参数

| 参数 | 示例 | 说明 |
| --- | --- | --- |
| `agentId` | `87` | agent 标识 |
| `imei` | `868183030457282` | 15 位设备号 |
| `startTime` | `1527050032` | 起始时间，秒级时间戳 |
| `endTime` | `1527850032` | 结束时间，秒级时间戳 |

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |

### 说明

获取设备一段时间内的gps点，坐标为wgs84格式
 
#### 参数说明
 

startTime 和endTime 均为时间戳。单位：秒
 
#### 正确返回体
 

```
{
    "success": true,
    "data": [
        {
            "lat": 30.5114,
            "lon": 114.425667,
            "timestamp": 1528968201,
            "speed": 0,
            "course": 11
        },
        {
            "lat": 30.511604,
            "lon": 114.425522,
            "timestamp": 1528968201,
            "speed": 0,
            "course": 11
        }]

```

各个字段的含义：
lat：纬度
lon：经度
timestamp: GPS 时间戳，单位：秒
speed: 速度，km/h
course: 方位角，单位：°度。正北方为0，顺时针方向 
 
#### 错误返回体
 

```
{
    success:true,
    error:{
        errormessage:"",
        promot:""
    }
}

```

---


## 5. 开车/锁车[实时]

- **方法**：`POST`
- **URL**：`https://api.xiaoantech.com/ebike/v1/lock`

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |

### 请求体

```json
{
  "imei": "865532048622039",
  "locked": 1,
  "agentId": 87,
  "volume": 60,
  "isTBeacon": 0
}
```

### 说明

开车/锁车
 
#### 参数说明
 

imei:设备的15位唯一标识
locked: 1为锁车，0为开车。注意参数值为数字类型
idx:关联播放的铃声位。不填表示播放默认关联铃声
volume: 关联播放的铃声音量大小。0-100，整数。不填则按默认音量播放
isTBeacon： 是否使用道钉控制 1为使用 0为不使用（默认值） 
 
#### 正确返回体
 

```
{
    success: true,
    data: {
        code: 0
    }
}

```

code为0表示操作成功，code的具体含义参考错误码说明 
 
#### 错误返回体
 

```
{
    success:true,
    error:{
        errormessage:"",
        promot:""
    }
}

```

---


## 6. 电门控制[实时]

- **方法**：`POST`
- **URL**：`https://api.xiaoantech.com/ebike/v1/acc`

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |
| `Content-Type` | 是 | `application/json` |

### 请求体

```json
{
  "imei": 865532048775619,
  "acc": 1,
  "agentId": 87,
  "idx": 2,
  "volume": 60,
  "isTBeacon": 0
}
```

### 说明

电门控制
可用于车辆驶出服务区的断电操作等场景 
 
#### 参数含义
 

acc：0，关闭电门。 1，开启电门(同时撤防，效果同开车)。注意为数字类型。
idx:关联播放的铃声位。不填表示播放默认关联铃声
volume: 关联播放的铃声音量大小。0-100，整数。不填则按默认音量播放
isTBeacon： 是否使用道钉控制 1为使用 0为不使用（默认值） 
 
#### 正确返回体
 

```
{
    success: true,
    data: {
        code: 0
    }
}

```

code为0表示操作成功，code其他含义参考错误码说明
 
#### 错误返回体
 

```
{
    success:false,
    error:{
        errormessage:"",
        promot:""
    }
}

```

---


## 7. 设防控制[实时]

- **方法**：`POST`
- **URL**：`https://api.xiaoantech.com/ebike/v1/defend`

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |
| `Content-Type` | 是 | `application/json` |

### 请求体

```json
{
  "imei": 865067022403441,
  "defend": 1,
  "agentId": 87,
  "idx": 2,
  "volume": 60,
  "isTBeacon": 0
}
```

### 说明

#### 设防控制
 

可用于车辆设防状态的设置和解除 
 
#### 参数含义
 

defend：0，撤防。 1，设防。注意为数字类型。
idx:关联播放的铃声位。不填表示播放默认关联铃声
volume: 关联播放的铃声音量大小。0-100，整数。不填则按默认音量播放
isTBeacon： 是否使用道钉控制 1为使用 0为不使用（默认值） 
 
#### 正确返回体
 

```
{
    success: true,
    data: {
        code: 0
    }
}

```

code为0表示操作成功，code其他含义参考错误码说明
 
#### 错误返回体
 

```
{
    success:false,
    error:{
        errormessage:"",
        promot:""
    }
}

```

---


## 8. 后轮锁控制[实时]

- **方法**：`POST`
- **URL**：`https://api.xiaoantech.com/ebike/v1/backWheel`

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |
| `Content-Type` | 是 | `application/json` |

### 请求体

```json
{
  "imei": 868183030457282,
  "locked": 1,
  "agentId": 87,
  "idx": 2,
  "volume": 60
}
```

### 说明

后轮锁控制
参数含义：
locked：0，开锁。1，上锁。注意为数字类型。
idx:关联播放的铃声位。不填表示播放默认关联铃声
volume: 关联播放的铃声音量大小。0-100，整数。不填则按默认音量播放 
 
#### 正确返回体
 

```
{
    success: true,
    data: {
        code: 0
    }
}

```

code为0表示操作成功，code其他含义参考错误码说明
 
#### 错误返回体
 

```
{
    success:false,
    error:{
        errormessage:"",
        promot:""
    }
}

```

---


## 9. 电池仓锁(鞍座锁)控制[实时]

- **方法**：`POST`
- **URL**：`https://api.xiaoantech.com/ebike/v1/batteryCompartment`

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |

### 请求体

```json
{
  "imei": "868183030457282",
  "locked": 0,
  "agentId": 87,
  "idx": 2,
  "volume": 60
}
```

### 说明

电池仓锁(鞍座锁)控制 
 
#### 参数说明
 

imei:设备的15位唯一标识 locked: 1为电池仓上锁，0为电池仓解锁。注意参数值为数字类型
idx:关联播放的铃声位。不填表示播放默认关联铃声
volume: 关联播放的铃声音量大小。0-100，整数。不填则按默认音量播放 
 
#### 正确返回体
 

```
{
    success: true,
    data: {
        code: 0
    }
}

```

code为0表示操作成功，code其他含义参考错误码说明 
 
#### 错误返回体
 

```
{
    success:true,
    error:{
        errormessage:"",
        promot:""
    }
}

```

---


## 10. 播放语音[实时]

- **方法**：`POST`
- **URL**：`https://api.xiaoantech.com/ebike/v1/deviceVoice`

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |
| `Content-Type` | 是 | `application/json` |

### 请求体

```json
{
  "imei": "868183030457282",
  "index": 1,
  "agentId": 87,
  "volume": 60
}
```

### 说明

#### 参数说明
 

imei:设备的15位唯一标识
index:关联播放的铃声位。不填表示播放默认关联铃声
volume: 关联播放的铃声音量大小。0-100，整数。不填则按默认音量播放 
 
#### 铃声位说明
                                      
| index | 含义 |
| --- | --- |
| 1 | 车辆设防提示音 |
| 2 | 车辆撤防提示音 |
| 3 | 车辆启动提示音 |
| 4 | 车辆熄火提示音 |
| 5 | 车辆告警提示音 |
| 6 | 找车提示音 |
| 7~20 | 可根据客户需要进行定制 |
 
#### 正确返回体
 

```
{
    success: true,
    data: {
        code: 0
    }
}

```

code为0表示操作成功，code的其他含义参考错误码说明 
 
#### 错误返回体
 

```
{
    success:true,
    error:{
        errormessage:"",
        promot:""
    }
}

```

---


## 11. 注册事件回调链接

- **方法**：`POST`
- **URL**：`https://api.xiaoantech.com/ebike/v1/callback`

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |
| `Content-Type` | 是 | `application/json` |

### 请求体

```json
{
  "agentId": 87,
  "event": 1,
  "url": "http://pre.xiaoantech.com"
}
```

### 说明

注册事件回调链接。同一事件支持注册多个链接，可以根据使用需要注册生产环境或测试环境回调链接。
 

url：用于接收事件的接口url。必须为 POST 接口
event：注册事件类型，定义如下：
 
#### event含义
                                    
| event | 含义 | 上报周期 |
| --- | --- | --- |
| 1 | 车辆上报PING数据 | 150s |
| 2 | 车辆上报GPS数据 | 5s/运动状态，10min/静止状态 |
| 3 | 车辆上报事件通知 | 实时 |
| 4 | 车辆上报BMS信息 | 600s |
| 5 | 车辆上报UART/485数据 | 实时 |
 
#### 正确返回体
 

```
{
    "success": true,
    "data": 1//成功注册的url 数量
}

```

data表示增加的回调地址的数量,data为0表示该url已存在或者添加失败
 
#### 错误返回体
 

```
{
    success:false,
    error:{
        errormessage:"",
        promot:""
    }
}

```

#### 各事件回调数据说明
 
##### 1 车辆上报 PING 数据
 

返回数据字段是否存在与设备类型有关，定义如下：
 

```
{ 
  "success": true,
    "data":{
      "imei":"865067022403441",
      "event":1,
      "data":{
        "gsm": 20, //GPS 信号强度
      "voltage": 0 //电压值，单位：V  
      }
    }
}

```

##### 2 车辆上报 GPS 数据
 

返回数据字段是否存在与设备类型有关，定义如下：
 

```
{ 
  "success": true,
    "data":{
      "imei":"865067022403441",
      "event":2,
      "data":{
        "gsm": 22,//GSM 信号强度
        "voltage": 45835,//电压值，mV
        "timestamp": 1528968201,//GPS 时间戳，单位：秒
        "longitude": 114.4253158569336,//wgs84格式，经度
        "latitude": 30.511945724487305,//wgs84格式，纬度
        "speed": 0,//速度，km/h
        "course": 11,//方位角，单位：°度。正北方为0，顺时针方向。
        "hdop": 1.8600000143051147,//水平精度因子
        "satellite": 7 //GPS 定位卫星数量,
        "sw":86,//Bit 位开关
      }
    }
}

```
sw 各bit位含义                                                                  
| bit 位 | 标识 | 含义 |
| --- | --- | --- |
| 0 | isDefendOn | 设防 1/撤防 0 |
| 1 | isAccOn | 电门开 1/关 0 |
| 2 | isWheelunLocked | 后轮锁解锁 1/上锁 0 |
| 3 | isSeatLock | 电池仓(鞍座)锁上锁 1/解锁 0 |
| 4 | powerExistence | 电瓶连接 1/断开 0 |
| 5 | isGPSFastMode | GPS运动 1/静止 0 |
| 6 | isMoving | 车辆运动 1/静止 0 |
| 7 | isWheelSpan | 后轮转动 1/静止 0 |
| 8 | isEcoMode | 低功耗模式 1/正常模式 0 |
| 9 | isSleeping | 休眠模式 1/正常模式 0 |
| 10 | isDisturb | 移动告警触发语音 1/不触发 0 |
 
##### 3 车辆上报通知事件
 

返回通知事件定义如下：
 

```
{
  "success": true,
    "data":{
      "imei":"865067022403441",
      "event":3,
      "data":1,//notify类型
    }
}

```

事件支持情况与设备类型有关。
 事件定义                                                                                                                                             
| notify | 含义 | 适用设备类型 |
| --- | --- | --- |
| 0 | 车辆已自动设防 | 通用 |
| 1 | 车辆已上锁，开锁车接口-上锁 事件上报 | 车厂类设备 |
| 2 | 车辆已撤防 | 车厂类设备 |
| 3 | 设备登录网关服务器 | 通用 |
| 4 | 设备与网关服务器失去连接，设备掉电后5分钟触发该事件 | 通用 |
| 5 | 锁车/设防状态下移动报警 | 通用 |
| 6 | 电瓶移除报警 | 通用 |
| 7 | 电门开启通知，开锁车接口-开车事件上报 | 通用 |
| 8 | 电门关闭通知，开锁车接口-锁车事件上报 | 通用 |
| 9 | 低电量通知 | 通用 |
| 10 | 设备与网关服务器主动断开连接 | 通用 |
| 11 | 鞍座(电池仓)已开锁 | 需外设锁支持到位信号 |
| 12 | 鞍座(电池仓)已上锁 | 需外设锁支持到位信号 |
| 13 | 车辆已自动设防 | 共享类设备 |
| 14 | 车辆已自动设防 | 车厂类设备 |
| 15 | 车辆已上锁，开锁车接口事件上报 | 车厂类设备 |
| 16 | 车辆已解锁，开锁车接口-开车/蓝牙-开电门命令/蓝牙-撤防命令/感应启动 事件上报 | 通用 |
| 17 | 出地理围栏通知 | 通用 |
| 18 | 入地理围栏通知 | 通用 |
| 19 | 锁车/设防状态下移动报警 | 通用 |
| 20 | 锁车/设防状态下震动报警 | 通用 |
| 21 | 电量剩余50% | 共享类设备 |
| 22 | 电量剩余30% | 共享类设备 |
| 23 | 电瓶恢复连接通知 | 通用 |
| 55 | 超载事件触发 | 通用 |
| 56 | 超载事件解除 | 通用 |
 
##### 4 车辆上报 BMS 数据
 

返回数据字段是否存在与设备类型有关，定义如下：
 

```
{ 
  "success": true,
    "data":{
      "imei":"865067022403441",
      "event":4,
      "data":{
        "sn": "ZTZL1909012307",// BMS串号 如无, 默认0x0000
        "hardVersion": 0,// BMS硬件版本 如无, 默认0x0000
        "softVersion": 120,//BMS软件版本 如无, 默认0x0000
        "MOSTemp": 27,// MOS温度 单位: 0.1摄氏度
        "MOSState": 0,// BMS MOS状态: 0x00: 无此功能 0x01: 充电状态 0x02: 放电状态 0x03: 存储状态
        "maxVoltage": 179,// 单体电池最大电压 单位: 0.1V  如无, 默认0x00
        "minVoltage": 105,// 单体电池最小电压 单位: 0.1V 如无, 默认0x00
        "healthState": 100,// 电池健康状态百分比 电池健康状态百分比
        "fault": 0, // 故障位开关:详述见下面
        "capacity":18000,// 电池绝对满充容量 mAH,
        "remainCapacity": 0,// 电池相对剩余容量 单位: mAH
        "soc": 0,// 电池相对剩余容量 单位: mAH
        "cycle": 70,// 电池循环次数
        "voltage": 4982,// 总实时电压 单位: mV
        "current": 0,// 总实时电流 单位: 0.1A (+:放电电流 -:充电电流)
        "timestamp": 1591798432,// 数据时间戳 单位:秒
      }
    }
}

```
sw 各bit位含义                                                          
| bit 位 | 含义 |
| --- | --- |
| 0 | 充电过温 |
| 1 | 短路 |
| 2 | 放电过流 |
| 3 | 充电过流 |
| 4 | 欠压 |
| 5 | 过压 |
| 6 | 单电池欠压 |
| 7 | 单电池过压 |
| 8 | 采集错误 |
| 9 | 放电低温 |
| 10 | 充电低温 |
| 11 | 充电低温 |
 
##### 5 车辆上报UART/485数据
 

返回数据定义如下：
 

```
{
  "success": true,
    "data":{
      "imei":"865067022403441",
      "event":5,
      "data":"MTIzNDU2QUJDREVGR0g=",//Base64Encode(Buffer)
      "tm":1547538878
    }
}

```

---


## 12. 获取事件回调链接

- **方法**：`GET`
- **URL**：`https://api.xiaoantech.com/ebike/v1/callback`

### Query 参数

| 参数 | 示例 | 说明 |
| --- | --- | --- |
| `agentId` | `87` | agent 标识 |
| `event` | `1` | 事件类型，见注册接口 |

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |
| `Content-Type` | 是 | `application/json` |

### 请求体

```json
{
  "agentId": 87,
  "event": 1,
  "url": "http://pre.xiaoantech.com"
}
```

### 说明

#### 参数含义
 

event:回调事件类型，具体说明请参《考注事件回调链接》事件说明
 
#### 正确返回体
 

```
{
    "success": true,
    "data": [
        "http://pre.xiaoantech.com"
    ]
}

```

data为注册的该事件的所有回调地址的数组
 
#### 错误返回体
 

```
{
    success:false,
    error:{
        errormessage:"",
        promot:""
    }
}

```

---


## 13. 注销事件回调链接

- **方法**：`DELETE`
- **URL**：`https://api.xiaoantech.com/ebike/v1/callback`

### Query 参数

| 参数 | 示例 | 说明 |
| --- | --- | --- |
| `agentId` | `87` | agent 标识 |
| `event` | `1` | 事件类型 |
| `url` | `http://pre.xiaoantech.com` | 要注销的回调地址 |

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |

### 说明

#### 参数含义
 

event:回调事件类型，具体说明请参《考注事件回调链接》事件说明
url：需要注销的回调链接
 
#### 正确返回体
 

```
{
    "success": true,
    "data": 1
}

```

data表示注销的回调地址数量，data为0表示该回调地址已被注销或者本次注销失败
 
#### 错误返回体
 

```
{
    success:false,
    error:{
        errormessage:"",
        promot:""
    }
}

```

---


## 14. 配置蓝牙参数[实时]

- **方法**：`POST`
- **URL**：`https://api.xiaoantech.com/ebike/v1/bluetooth`

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |

### 请求体

```json
{
  "imei": "868183030409937",
  "token": 1531279386,
  "agentId": 87,
  "name": "yuean"
}
```

### 说明

配置蓝牙参数。接口返回成功后，等60s待设备重启，生效配置。 
 
#### 参数说明：
 

token：蓝牙通信的token（u32），可单独配置,若不配置,可以忽略, 0 到 4,294,967,295(2^32-1) 之间的无符号整数
name：蓝牙名称 {0:15} 可单独配置,若不配置,可以忽略
agentId：代理商id imei：设备编号
 
#### 正确返回体
 

```
{
    success: true,
    data: {
        code: 0,
        tm: 1531279433
    }
}

```

code为0表示操作成功，code其他含义参考错误码说明
 
#### 错误返回体
 

```
{
    success:false,
    error:{
        errormessage:"",
        promot:""
    }
}

```

---


## 15. 重启设备[实时]

- **方法**：`POST`
- **URL**：`https://api.xiaoantech.com/ebike/v1/reboot`

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |

### 请求体

```json
{
  "agentId": 87,
  "imei": "868183032288834"
}
```

### 说明

重启设备 
 
#### 参数说明
 

imei:设备的15位唯一标识

---


## 16. 电池放电开关控制[实时]

- **方法**：`POST`
- **URL**：`https://api.xiaoantech.com/ebike/v1/batteryPowerSwitch`

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |
| `Content-Type` | 是 | `application/json` |

### 请求体

```json
{
  "imei": 868183030457282,
  "sw": 1,
  "agentId": 87
}
```

### 说明

参数含义：
sw：0，关闭放电开关。 1，打开放电开关。注意为数字类型。
 
#### 正确返回体
 

```
{
    success: true,
    data: {
        code: 0
    }
}

```

code为0表示操作成功，code其他含义参考错误码说明
 
#### 错误返回体
 

```
{
    success:false,
    error:{
        errormessage:"",
        promot:""
    }
}

```

---


## 17. 设置控制器限速百分比[实时]

- **方法**：`POST`
- **URL**：`https://api.xiaoantech.com/ebike/v1/mc`

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |

### 请求体

```json
{
  "imei": "868183036822257",
  "speed": 43,
  "agentId": 87
}
```

### 说明

#### 参数说明
 

imei:设备的15位唯一标识
speed: 车速百分比。单位：%。数字整数类型
 
#### 正确返回体
 

```
{
    success: true,
    data: {
        code: 0
    }
}

```

code为0表示操作成功，code的具体含义参考错误码说明 
 
#### 错误返回体
 

```
{
    success:true,
    error:{
        errormessage:"",
        promot:""
    }
}

```

---


## 18. 获取车辆电池BMS信息[实时]

- **方法**：`GET`
- **URL**：`https://api.xiaoantech.com/ebike/v1/currentBmsInfo`

### Query 参数

| 参数 | 示例 | 说明 |
| --- | --- | --- |
| `agentId` | `87` | agent 标识 |
| `imei` | `865532040744076` | 15 位设备号 |

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |

### 说明

与设备通信获取车辆电池BMS信息。
 
#### 参数说明
 

imei:设备的15位唯一标识
 
#### 正确返回体
 

```
{
    "success": true,
    "data": {
        "voltage": 49427,
        "current": 65317,
        "temperature": 257,
        "SOC": 45,
        "remain": 11005
    }
}

```

各个字段的含义：
voltage:电压。单位:mV
current:电流
temperature:温度
SOC:剩余电量百分比。单位：% 
 
#### 错误返回体
 

```
{
  success:true,
  error:{
    errormessage:"",
    promot:""
  }
}

```

---


## 19. 获取设备最新定位地址信息

- **方法**：`GET`
- **URL**：`https://api.xiaoantech.com/ebike/v1/address`

### Query 参数

| 参数 | 示例 | 说明 |
| --- | --- | --- |
| `agentId` | `87` | agent 标识 |
| `imei` | `868183030470848` | 15 位设备号 |

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |

### 说明

获取设备最后定位点信息 
 
#### 参数含义
 

imei：设备编号 
 
#### 正确返回体
 

```
{
    success: true,
    data: {
        timestamp: 1541683308,
        address: "嘉鱼县新街镇新村路32号北",
        originInfo: {}
    }
}

```

若有其他地址信息需求，可从originInfo中解析相应字段。经纬度坐标系：gcj02 
 
#### 错误返回体
 

```
{
    success:false,
    error:{
        errormessage:"",
        promot:""
    }
}

```

---


## 20. 基站定位数据转GPS定位

- **方法**：`POST`
- **URL**：`https://api.xiaoantech.com/ebike/v1/lbs2gps`

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |
| `Content-Type` | 是 | `application/json` |

### 请求体

```json
{
  "agentId": 87,
  "cell": {
    "mcc": 460,
    "mnc": 0,
    "cells": [
      {
        "lac": 28734,
        "ci": 44603,
        "dBm": -57
      },
      {
        "lac": 28734,
        "ci": 54186,
        "dBm": -73
      },
      {
        "lac": 28734,
        "ci": 44601,
        "dBm": -74
      },
      {
        "lac": 28734,
        "ci": 54602,
        "dBm": -77
      },
      {
        "lac": 28712,
        "ci": 232,
        "dBm": -81
      }
    ]
  }
}
```

### 说明

参数含义：
cell: GSM/UMTS，兼容两种格式的数据: MCC,MNC,LAC,CI,RXL 或 MCC,MNC,LAC,CI,dBm 。dBm=RXL-113
 
#### 格式1
 

```
{   
    "agentId": 87,
    "cell": {
        "mcc": 460,
        "mnc": 0,
        "cells": [
            {
                "lac": 28734,
                "ci": 44603,
                "dBm": -57
            },
            {
                "lac": 28734,
                "ci": 54186,
                "dBm": -73
            },
            {
                "lac": 28734,
                "ci": 44601,
                "dBm": -74
            },
            {
                "lac": 28734,
                "ci": 54602,
                "dBm": -77
            },
            {
                "lac": 28712,
                "ci": 232,
                "dBm": -81
            }
        ]
    }
}

```

#### 格式2
 

```
{   
    "agentId": 87,
    "cell": {
        "mcc": 460,
        "mnc": 0,
        "cells": [
            {
                "lac": 28734,
                "ci": 44603,
                "rxl": 56
            },
            {
                "lac": 28734,
                "ci": 54186,
                "rxl": 40
            },
            {
                "lac": 28734,
                "ci": 44601,
                "rxl": 39
            },
            {
                "lac": 28734,
                "ci": 54602,
                "rxl": 36
            },
            {
                "lac": 28712,
                "ci": 232,
                "rxl": 22
            }
        ]
    }
}

```

#### 正确返回体
 

```
{
    "success": true,
    "data": {
        "count": 5,
        "result": [
            {
                "id": "460-000-0000028734-00000000000000044603",
                "lat": "30.50451142",
                "lng": "114.43048877",
                "radius": "764",
                "address": "湖北省武汉市洪山区关东街道瑞成佳苑关东康居园",
                "roads": "文园路东约36米",
                "lats": "30.502173",
                "lngs": "114.435982",
                "rid": "420111",
                "rids": "420111080000"
            },
            {
                "id": "460-000-0000028734-00000000000000054186",
                "lat": "30.50234387",
                "lng": "114.42816352",
                "radius": "834",
                "address": "湖北省武汉市洪山区关东街道光谷创业街52号中国武汉留学生创业园(光谷创业街)",
                "roads": "百合路西南约38米",
                "lats": "30.500011",
                "lngs": "114.433667",
                "rid": "420111",
                "rids": "420111080000"
            },
            {
                "id": "460-000-0000028734-00000000000000044601",
                "lat": "30.50260513",
                "lng": "114.43124154",
                "radius": "732",
                "address": "湖北省武汉市洪山区关东街道光谷创业街关山村上下毕还建房",
                "roads": "光谷创业街西北约4米",
                "lats": "30.500267",
                "lngs": "114.436734",
                "rid": "420111",
                "rids": "420111080000"
            },
            {
                "id": "460-000-0000028734-00000000000000054602",
                "lat": "30.50333362",
                "lng": "114.43244187",
                "radius": "762",
                "address": "湖北省武汉市洪山区关山街道光谷创业街加油加气站(光谷创业街)(装修中)",
                "roads": "光谷创业街南约0米",
                "lats": "30.500996",
                "lngs": "114.437935",
                "rid": "420111",
                "rids": "420111002000"
            },
            {
                "id": "460-000-0000028712-00000000000000000232",
                "lat": "30.50548567",
                "lng": "114.42741022",
                "radius": "754",
                "address": "湖北省武汉市洪山区关东街道百合路关东科技工业园电子港",
                "roads": "百合路西约27米",
                "lats": "30.503152",
                "lngs": "114.432913",
                "rid": "420111",
                "rids": "420111080000"
            }
        ],
        "latitude": 30.5036521024309,
        "longitude": 114.429900338729
    }
}

```

code为0表示操作成功，code其他含义参考错误码说明
 
#### 错误返回体
 

```
{
    success:false,
    error:{
        errormessage:"",
        promot:""
    }
}

```

---


## 21. 获取服务端缓存的电池信息

- **方法**：`GET`
- **URL**：`https://api.xiaoantech.com/ebike/v1/batteryInfo`

### Query 参数

| 参数 | 示例 | 说明 |
| --- | --- | --- |
| `agentId` | `87` | agent 标识 |
| `imei` | `865532040744076` | 15 位设备号 |

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |

### 说明

对于设备在线状态，由于设备与网关服务器的保活机制，设备登录网关服务器时立即判定为设备上线，而在设备与网关服务器失去连接5分钟后判定为离线。
 
#### 正确返回体
 

```
{
    succcess:true,
    data:{
        imei: "868183030457282",
        isOnline: 1,    
        version: 67328,   
        capacity: 30000,
        cycle: 23,
        SN: 6756886678668800,
        remaining: 16000,
        temperature: 54,
        isUndervoltage: 1,
        isOvervoltage: 0,
        isOverload: 1,
        isHighTemperature: 0,
        isDischargeShortcircuit: 0,
        isChargeShortcircuit: 1,
        isDischargeOvercurrent: 0,
        isChargeOvercurrent: 1
    }
}

```

各个字段的含义：
imei：设备号
isOnline: 设备在线状态，1：在线，0：离线
version：BMS版本号
capacity：电池总容量，单位:mAH
cycle：电池循环次数
SN：电池串号，16位,末尾补0
remaining：电池剩余容量，单位:mAH
temperature：电池温度，单位:摄氏度
isUndervoltage：是否欠压，1/0
isOvervoltage：是否过压，1/0
isOverload：是否过载，1/0
isHighTemperature：是否高温，1/0
isDischargeShortcircuit：是否放电短路，1/0
isChargeShortcircuit: 是否充电短路，1/0
isDischargeOvercurrent: 是否放电过流，1/0
isChargeOvercurrent: 是否充电过流，1/0 
 
#### 错误返回体
 

```
{
    succcess:true,
    error:{
        errormessage:"",
        promot:""
    }
}

```

---


## 22. 获取车辆电池BMS信息

- **方法**：`GET`
- **URL**：`https://api.xiaoantech.com/ebike/v1/bmsInfo`

### Query 参数

| 参数 | 示例 | 说明 |
| --- | --- | --- |
| `agentId` | `87` | agent 标识 |
| `imei` | `865532040744076` | 15 位设备号 |

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `Content-Type` | 是 | `application/json` |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |

### 说明

获取车辆电池BMS信息。
 
#### 参数说明
 

imei:设备的15位唯一标识
 
#### 正确返回体
 

```
{
    "success": true,
    "data": {
        "SN": "57402AIAJAAW2373",
        "capacity": 26880,
        "cycle": 0
    }
}

```

各个字段的含义(字段是否存在与设备类型相关)：
SN：电池BMS SN号
capacity：电压。单位：mV
cycle：循环次数 
 
#### 错误返回体
 

```
{
  success:true,
  error:{
    errormessage:"",
    promot:""
  }
}

```

---


## 23. 短信接口[实时]

- **方法**：`POST`
- **URL**：`https://api.xiaoantech.com/ebike/v1/sms`

### Header

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `xc-access-token` | 是 | 访问凭证，每个请求必带 |
| `Content-Type` | 是 | `application/json` |

### 请求体

```json
{
  "imei": "868183033267985",
  "cmd": {
    "c": 3,
    "p": {
      "server": "server.xiaoantech.com:9880"
    }
  },
  "agentId": 87
}
```

### 说明

#### 参数含义
 

imei:设备号
agentId:客户id
cmd:车辆设备的短信控制命令，详细定义参见《中控设备SMS通信协议》 
 
#### 格式
 

```
{
    "imei": "868183033267985",
    "cmd": {
            "c": 404,
            "p": {
                 "param":"server.host.com:port"
            }
        },
    "agentId": 1
}

```

#### 正确返回体
 

```
{
    "success": true,
    "data": {
        "msg": "提交成功",
        "data": {
            "rows": [
                {
                    "msisdn": "1440457359015",
                    "sms_id": 17318552
                }
            ],
            "failed": []
        },
        "code": 0
    }
}

```

code为0表示操作成功，code其他含义参考本文档-术语解释-code含义对照表 
 
#### 错误返回体
 

```
{
    success:false,
    error:{
        errormessage:"",
        promot:""
    }
}

```

---
