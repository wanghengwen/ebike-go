import SwiftUI
import SharedUi

/// The whole UI is the shared Compose Multiplatform tree from `:sharedUi`.
/// Swift only bridges a UIViewController - no screen is re-implemented in SwiftUI.
struct ContentView: UIViewControllerRepresentable {
    let ops: OpsApp

    func makeUIViewController(context: Context) -> UIViewController {
        OpsAppViewControllerKt.OpsAppViewController(app: ops)
    }

    func updateUIViewController(_ uiViewController: UIViewController, context: Context) {}
}
