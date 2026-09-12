plugins {
    alias(libs.plugins.androidApplication) apply false
    alias(libs.plugins.androidKotlinMultiplatformLibrary) apply false
    alias(libs.plugins.kotlinMultiplatform) apply false
    alias(libs.plugins.kotlinSerialization) apply false
    alias(libs.plugins.composeCompiler) apply false
    alias(libs.plugins.jetbrainsCompose) apply false
}

/**
 * 编译各 KMP 模块的 commonMain 元数据，卡住漏进公共代码的平台专有 API。
 *
 * 这一步不能省：`assembleDebug` 和 `testAndroidHostTest` 走的都是 Android/JVM 目标，
 * `@Volatile`、`System.currentTimeMillis()`、`String.format` 这类 JVM 专有 API 在那儿
 * 一路绿灯；iOS 目标又只能在 macOS 上编译。结果就是声明了 iOS target 也没人验，
 * 曾经有 13 处泄漏这么潜伏了很久。元数据编译在任何宿主上都能跑，是唯一的防线。
 */
tasks.register("verifyCommonMain") {
    group = "verification"
    description = "Compile commonMain metadata for every KMP module"
    dependsOn(
        ":shared:compileCommonMainKotlinMetadata",
        ":sharedUi:compileCommonMainKotlinMetadata",
    )
}
