import SwiftUI
import Shared

/// Minimal host proving Shared.framework linkage.
/// Full Flow observation / suspend bridging lands with SKIE (or thin wrappers) on a Mac.
struct ContentView: View {
    let ops: OpsApp

    var body: some View {
        NavigationStack {
            Form {
                Section("Shared") {
                    Text(ops.config.app.displayName.isEmpty ? "Ops Demo" : ops.config.app.displayName)
                    Text("version \(OpsApp.companion.LIBRARY_VERSION)")
                    Text(ops.isDemoMode ? "mode: demo" : "mode: remote")
                    Text("platform \(ops.deviceInfo.platform) \(ops.deviceInfo.osVersion)")
                }
                Section("Next on Mac") {
                    Text("1. ./gradlew :shared:linkDebugFrameworkIosSimulatorArm64")
                    Text("2. xcodegen generate && open OpsAppHost.xcodeproj")
                    Text("3. SKIE / wrappers → AuthFeature, task features, tabs")
                    Text("4. AVFoundation scanner · Keychain · CoreLocation")
                    Text("Blocked: full Android shell parity without Mac + private BLE SDK")
                }
            }
            .navigationTitle("OpsAppHost")
        }
    }
}
