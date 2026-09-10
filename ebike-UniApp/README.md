# ebike-UniApp

共享电单车客户端（uni-app Vue3 + Vite）。

同一套业务代码输出：

- 微信小程序 `mp-weixin`
- H5
- Android / iOS（`app-plus`）

## 已实现功能

按主包 / 分包与业务域梳理（页面与 `features/`、`api/` 对齐）：

### 启动与首页

- 启动页、首页地图（站点/车辆/围栏、快捷入口、引导与运营弹窗）
- 全屏地图（定位、选车、寻车铃、导航、扫码用车）

### 账号与认证

- 微信一键登录 / 手机号登录、实名认证、人脸核验
- 个人中心、设置、换绑手机、账号注销、关于
- 职业认证、用车资格、信用分、微信支付分、绑卡

### 用车与骑行

- 扫码 / 输入编号开锁（预开锁确认）
- 网络开锁 + 蓝牙回退、骑行中（临停、还车、导航、围栏提示）
- 还车引导、文明还车、头盔提醒、申请还车、行程轨迹
- 红包车攻略、停车点搜索

### 支付与资产

- 行程费用支付、费用明细、费用异议
- 钱包（可用/充值/赠送余额）、充值、提现、退押金、流水
- 骑行卡购买与持有、兑换券、开发票

### 行程与运营活动

- 我的行程列表 / 详情
- 邀请有礼（分享、记录、规则）
- 活动页、计费规则

### 客服与报修

- 帮助中心、FAQ、消息中心、在线客服入口
- 车辆报修（部位选择、记录、进度）
- 违规举报、申请还车站点

### 工程能力

- vue-i18n（zh-CN / en-US）
- 多租户 `config/{env}_{mode}.json` 构建发布
- BLE 指令编解码、地图/支付 vendor 薄封装

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

按租户构建示例：

```bash
npm run build:mp-weixin -- --env=xiaolongyu --mode=release
```

## 目录

```text
ebike-UniApp/
  config/                 多租户发布包 config/{env}_{mode}.json（正式密钥勿入库）
  scripts/                tenant import/merge、run-build、分包 JS 搬迁
  tenants/                local.secrets.example（可选本机覆盖）
  src/
    pages/                主包页面
      launch/             启动
      home/               首页地图
      map/                全屏地图
      riding/             骑行中
      pay/                行程费用支付
      auth/               登录 / 实名
      account/            个人中心 / 设置
      webview/            内嵌页
    pages-sub/            分包
      ride/               预开锁、扫码、还车、轨迹等
      pay/                钱包、充值、提现、发票、费用明细等
      account/            行程、卡券、活动、邀请、信用等
      support/            帮助、消息、报修、举报、申请站点等
    features/             业务编排
      auth/ bike/ ble/ credit/ guide/ map/ pay/ support/
    widgets/              弹层与业务组件（引导、开锁提示、还车等）
    api/                  gateway API（对齐旧 wechat 业务接口）
    shared/               request · storage · config · navigate · logger 等
    vendor/               地图 / 支付薄封装
    stores/               pinia（user / locale / tempData）
    locales/              vue-i18n（zh-CN / en-US）
    tenants/              运行时 schema / merged.runtime（由 config merge 生成）
    styles/ static/       全局样式与静态资源
```

## 租户配置

以 `config/{env}_{mode}.json` 为准（含 `platformSecret` / `platformSign`，与旧版一致）。  
正式环境配置勿提交版本库；本地可保留，构建时通过 `--env` / `--mode` 指定。

详见 `config/README.md`。`tenants/local.secrets.json` 仅可选本机覆盖，一般不需要。

## Android / iOS

`manifest.json` 已启用 Maps / Payment / Bluetooth / Camera / Geolocation。  
包名前缀建议 `com.luopingtech.ebike.*`。真机冒烟：定位、扫码、BLE、支付。
