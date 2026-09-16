# iosApp — RiderAppHost

iOS 宿主。界面一行都不在这里：整屏是 `:sharedUi` 的 Compose 树，Swift 只桥一个
`UIViewController`（见 `ContentView.swift`）。对应的 Android 宿主是
`androidApp/src/main/kotlin/com/luopingtech/ebike/rider/MainActivity.kt`。

## 入库内容

| 路径 | 说明 |
|------|------|
| `RiderAppHost.xcodeproj` | Xcode 工程（已入库，可直接打开） |
| `RiderAppHost.xcworkspace` | CocoaPods workspace（**请打开这个**） |
| `Podfile` / `Podfile.lock` | QMapKit 依赖锁定 |
| `project.yml` | XcodeGen 源；结构大改时再 `xcodegen generate` |
| `RiderAppHost/*.swift` 等 | 宿主源码 |

**不入库**：`Pods/`（本地 `pod install`）、`xcuserdata/`、`tenant.json`（由 `sync_tenant.sh` 生成）。

## 在 Mac 上开发（日常）

```bash
# 一次性：Homebrew
brew install cocoapods          # xcodegen 仅在要重生成工程时需要

# 首次 / 换租户 / 换模拟器↔真机 时
cd /path/to/ebike-RiderApp
# local.properties 里设 rider.tenant / rider.mode（与 Android 共用）
bash iosApp/bootstrap.sh sim      # 或 device

open iosApp/RiderAppHost.xcworkspace
```

`bootstrap.sh` 会：同步租户 → 链接 `SharedUi.framework` → `pod install`。

手动等价步骤：

```bash
bash iosApp/sync_tenant.sh
./gradlew :sharedUi:linkDebugFrameworkIosSimulatorArm64   # 真机用 IosArm64
cd iosApp && pod install
open RiderAppHost.xcworkspace
```

已接腾讯地图（QMapKit）：**真机**走腾讯地图，模拟器无 SDK arm64 slice，自动降级 MapKit。

## 何时再跑 xcodegen

改了 `project.yml`（Bundle ID 模板、编译条件、Framework 搜索路径等）之后：

```bash
cd iosApp
xcodegen generate
pod install
# 把更新后的 .xcodeproj / .xcworkspace 一并提交
```

日常只改 Swift / Kotlin 不必重跑。

## Windows ↔ Mac Mini

Windows 同步源码：`../../_sync_build_ios_rider.ps1`  
Mac 整包 Release：`_build_ios_release_rider.sh [sim|device]`

## 没有 iosX64

Compose Multiplatform 1.11 起不再发布 Intel 模拟器产物，只剩 arm64 真机与 arm64
模拟器（Apple Silicon）。`:shared` 那边保留了 `iosX64`，它不依赖 compose。

P6 H5：`WKScriptMessageHandler` 名 `riderNative`，随 `SharedUi` framework 一起 link。
本地 http 调试页需在 Info.plist 放行 ATS。租户 `h5.baseUrl` 见 `docs/README.md`。
