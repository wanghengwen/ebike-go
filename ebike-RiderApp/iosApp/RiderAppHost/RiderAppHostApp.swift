import SwiftUI
import SharedUi

@main
struct RiderAppHostApp: App {
    private let rider: RiderApp

    init() {
        // 跟 Android 一样读 bundle 里的 tenant.json（由 iosApp/sync_tenant.sh 从
        // config/{tenant}_{mode}.json 同步，与 local.properties 的 rider.tenant 对齐）。
        // 没有资源文件时 createBundleIosRiderApp 内部会退回 demo。
        rider = RiderAppIosKt.createBundleIosRiderApp()

        // 腾讯地图（闭源 SDK + 按 bundleId 绑定的 Key）如果接进来了，在这里注册；
        // 没注册时共享层自动用 MapKit。模拟器构建没有 RIDER_TENCENT_MAP，会打日志后跳过。
        TencentMapHost.registerIfAvailable(config: rider.config)
    }

    var body: some Scene {
        WindowGroup {
            ContentView(rider: rider)
                .ignoresSafeArea()
        }
    }
}
