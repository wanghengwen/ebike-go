// 闭源 ObjC SDK 的桥接位。
//
// `QMapKit.framework` 这类老式预编译 framework 里没有 `Modules/module.modulemap`，
// Swift 的 `import QMapKit` 用不了（`canImport(QMapKit)` 恒为 false，会静默跳过整个
// 实现文件）。走 bridging header 才能让 ObjC 头文件直接可见。
//
// `__has_include` 守卫：模拟器构建不链 QMapKit 时也能编过。
#if __has_include(<QMapKit/QMapKit.h>)
#import <QMapKit/QMapKit.h>
#endif
