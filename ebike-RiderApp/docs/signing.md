# 请求签名

C 端沿用旧商户 App 的签名契约，网关那边不认第二种写法。三个 header：

| header | 含义 |
| --- | --- |
| `_t` | 客户端时间戳，epoch 毫秒的十进制字符串 |
| `_s` | 签名，SHA-256 的小写 hex（64 字符） |
| `Authorization` | `Bearer {accessToken}`，或 `/oauth/token` 上的 `Basic base64(tenantId:businessSecret)` |

实现在 `shared/core/signing/RequestSigner.kt`，header 组装在
`shared/core/network/HttpClientFactory.kt` 的 `RequestAuth`。

## 三种明文拼法

| 场景 | 明文 |
| --- | --- |
| POST JSON | `{body}` + `_t={timestamp}{signSecret}` |
| GET / 排序参数 | 按 key 排序的 `k=v&` 逐个拼接，再接 `_t={timestamp}{signSecret}` |
| 表单式（旧 `tenant/queryList` 那条路） | 按 key 排序的 `k=v&`，再接 `time={timestamp}&secret={businessSecret}` |

几个容易踩的点：

- **JSON 体原样进签名串**，不做任何转义。体里的 `+` / `&` / `=` 都是字面量。
- **按 UTF-8 字节算**，不是按字符。中文体是最容易掉进 Latin-1 的一格。
- **公共请求体的键序是签名的一部分**。`CommonRequestBody` 用 `buildJsonObject` 按写入
  顺序输出 `traceId` / `platform` / `deviceId` / `tenantId`，改字段顺序等于换掉所有签名。
- `Accept-Language` 只走 header，不进 body —— 放进去会一起进签名串，换个语言签名就变了。

## 黄金向量

固定 `timestamp = 1700000000000`、`signSecret = s3cret`、`businessSecret = biz-secret`、
`tenantId = 1001`。

期望 hex 由**仓库外的** SHA-256 实现（.NET `System.Security.Cryptography`）对 payload 列
算出，再写进断言 —— 用项目自己的 `Sha256` 算了再回填只能证明代码没变，证明不了算法对。
断言在 `shared/src/commonTest/.../core/signing/RiderSignVectorsTest.kt`。

| # | 场景 | payload（明文，逐字节） | 期望 `_s` |
| --- | --- | --- | --- |
| V1 | POST JSON 基线 | `{"phone":"13800000000"}_t=1700000000000s3cret` | `a57fc306490f67409b0bdf770a672b97f066e71fb62c159d2c30062664cd4c8f` |
| V2 | C 端公共体 | `{"traceId":"11111111-1111-4111-8111-111111111111","platform":"android","deviceId":"dev-abc","tenantId":"1001"}_t=1700000000000s3cret` | `7c7ef6681ea2e1a07df307911e51242b49f69572b880c103d5fd869a772bb298` |
| V3 | 空体 | `{}_t=1700000000000s3cret` | `1039e50112b6794e94322ccd929949d58ed8e4e27df0bf71a2b26d171c2accc2` |
| V4 | 中文 UTF-8 | `{"name":"张三"}_t=1700000000000s3cret` | `162c83b932820c62e572b162addb84727b11567eb33bf5f55b328e61a21bed7b` |
| V5 | 特殊字符 | `{"q":"a+b&c=d","note":"1+1=2&ok"}_t=1700000000000s3cret` | `64d826d4ff07483ac3343fb1090919871e1a86c7f2983d4da50c9c507fa0d5c7` |
| V6 | GET 排序 | `a=1&b=2&_t=1700000000000s3cret` | `105dbecf3895d681684f70ef6d2d1d213fd1d9e43ad4762f4fb790f994f2df6e` |
| V7 | 表单式 | `izFilterRootTenant=true&phone=+86-13800138000&time=1700000000000&secret=biz-secret` | `96aa9481fd451e888b15f23c201b6801e3d6628780d11ec24b92723112e914c2` |

V6 的 payload 注意末尾那个 `&`：排序参数是每个 `k=v` 后面都跟一个 `&`，最后一段
`_t=...` 前面因此一定有分隔符。

Authorization 的 Basic：`base64("1001:biz-secret")` = `MTAwMTpiaXotc2VjcmV0`。

另有一条 NIST 向量守着 SHA-256 本身：`Sha256.hex("")` =
`e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`。

## 三栈比对

payload 列是逐字节的明文，任一栈直接对它做 SHA-256 就该得到同一个 hex。

Node（UniApp 侧）：

```js
const { createHash } = require('node:crypto')
const hex = (s) => createHash('sha256').update(s, 'utf8').digest('hex')
console.log(hex('{"phone":"13800000000"}_t=1700000000000s3cret'))
// a57fc306490f67409b0bdf770a672b97f066e71fb62c159d2c30062664cd4c8f
```

Go（网关侧）：

```go
sum := sha256.Sum256([]byte(`{"phone":"13800000000"}_t=1700000000000s3cret`))
fmt.Println(hex.EncodeToString(sum[:]))
// a57fc306490f67409b0bdf770a672b97f066e71fb62c159d2c30062664cd4c8f
```

Kotlin（本仓库）：

```
./gradlew :shared:testAndroidHostTest --tests '*RiderSignVectorsTest*'
```

## 时钟注入

`RequestAuth` 的 `clock` 参数默认 `nowEpochMillis()`，测试传固定值：

```kotlin
val auth = RequestAuth(config, InMemorySecureStore(), clock = { 1700000000000L })
```

`RiderSignVectorsTest.requestAuth_withFixedClockEmitsGoldenHeaders` 靠它做端到端断言 ——
只测 `RequestSigner` 拦不住 header 组装与 signer 走岔的情况。
