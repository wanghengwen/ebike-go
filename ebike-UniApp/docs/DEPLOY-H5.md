# UniApp H5 部署清单（Rider WebView / 浏览器）

## 是否就绪

| 项 | 状态 |
| --- | --- |
| `nativeHost` 运行时桥（小程序不受影响） | ✅ |
| `request` / 登录跳转 / 支付占位 / 语言同步 | ✅ |
| `uploadFile` 在 native 走原生选图上传 | ✅ |
| `hash` + `base: ./`（可挂任意 CDN 子目录） | ✅ |
| `npm run build:h5:renren` | ✅ |
| 产物 `dist/build/h5/`（相对路径资源） | ✅ |

## 构建（人人骑行）

```bash
cd ebike-UniApp
npm ci   # 首次
npm run build:h5:renren
```

产物目录：`dist/build/h5/`

- `index.html`
- `assets/*`（JS/CSS，相对路径 `./assets/...`）
- `static/*`

## 部署

把 **整个** `dist/build/h5/` 上传到静态站点，例如：

```text
https://你的域名/h5/renren/index.html
```

要求：

1. **HTTPS**（App WebView 与混合内容）
2. 支持 SPA：直接打开带 `#/pages-sub/...` 的 URL 应回到同一 `index.html`（hash 模式通常无需 rewrite）
3. 勿单独只传 `index.html`，必须带上 `assets/`、`static/`

CORS：嵌在 Rider 里时接口走原生桥，**不依赖**浏览器跨域打 `client.luopingtech.com`。  
若用浏览器直接打开 H5（非 App），仍走 H5 自签请求，需按原 UniApp 方式部署。

## 接到 RiderApp

编辑本机（不入库）`ebike-RiderApp/config/renren_release.json`：

```json
"h5": {
  "baseUrl": "https://你的域名/h5/renren/index.html"
}
```

然后：

```bash
cd ebike-RiderApp
# local.properties 已是 rider.tenant=renren
.\gradlew.bat :androidApp:assembleDebug
```

App 内「我的钱包 / 帮助」会打开：

```text
{h5.baseUrl}#/pages-sub/pay/wallet/wallet?lang=...&tenantId=2&themeColor=...
```

origin 白名单只认 `h5.baseUrl` 的 scheme+host，换域名必须同步改租户配置。

## 安全说明

- renren 构建会把 `platformSecret` / `platformSign` 打进 `merged.runtime.json`（与历史小程序/H5 一致）。
- **在 Rider WebView 内**，业务请求走 `nativeHost.request`，密钥不下发到桥；但静态 JS 里仍可能含配置，CDN 访问控制按你们现网 H5 策略执行。
- 支付在 App 内仍返回 `UNSUPPORTED`（P5 未做）；钱包页可看余额，渠道支付请用小程序或等 P5。

## 本地预览（可选）

```bash
npx serve dist/build/h5 -p 4173
# 浏览器打开 http://127.0.0.1:4173/
# Rider 真机调试可把 h5.baseUrl 临时设为 http://电脑局域网IP:4173/
```

Android 清文本 HTTP 需在 Manifest 放行（仅调试）；正式环境用 HTTPS。
