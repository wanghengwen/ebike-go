# ebike-UniApp

luopingtech 共享电单车客户端（uni-app Vue3 + Vite）。

同一套业务代码输出：

- 微信小程序 `mp-weixin`
- H5
- Android / iOS（`app-plus`）

本工程位于 Go monorepo 子目录 `ebike-go/ebike-UniApp`，不单独建 git，不附带 LICENSE。

## 环境

- Node.js **22+**
- npm 10+

```bash
cd ebike-UniApp
npm install
npm run dev:mp-weixin
# 或
npm run dev:h5
npm run dev:app
```

## 构建

```bash
npm run build:mp-weixin
npm run build:h5
npm run build:app   # 产物用 HBuilderX 云打包/离线打包
```

## 目录

```text
src/
  pages/          主包：启动/首页/地图/骑行/支付/登录/个人中心
  pages-sub/      分包：ride / pay / account / support
  features/       bike · ble · pay · map · auth
  widgets/        自研 UI
  locales/        vue-i18n（zh-CN / en-US）
  shared/         request · storage · config · protocol · logger
  vendor/         地图/支付薄封装
  stores/         pinia
  api/            gateway API（对齐旧 wechat 业务接口）
  tenants/        运行时 schema / merged.runtime（配置源在 config/）
scripts/          租户 import / merge / build
config/           多租户发布包 config/{env}_{mode}.json（对齐旧 wechat/config）
```

## 已覆盖能力

- 登录（微信一键 / 短信）、实名、职业认证、注销
- 地图寻车、扫码/编号开锁、网络+蓝牙开还车、临停
- 钱包/充值/提现/退押金/发票/卡券/兑换券/活动
- 行程列表与详情、费用明细、轨迹、费用异议
- 消息、FAQ、客服、报修、违规举报、申请站点、还车申请
- 信用分 / 微信支付分、邀请有礼、计费规则、用车资格
- i18n、多租户 config 发布、无扬歌广告

## 租户配置

以 `config/{env}_{mode}.json` 为准（含 `platformSecret` / `platformSign`，与旧版一致）：

```bash
npm run tenant:import
npm run build:mp-weixin -- --env=xiaolongyu --mode=release
```

详见 `config/README.md`。`tenants/local.secrets.json` 仅可选本机覆盖，一般不需要。

## Android / iOS

`manifest.json` 已启用 Maps / Payment / Bluetooth / Camera / Geolocation。  
包名前缀建议 `com.luopingtech.ebike.*`。真机冒烟：定位、扫码、BLE、支付。

## 明确不做

- 扬歌广告及广告 SDK
- hello-uniapp / ebikeMinaSass 等旧模板标识
- 仓库内整包拷贝的 vant / uni-ui 源码
