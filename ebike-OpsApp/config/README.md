# 多租户配置

沿用 `ebike-UniApp` 的约定：`config/{tenant}_{mode}.json` 是唯一必需的配置源，构建时按参数选择。

## 目录

| 文件 | 是否入库 | 说明 |
|------|---------|------|
| `_defaults.json` | 是 | 字段骨架与默认值，不含任何密钥 |
| `demo_release.json` | 是 | 可公开的 demo 租户，密钥字段留空 |
| `renren_debug.json` / `renren_release.json` | **否** | 人人运维联调配置（源自 `Merchant-Android/.../renren.gradle`） |
| `{tenant}_{mode}.json` | **否** | 租户完整配置，含 `businessSecret` / `signSecret` / 地图 Key |
| `local.properties` | **否** | 本机构建参数与签名库口令，模板见 `local.properties.example` |

`.gitignore` 已排除除 demo 之外的所有 `*_release.json` / `*_debug.json`。

## 人人运维联调

1. `local.properties` 设置：

```properties
ops.tenant=renren
ops.mode=release
```

2. 构建时 `androidApp` 会把 `config/renren_release.json` 同步到 `assets/tenant.json`，并把 `applicationId` 设为 `com.renren.inspections`。
3. 字段映射（相对遗留 `renren.gradle`）：

| 遗留 BuildConfig | 新配置字段 |
|------------------|------------|
| `BASE_URL_*` | `api.baseUrl` |
| `SECRET_VALUE_*` | `auth.signSecret`（请求签名） |
| `BUSSINESS_SECRET_*` | `auth.businessSecret`（OAuth Basic） |
| `BUSSINESS_ID_*` | `tenantId` |
| `TENCENT_MAP_KEY` | `map.tencentKey` |
| `app_name` / `themeColor` | `app.displayName` / `branding.primaryColor` |

Debug / Release 在原 flavor 中指向同一生产域名与同一组密钥；新项目分别提供 `renren_debug.json`、`renren_release.json`，内容一致，便于后续拆环境。

## 字段归属

静态 JSON 只放**构建期不变**的内容：品牌名、包名、API 基址、租户 ID、签名密钥、
地图与推送 Key、平台能力开关。

运行时由服务端接口下发的内容不要写进 JSON，例如权限码与菜单可见性、运维阈值、
任务类型配置、围栏策略。遗留工程把部分此类配置放在 flavor 里，导致改配置要发版，本项目不再这样做。

## 主题色（branding）

对应遗留 Android flavor 的 `themeColor*` / `textColorBlack*`，写入 `branding` 段后由
`OpsTheme(branding = app.config.branding)` 注入 Compose；H5 打开时会附带 `themeColor` /
`lightTxtColor` 查询参数。

| 配置字段 | 遗留 resValue | 用途 |
|----------|---------------|------|
| `primaryColor` | `themeColor` | 主色（按钮、选中 Tab、顶栏） |
| `primaryMutedColor` | `themeColor1A` | 主色浅透（未填则由主色生成 10% alpha） |
| `onPrimaryColor` | （白字） | 主色底上的文字 |
| `lightTextColor` | `lightTxtColor` | 浅色区域辅助字色 |
| `navigationBarTextColor` | `navigationBarTxtColor` | 导航栏文字 |
| `navigationBarBackgroundColor` | `navigationBackgroundColor` | 导航栏背景 |
| `disabledColor` | `disabledColor` | 禁用态 |
| `textColorPrimary` | `textColorBlack3` | 正文 |
| `textColorSecondary` | `textColorBlack6` | 次级正文 |
| `textColorTertiary` | `textColorBlack9` | 提示 / placeholder |
| `pageBackgroundColor` | 白底 | 页面背景 |
| `dividerColor` | `#D7D7D7` | 分割线 |
| `chipBackgroundColor` | `#F4F4F4` | 未选中卡片底 |
| `tabUnselectedColor` | `#242936` | 未选中 Tab |

留空字段回退到人人运维默认色板（`BrandPalette.defaults()`）。


- 开源仓库内**不存在**任何真实密钥，`demo_release.json` 的密钥字段一律为空字符串
- 真实租户配置由使用者自行提供，或从 CI Secret 注入
- 私有蓝牙 SDK 与地图 SDK 需自备，详见 [`../docs/OPEN-SOURCE.md`](../docs/OPEN-SOURCE.md)

## 能力开关

`features` 段用于在缺少闭源 SDK 时降级运行：

| 开关 | 默认 | 说明 |
|------|------|------|
| `bleTransport` | `simulator` | `simulator` 用模拟实现，`native` 使用自备的私有 BLE SDK |
| `mapProvider` | `none` | `tencent` / `google` / `simulator` / `none`。`tencent` 需配置 Key（见下）；无 Key 时自动降级为模拟地图 |
| `push` | `false` | 需自备推送 SDK 与 AppKey |
| `backgroundLocation` | `false` | 轨迹上报，开启前需确认合规与权限文案 |

## 腾讯地图

1. 在 [腾讯位置服务](https://lbs.qq.com) 创建应用，勾选 Android SDK，填写包名（demo 默认 `com.luopingtech.ebike.ops.demo`）与 SHA1。
2. 将 Key 写入 `local.properties`（不入库）：

```properties
ops.map.tencentKey=你的Key
```

3. （可选）`androidApp/src/main/assets/tenant.json` 设置 `"mapProvider": "tencent"`；若仅配置了 `ops.map.tencentKey` 且 provider 为 `none`/`tencent`，启动时会自动切到腾讯图。
4. 无 Key 时仍可编译运行，首页使用 Canvas 模拟地图。

Android 侧：`TextureMapView` + 车辆 Marker（低电橙 / 骑行绿 / 选中青），与遗留相同依赖坐标 `com.tencent.map:tencent-map-vector-sdk:4.5.12`。

