import SwiftUI
import Shared

@main
struct OpsAppHostApp: App {
    private let ops: OpsApp

    init() {
        // Demo tenant (blank api.baseUrl). Replace with JSON-loaded TenantConfig for remote.
        ops = OpsApp.companion.demo()
    }

    var body: some Scene {
        WindowGroup {
            ContentView(ops: ops)
                .onAppear {
                    // Keep a reference for the app lifetime; close() on terminate if needed.
                }
        }
    }
}
