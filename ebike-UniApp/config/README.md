# 多租户配置发布

旧版：`npm run build:mp-weixin --env=xiaolongyu --mode=release`  
新版：同样以 `config/{env}_{mode}.json` 为**唯一必需**配置源（含 `platformSecret` / `platformSign`）。

## 目录

| 文件 | 说明 |
|------|------|
| `_defaults.json` | 全量默认（皮肤字段骨架、默认支付通道等） |
| `{env}_{mode}.json` | **租户完整配置**（品牌、API、appid、支付、皮肤、签名密钥） |
| `../tenants/local.secrets.json` | **可选**本机覆盖（仅非空字段），一般不需要 |

## 发布

```bash
# 按租户构建（会先 merge config → manifest/pages/runtime）
npm run build:mp-weixin -- --env=xiaolongyu --mode=release
```

## 哪些不该写进 JSON

人脸核验、文明还车、头盔弹窗、广告位 —— 由运行时接口按 `serviceId` 控制：

- `getUseCarConfig.izOnUseCar`
- `getbackCarConfig.izCivilizationRemind` / `dispatchFee` / 免罚项
- 用户 `needFaceCheck`
- 订单 `helmetPopup` / `returnType`

静态 JSON 负责：品牌名、tenantId、签名密钥、API 域名、微信 appid、支付 channelType、皮肤资源、少量租户业务开关（如 `autoRefundBalance`）。
