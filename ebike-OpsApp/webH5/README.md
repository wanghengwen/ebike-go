# webH5

OpsApp 的**运营大屏 / 营收大屏** H5，由 App 内 WebView（`androidApp/.../ui/h5/H5Screen.kt`）或浏览器直接打开。

从遗留的 `APP/Merchant-H5`（Vue 2.6 + Vant 2 + ECharts 4 + vue-cli）重写而来，
技术栈与 `ebike-UniApp` 对齐：Vue 3 + TypeScript + Vite + Pinia。

本目录是独立的 npm 工程，**不参与 Gradle 构建**，`settings.gradle.kts` 里没有它。

---

## 与 App 之间的契约

App 侧由 `shared/.../core/config/H5ScreenUrls.kt` 的 `resolve()` 拼出 URL，
参数经 hash 路由后追加。**改动下表任何一项都必须同步改 Kotlin 侧。**

| 参数 | 用途 |
|------|------|
| `apiHost` | 网关基址 |
| `accessToken` / `refreshToken` | 会话；过期时 H5 自己走 `/oauth/token` 刷新 |
| `tenantId` / `tenantSecret` | 刷新 token 时的 Basic 认证 |
| `sign` | 请求签名密钥，参与 `_s` 头计算 |
| `platform` / `deviceId` / `opMan` | 埋点与审计字段 |
| `themeColor` / `lightTxtColor` | 主题色，写入 CSS 变量 `--ops-theme-color` |
| `lang` | 可选，`zh-CN` / `en`；不传则按浏览器语言 |

路由为 hash 模式，两块大屏的路径沿用遗留值：

- 运营大屏 `#/newOperationScreen`（权限码 `0221`）
- 营收大屏 `#/newRevenueHome`（权限码 `LargeRevenueScreen`）

租户配置里对应 `h5.operationUrl` / `h5.revenueUrl`，见 [`../config/README.md`](../config/README.md)。

> **已知风险**：`sign` 与 `tenantSecret` 目前以明文出现在 URL query 中，会进入 WebView 历史与日志。
> 换成 JS Bridge 注入需要同步改 `H5Screen.kt`，留作后续独立一期。

---

## 开发

```bash
cd webH5
npm install
cp .env.example .env   # 填本地联调用的 apiHost / tenantId / sign
npm run dev            # http://localhost:5180/mop-saas/
npm run typecheck
npm run build          # 产物在 dist/
```

Node 需要 22.12 以上。

`.env` 里的 `VITE_DEV_*` 只在 `import.meta.env.DEV` 下作为兜底生效，
生产一律以 URL query 为准。**不要把真实密钥提交进仓库。**

---

## 目录

```text
src/
  api/          网关路径表、请求签名与 token 刷新、出入参类型
  components/   跨页复用：BaseChart / RevenueChart / MultiSelectPanel / CountUpValue
  composables/  useEcharts（实例生命周期）、useDateRange、echartsCore（按需注册）
  i18n/         vue-i18n 词条，zh-CN / en
  pages/
    operation-screen/  运营大屏：实时数据页签 + 查询数据页签
    revenue-home/      营收大屏：数据页签 + 图表页签
    trend/             趋势折线详情页
  router/       hash 路由 + 配置注入 / 主题 / 权限的全局守卫
  stores/       config（运行时配置）、permission、serviceArea、trend
  utils/        storage / theme / numFormat / carType / deviceFilters
```

---

## 与遗留实现的行为差异

移植过程中保留了全部业务口径，但修掉了几处确定的缺陷：

| 位置 | 遗留行为 | 现在 |
|------|---------|------|
| `common/axios.js` | 给全局 axios 实例挂拦截器 | 改用 `axios.create()` 独立实例 |
| 同上 | 401 后把请求推进 `requestArr` 但从不消费（重放代码被注释），靠 `window.location.href` 整页刷新兜底 | 单飞刷新 + 真正的队列重放，每个请求最多重试一次 |
| 同上 | 模块加载时读一次 `storage.get('config')`，后续 `setConfig` 不会更新这份快照 | 统一从 Pinia store 读 |
| 各图表组件 | `mounted` 和 `updated` 里都调 `echarts.init`，实例只创建不销毁 | `useEcharts` 只 init 一次，卸载时 dispose，并接 `ResizeObserver` |
| 运营大屏「趋势折线图」 | `$router.push` 到 `/orderAmount/${JSON}` 等**未注册的路由**，点击静默失败 | 新增 `/trend` 详情页，数据走 store 传递而非塞进 URL |
| 营收大屏「查看详情」 | 同样跳未注册路由（`/carsDetail`、`/userStateTotal` 等），16 个入口全部静默失败 | 改成底部弹层直接展示明细行 |
| `xcStorage.js` | `JSON.parse` 未捕获异常 | 解析失败回落 `null` |
| `NumFormat` | `NaN` 渲染成 `"NaN.undefined"`、`-Infinity` 渲染成 `"-Infinity.undefined"` | 非有限数统一渲染 `-- --` |
| 响应码 `00006` / `00015` | 命中白名单后 promise 永不 resolve，调用方连同 loading 一起挂死 | 正常返回，由调用方按 `success:false` 处理 |
| `index.html` | 引用阿里 CDN 图标字体，只为两个下箭头 | 换成 Vant 内置 `van-icon`，去掉外部依赖 |
| `accept-language` | 写死 `zh-CN` | 跟随 `lang` 参数，与 App 的语言设置联动 |
| 「数据统计」折叠面板 | `activeNames: [1]`（数字）配 `name="1"`（字符串），比较不相等，实际默认全折叠 | 保持默认折叠的观感 |
| 营收大屏 14 个图表 | grid / legend 尺寸在 20%~27%、10%~15%、18×8 与 12×12 之间各不相同（复制粘贴漂移） | 统一取出现次数最多的一组；配色仍保留原有的两套 |

以下遗留内容确认为死代码，未迁入：

- `xcAgent.vue` / `xcDeleteService.vue`：加盟商选择弹窗引用了页面上不存在的 `showAgent` / `agentData` 等变量，
  提交的 `setAgentId` mutation 在 Vuex store 里也没有定义
- `orderWater.vue` / `reportData.vue` / `newOperationScreenChart/trendChart.vue` /
  `operationScreenChart/` 下的 `orderLineChart`、`orderNumChart`、`orderPieChart`：从未被任何模板引用
- `v-hasCode` 指令：注册了但零处使用；权限判断改由 `usePermissionStore().hasExpression()` 承担
- `common/api.js` 中约 40 个标注「未调用」的 `/mieba/v2/big_screen/*` 路径
- 营收大屏图表页签的「坏账趋势」（`currentIndex == 0`）：入口按钮被注释掉，配套的年月选择器、
  `arrearsChart.vue` 和 `/mieba/.../arrears_proportion_new` 接口一并不可达
- `getActivityRidingTimes`：路径与 `getActivityRidingCard` 完全相同，只是多读一个字段，
  合并成一次请求
- `ridingCardDateMap`（骑行卡类型码 → 天数）：挂在 `data()` 上但没有任何消费方，
  售卖次数图的图例名是接口直接下发的

### 已知未覆盖

- 赠送统计里「会员卡赠送 / 骑行卡赠送 / 优惠卡赠送」三项没有明细弹层：
  遗留的对应入口本来就跳不到页面，且 `getActivityDepositTimes` 的明细数据当前没有取
- 主题色：`H5ScreenUrls.kt` 的 `resolve()` 没有下发 `themeColor` / `lightTxtColor`，
  两版都只会用到默认的 `#FF9C80`；需要跟随租户主题时要在 Kotlin 侧补参数

---

## 打包体积

ECharts 走按需注册（`composables/echartsCore.ts`），只装了 bar / line / pie / tree 四类图与
tooltip / legend / grid 三个组件。新增图表类型时必须在那里补 `use()`，否则运行期静默不渲染。
