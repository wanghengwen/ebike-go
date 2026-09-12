# iosApp

SwiftUI host that consumes KMP `Shared.framework` from `:shared`.

## Status (2026-09-11)

| Piece | State | Blocker |
|-------|--------|---------|
| SwiftUI scaffold (`OpsAppHost`) | Ready (source + XcodeGen spec) | — |
| `Shared.framework` link | Build **on macOS** via Gradle | Windows CI cannot link iOS frameworks |
| Login / task / map shell parity | Not started | Needs Mac + SKIE (or suspend wrappers) + Compose-equivalent SwiftUI |
| Camera `CodeScanner` | Android wired | iOS AVFoundation TBD |
| Keychain `SecureStore` | In-memory on iOS (`createSecureStore`) | Keychain helper TBD |
| Location / map | Simulator / unsupported | CoreLocation + map SDK keys |
| BLE | `SimulatorBleTransport` / `UnavailableBleTransport` | Same closed-source SDK wall as Android |

## What is intentionally not claimed

Full field-ops parity with Android Compose is **blocked on a Mac build machine** and private BLE/map binaries. This folder proves linkage and documents the remaining host work; it is not a shipping iOS client.

## Build (macOS)

```bash
# 1) Framework (pick the target that matches your run destination)
./gradlew :shared:linkDebugFrameworkIosSimulatorArm64
# device: ./gradlew :shared:linkDebugFrameworkIosArm64

# 2) Xcode project
cd iosApp
brew install xcodegen   # once
xcodegen generate
open OpsAppHost.xcodeproj
```

If the framework path for device vs simulator differs, adjust `FRAMEWORK_SEARCH_PATHS`
in `project.yml` or use a fat/XCFramework script later in CI.

## Host roadmap (when Mac is available)

1. SKIE or thin completion wrappers for `StateFlow` / suspend features
2. SwiftUI shell mirroring Android tabs: Map · Tasks · Scan · Workbench
3. Wire `AuthFeature` → service area → task features (same shared APIs as Android)
4. `AVFoundation` scanner + Keychain `SecureStore` + CoreLocation
5. Inject real `BleTransport` only from a private module (see `docs/BLE.md`)

## What the host proves today

- `OpsApp.demo()` starts
- Config / version / demo-mode labels render from Shared

Do not commit private BLE frameworks or map API keys here.
