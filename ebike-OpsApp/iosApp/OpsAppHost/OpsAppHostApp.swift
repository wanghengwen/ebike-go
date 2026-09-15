import SwiftUI
import SharedUi

@main
struct OpsAppHostApp: App {
    private let ops: OpsApp

    init() {
        // 跟 Android 一样读 bundle 里的 tenant.json（由 iosApp/sync_tenant.sh
        // 从 config/{tenant}_{mode}.json 同步，与 local.properties 的 ops.tenant 对齐）。
        // 地图 Key / Bundle ID 也走同一份配置；没有资源文件时 createBundleIosOpsApp
        // 内部会退回 demo。
        ops = OpsAppIosKt.createBundleIosOpsApp()

        // 腾讯地图（闭源 SDK + 按 bundleId 绑定的 Key）如果接进来了，在这里注册；
        // 没注册时共享层自动用 MapKit，见 TencentMapHost.swift。
        // 模拟器构建没有 OPS_TENCENT_MAP，这里会打日志后跳过。
        TencentMapHost.registerIfAvailable(config: ops.config)
    }

    var body: some Scene {
        WindowGroup {
            ContentView(ops: ops)
                .ignoresSafeArea()
        }
    }
}
