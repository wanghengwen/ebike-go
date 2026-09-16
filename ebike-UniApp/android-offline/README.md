# Android 离线打包（Android Studio）

本工程 CLI 版本 `3.0.0-5020420260813003` 对应 **HBuilderX 5.24.2026081301**。离线 SDK 必须与此版本一致。

## 一次性准备

### 1. 下载并解压 SDK

- 下载页：https://nativesupport.dcloud.net.cn/AppDocs/download/android.html  
- 百度网盘（提取码 `jrrb`）：https://pan.baidu.com/s/1AFjLggD7g6ue0iKgZ8yVyA?pwd=jrrb  
- 和彩云：https://yun.139.com/shareweb/#/w/i/2w2KLgoTHEPz0  

解压后，把其中的 `HBuilder-Integrate-AS` 目录放到：

```text
ebike-UniApp/android-offline/HBuilder-Integrate-AS/
```

（同级通常还有 `SDK/`、`HBuilder-Hello/`、`Readme.txt`，保留 `SDK` 与 Integrate 工程即可。）

### 2. 申请离线打包 AppKey

1. 打开 https://dev.dcloud.net.cn/  
2. 创建应用，appid 与 `src/manifest.json` 中一致（当前占位 `__UNI__LUOPINGTECH`，正式环境请换成 DCloud 正式 appid）  
3. 配置 Android 包名 `com.luopingtech.ebike.demo`、签名证书 SHA1  
4. 将得到的 `dcloud_appkey` 填入  
   `HBuilder-Integrate-AS/simpleDemo/src/main/AndroidManifest.xml` 的 meta-data  

### 3. Android Studio

已安装路径：

`C:\Program Files\Android\Android Studio\bin\studio64.exe`

## 日常同步与调试

在 `ebike-UniApp` 目录：

```bash
# 编译 App 资源并同步到 Integrate 工程
npm run android:sync -- --build --env=demo --mode=release

# 若已编译过，仅同步
npm run android:sync
```

然后：

1. 启动 Android Studio，Open → `android-offline/HBuilder-Integrate-AS`  
2. 等待 Gradle Sync  
3. 选中 `simpleDemo`，连接模拟器或真机 Run  

修改前端代码后再次执行 `npm run android:sync -- --build`，再在 Studio 里重新 Run（或只覆盖 assets 后重启 App）。

## 包名与模块

| 项 | 值 |
|----|----|
| 默认包名 | `com.luopingtech.ebike.demo`（见租户 `appPlus.androidPackage`） |
| 地图 | App 使用高德；在 `local.secrets` / 租户配置填 `appPlus.amapKey`，并按 SDK 文档接入高德 AAR |
| 支付 | App 渠道收银台尚未接后端；行程结清走钱包 `deductWallet` |

`simpleDemo/build.gradle` 的 `applicationId` 需与包名、DCloud AppKey 一致。

## 多端说明

同一套业务代码：

| 端 | 登录 | 渠道支付 | 行程结清 |
|----|------|----------|----------|
| 微信小程序 | 一键登录 `quick-login` | `WXLITE` / `BAOFU_WXLITE` | 余额够 → 钱包；不够 → 去充值 |
| 支付宝小程序 | 手机号登录 | 既有通道（条件编译） | 同上 |
| Android / iOS | 手机号登录 | 暂不支持在线充值 | 余额够 → 钱包 |

小程序逻辑通过 `#ifdef MP-WEIXIN` / `#ifndef APP-PLUS` 隔离，App 改动不会改掉微信一键登录与 JSAPI 支付。
