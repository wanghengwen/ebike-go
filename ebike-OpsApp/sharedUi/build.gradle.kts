plugins {
    alias(libs.plugins.kotlinMultiplatform)
    alias(libs.plugins.androidKotlinMultiplatformLibrary)
    alias(libs.plugins.jetbrainsCompose)
    alias(libs.plugins.composeCompiler)
}

// 跨端界面层。iOS target 声明了但在 Windows 上编译不了（Apple target 需要 macOS 宿主），
// 真正的把关是 :sharedUi:compileCommonMainKotlinMetadata —— 它在任何宿主上都会检查
// commonMain 只用了公共 API，写进 android.* 会当场编译失败。
kotlin {
    android {
        namespace = "com.luopingtech.ebike.ops.sharedui"
        compileSdk = libs.versions.android.compileSdk.get().toInt()
        minSdk = libs.versions.android.minSdk.get().toInt()
    }

    // 没有 iosX64：Compose Multiplatform 1.11 起不再发布 Intel 模拟器产物，
    // 只剩 arm64 真机与 arm64 模拟器（Apple Silicon）。shared 那边保留 iosX64 无妨，
    // 它不依赖 compose。
    listOf(
        iosArm64(),
        iosSimulatorArm64(),
    ).forEach { iosTarget ->
        iosTarget.binaries.framework {
            baseName = "SharedUi"
            isStatic = true
            // iOS 宿主只链这一个 framework：界面签名里到处是 OpsApp，
            // 不导出 :shared 的话 Swift 那边拿不到类型，还得再链一个 Shared。
            export(project(":shared"))
        }
    }

    sourceSets {
        commonMain.dependencies {
            // api 而不是 implementation：界面签名里直接出现 OpsApp，宿主要看得到。
            api(project(":shared"))
            implementation(compose.runtime)
            implementation(compose.foundation)
            implementation(compose.material3)
            implementation(compose.ui)
            implementation(libs.kotlinx.coroutines.core)
        }
        androidMain.dependencies {
            // BackHandler：H5 的多页返回要拦系统返回键。CMP 的公共 BackHandler 在 1.10
            // 起已废弃、要换 navigationevent，而返回语义本来就分平台，索性各自实现。
            implementation(libs.androidx.activity.compose)
        }
    }
}

// 见根项目 verifyCommonMain：不接进 check，平台专有 API 就没人拦。
tasks.named("check").configure {
    dependsOn("compileCommonMainKotlinMetadata")
}
