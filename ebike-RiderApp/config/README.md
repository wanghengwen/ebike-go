# 租户配置

文件名：`{tenant}_{mode}.json`，由 `local.properties` 的 `rider.tenant` / `rider.mode` 选择。

## 与 UniApp 字段对照

| UniApp | RiderApp |
| --- | --- |
| `platformTenantId` | `tenantId` |
| `platformSecret` | `auth.businessSecret`（Basic `tenantId:secret`） |
| `platformSign` | `auth.signSecret`（`_s` 签名） |
| `api.baseUrl` | `api.baseUrl` |
| `api.logApi` | `api.logUploadUrl` |
| `customSetting.buttonGreenColor` 等 | `branding.*` |
| `customSetting.documentCfg.userProtocol` / `privacyProtocol` | `documents.*` |
| `platform.h5.router.base` | 仅说明部署路径；原生要填完整 `h5.baseUrl`（含 `index.html`） |

## 安全

- `*_release.json` / `*_debug.json` **不入库**（`demo_release.json` 除外）
- 密钥只放本机 `config/` 或 CI 机密，不要提交

## 当前联调租户

推荐：`renren` + `release` → `https://client.luopingtech.com`，密钥来自 `ebike-UniApp/config/renren_release.json`。

```
# local.properties
rider.tenant=renren
rider.mode=release
```

`h5.baseUrl`：部署 UniApp `npm run build:h5` 产物后填完整入口，例如 `https://cdn.example.com/h5/test/index.html`。空则 App 内只显示 H5 占位。

## 腾讯地图 Key

与 `ebike-OpsApp/config/renren_release.json` 同源：`map.tencentKey` /
`local.properties` 的 `rider.map.tencentKey`。

注意：腾讯控制台按 **包名 / Bundle ID** 鉴权。OpsApp renren 登记的是运维端包名；
Rider 当前是 `com.luopingtech.ebike.rider.renren`。若地图白屏，需在腾讯位置服务控制台
把 Rider 包名加进同一 Key，或另开 Key。
