import SwiftUI
import SharedUi

/// 整个界面就是 `:sharedUi` 里那棵 Compose Multiplatform 树。
/// Swift 只桥一个 UIViewController —— 没有任何一屏在 SwiftUI 里重写。
struct ContentView: UIViewControllerRepresentable {
    let rider: RiderApp

    func makeUIViewController(context: Context) -> UIViewController {
        RiderAppViewControllerKt.RiderAppViewController(app: rider)
    }

    func updateUIViewController(_ uiViewController: UIViewController, context: Context) {}
}
