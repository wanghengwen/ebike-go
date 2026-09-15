import Foundation
import SharedUi

/// 腾讯地图 iOS SDK 的注册入口。
///
/// SDK 是闭源二进制（CocoaPods `QMapKit`）且 Key 按 bundleId 绑定，所以它只出现在宿主，
/// 跟 Android 把 `TencentMapView` 留在 `androidApp` 是同一条边界；`sharedUi` 只认
/// `IosHostMapFactory` 协议。
///
/// 未启用 SDK 时这里什么都不做，共享层自动用 MapKit —— 用编译条件而不是删代码，
/// 是为了让模拟器（没有 arm64 slice 可链）和真机两种构建都能过。
enum TencentMapHost {
    static func registerIfAvailable(config: TenantConfig) {
        let key = config.map.tencentKey.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !key.isEmpty else {
            // demo 租户没有 Key：不注册，共享层落到 MapKit。
            NSLog("[OpsApp] Tencent map key empty; using MapKit.")
            return
        }
        #if OPS_TENCENT_MAP
        let prefix = String(key.prefix(6))
        NSLog("[OpsApp] Registering QMapKit key \(prefix)… (\(key.count) chars)")
        TencentMapFactory.register(apiKey: key)
        #else
        NSLog("[OpsApp] Tencent map SDK not enabled for this destination (simulator); using MapKit. Key present (\(key.count) chars).")
        #endif
    }
}
