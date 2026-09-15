# iosApp

SwiftUI host that runs the shared Compose Multiplatform UI from `:sharedUi`.
Swift code here is 30 lines: it links `SharedUi.framework`, builds `OpsApp`, and
hands a `UIViewController` to SwiftUI. Every screen comes from `commonMain`.

## Status (2026-09-14)

| Piece | State | Blocker |
|-------|--------|---------|
| Shared UI entry (`OpsAppViewController`) | Written in `sharedUi/iosMain` | Needs a Mac to compile/link |
| Login / service area / 4 tabs / scan shell | Shared (`ui/shell`, commonMain) | — |
| `SharedUi.framework` link | Build **on macOS** via Gradle | Windows cannot link Apple targets |
| Camera `OpsScanPreview` | Android wired | iOS AVFoundation TBD (falls back to a visible notice) |
| Map `OpsMapRenderer` | Tencent on Android | iOS SDK + keys TBD (falls back to `SimulatorMapRenderer`) |
| `PlatformWebView` (H5 screens) | Android WebView | iOS `WKWebView` TBD |
| Toast `LocalOpsToast` | Android Toast | iOS HUD TBD (no-op default) |
| Keychain `SecureStore` | In-memory on iOS (`createSecureStore`) | Keychain helper TBD |
| Location / BLE | Simulator / unsupported | CoreLocation + the same closed-source SDK wall as Android |

Everything in the "iOS TBD" rows has a contract in `commonMain` with a default
implementation, so the app composes and navigates without them — an unwired
capability shows a placeholder instead of crashing.

## What is intentionally not claimed

This has never been compiled: Apple targets need a macOS host. The shared UI is
verified by `:sharedUi:compileCommonMainKotlinMetadata` (which rejects any
platform-specific API in `commonMain`), not by an iOS build.

## Build (macOS)

```bash
# 1) Framework (pick the target that matches your run destination)
./gradlew :sharedUi:linkDebugFrameworkIosSimulatorArm64
# device: ./gradlew :sharedUi:linkDebugFrameworkIosArm64

# 2) Xcode project
cd iosApp
brew install xcodegen   # once
xcodegen generate
open OpsAppHost.xcodeproj
```

`SharedUi` exports `:shared`, so `import SharedUi` also brings in `OpsApp` and
the feature/domain types; there is no second framework to link. If device and
simulator paths differ, adjust `FRAMEWORK_SEARCH_PATHS` in `project.yml` or move
to an XCFramework in CI.

## Host roadmap (when a Mac is available)

1. Build once, fix whatever Kotlin/Native interop the Windows box could not check
2. `AVFoundation` scanner → provide `LocalOpsScanPreview` (unlocks scan + field ops)
3. Map renderer → provide `LocalOpsMapRenderer` (unlocks the map tab and 4 map screens)
4. `WKWebView` → `PlatformWebView` actual (unlocks every H5 management screen)
5. Keychain `SecureStore`, CoreLocation tracker, HUD for `LocalOpsToast`
6. Inject a real `BleTransport` only from a private module (see `docs/BLE.md`)

Do not commit private BLE frameworks or map API keys here.
