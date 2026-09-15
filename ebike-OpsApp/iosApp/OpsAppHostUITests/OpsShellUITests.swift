import XCTest

/// 走一遍主壳的五个入口并把每屏截图挂到测试结果里。
///
/// Compose 在 iOS 上渲染成单个 `UIView`，XCUITest 拿不到控件树，所以这里全部用
/// 归一化坐标点击 —— 换句话说它测的是「点下去不崩、该出的界面出来了」，
/// 具体控件的行为归共享层的单元测试管。
final class OpsShellUITests: XCTestCase {

    override func setUp() {
        super.setUp()
        continueAfterFailure = false
    }

    func testMainShellTabs() {
        let app = XCUIApplication()
        app.launch()
        XCTAssertTrue(app.wait(for: .runningForeground, timeout: 20), "app did not reach foreground")
        // 首帧之后共享层还要装配 feature / 贴地图视野，给它一点时间再截图。
        sleep(6)
        attach(app, name: "01-map")

        // 底部 tab：地图 / 任务 / 扫码 / 数据分析 / 工作台。
        let tabs: [(String, CGFloat)] = [
            ("02-task", 0.30),
            ("03-analysis", 0.70),
            ("04-workbench", 0.90),
            ("05-map-back", 0.10),
        ]
        for (name, x) in tabs {
            tap(app, x: x, y: TAB_BAR_Y)
            sleep(4)
            attach(app, name: name)
            XCTAssertEqual(app.state, .runningForeground, "app left foreground after \(name)")
        }

        // 中间的扫码按钮会拉起相机权限弹窗，单独放最后：它一旦弹出就挡住后面的点击。
        tap(app, x: 0.50, y: TAB_BAR_Y)
        sleep(4)
        attach(app, name: "06-scan")
        XCTAssertEqual(app.state, .runningForeground, "app left foreground on scan")
    }

    /// 工作台里的「运营数据大屏」是 H5（WKWebView）入口，要先滚到运营模块。
    func testWorkbenchScrollToH5() {
        let app = XCUIApplication()
        app.launch()
        XCTAssertTrue(app.wait(for: .runningForeground, timeout: 20), "app did not reach foreground")
        sleep(6)
        tap(app, x: 0.90, y: TAB_BAR_Y)
        sleep(3)
        // swipeUp 带惯性，每次滚多少不一定；慢速定距拖拽才能复现同一个位置。
        dragUp(app, from: 0.80, to: 0.30)
        sleep(2)
        attach(app, name: "10-workbench-scrolled")

        // 运营模块第一行第三个「运营数据」= H5 大屏，走 WKWebView。
        tap(app, x: 0.613, y: 0.440)
        sleep(12)
        attach(app, name: "11-h5-operation")
        XCTAssertEqual(app.state, .runningForeground, "app left foreground opening H5")
    }

    private func tap(_ app: XCUIApplication, x: CGFloat, y: CGFloat) {
        app.coordinate(withNormalizedOffset: CGVector(dx: x, dy: y)).tap()
    }

    private func dragUp(_ app: XCUIApplication, from: CGFloat, to: CGFloat) {
        let start = app.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: from))
        let end = app.coordinate(withNormalizedOffset: CGVector(dx: 0.5, dy: to))
        start.press(
            forDuration: 0.2,
            thenDragTo: end,
            withVelocity: .slow,
            thenHoldForDuration: 0.3
        )
    }

    private func attach(_ app: XCUIApplication, name: String) {
        let shot = XCTAttachment(screenshot: XCUIScreen.main.screenshot())
        shot.name = name
        shot.lifetime = .keepAlways
        add(shot)
    }
}

private let TAB_BAR_Y: CGFloat = 0.945
