// 腾讯地图 SDK 的桥接。
//
// `QMapKit.framework` 是老式预编译 framework，里面没有 `Modules/module.modulemap`，
// 所以 Swift 的 `import QMapKit` 用不了（`canImport(QMapKit)` 恒为 false，
// 会静默跳过整个实现文件）。改走 bridging header 让 ObjC 头文件直接可见，
// 是否启用则由 `OPS_TENCENT_MAP` 编译条件控制（只在 iphoneos 下定义）。
#if __has_include(<QMapKit/QMapKit.h>)
#import <QMapKit/QMapKit.h>
#endif
