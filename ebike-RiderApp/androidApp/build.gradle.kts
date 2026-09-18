import groovy.json.JsonSlurper
import java.util.Properties

plugins {
    alias(libs.plugins.androidApplication)
    // 不要加 kotlin-android：AGP 9 内置 Kotlin 支持，同时应用会直接报配置错误。
    alias(libs.plugins.composeCompiler)
}

val localProps = Properties().apply {
    val file = rootProject.file("local.properties")
    if (file.exists()) {
        file.inputStream().use { load(it) }
    }
}

fun prop(name: String, default: String): String {
    // local.properties > gradle.properties / -P > default
    // IDE Sync 有时读不全 local.properties，故 gradle.properties 也写一份租户。
    return localProps.getProperty(name)?.takeIf { it.isNotBlank() }
        ?: (findProperty(name) as? String)?.takeIf { it.isNotBlank() }
        ?: default
}

val riderTenant: String = prop("rider.tenant", "renren")
val riderMode: String = prop("rider.mode", "release")
val tenantConfigFile = rootProject.file("config/${riderTenant}_${riderMode}.json")
val tenantConfig: Map<*, *>? = if (tenantConfigFile.exists()) {
    @Suppress("UNCHECKED_CAST")
    JsonSlurper().parse(tenantConfigFile) as Map<*, *>
} else {
    null
}

fun tenantString(vararg path: String, default: String = ""): String {
    var cur: Any? = tenantConfig
    for (key in path) {
        cur = (cur as? Map<*, *>)?.get(key) ?: return default
    }
    return cur?.toString()?.takeIf { it.isNotBlank() } ?: default
}

// 包名跟着租户走：同一份代码给不同运营方出不同的包，不能写死在 gradle 里。
val configApplicationId: String =
    tenantString("app", "androidApplicationId", default = "com.luopingtech.ebike.rider.renren")
val configDisplayName: String = tenantString("app", "displayName", default = "Rider")
val configTencentKey: String = tenantString("map", "tencentKey")
val tencentMapKey: String = prop("rider.map.tencentKey", configTencentKey)
    .ifBlank { configTencentKey }
val configWechatAppId: String = tenantString("pay", "wechatAppId")
val wechatAppId: String = prop("rider.pay.wechatAppId", configWechatAppId)
    .ifBlank { configWechatAppId }

android {
    // 包名固定，applicationId 才是随租户变的那个。
    namespace = "com.luopingtech.ebike.rider"
    compileSdk = libs.versions.android.compileSdk.get().toInt()

    defaultConfig {
        applicationId = configApplicationId
        minSdk = libs.versions.android.minSdk.get().toInt()
        targetSdk = libs.versions.android.targetSdk.get().toInt()
        versionCode = 1
        versionName = "1.0.0"
        buildConfigField("String", "RIDER_TENANT", "\"${riderTenant.replace("\"", "\\\"")}\"")
        buildConfigField("String", "RIDER_MODE", "\"${riderMode.replace("\"", "\\\"")}\"")
        buildConfigField("String", "RIDER_DISPLAY_NAME", "\"${configDisplayName.replace("\"", "\\\"")}\"")
        // 地图 SDK 还没接，Key 先透进来，接的时候不用再改构建脚本。
        manifestPlaceholders["TENCENT_MAP_KEY"] = tencentMapKey
        buildConfigField("String", "TENCENT_MAP_KEY", "\"${tencentMapKey.replace("\"", "\\\"")}\"")
        buildConfigField("String", "WECHAT_APP_ID", "\"${wechatAppId.replace("\"", "\\\"")}\"")
        // 真机只跑 ARM；地图 / 蓝牙类 AAR 自带 x86 slice，不滤掉会白白进包。
        ndk {
            abiFilters += listOf("armeabi-v7a", "arm64-v8a")
        }
    }

    buildFeatures {
        compose = true
        buildConfig = true
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    packaging {
        resources {
            excludes += "/META-INF/{AL2.0,LGPL2.1}"
        }
    }
}

/**
 * 运行时读的是 `assets/tenant.json`，仓库里维护的是 `config/{tenant}_{mode}.json`。
 * 挂在 preBuild 上，assets 那份就永远是构建当时的配置，不会出现改了 config 忘了拷的漂移。
 * iOS 侧的对应物是 `iosApp/sync_tenant.sh`，两边同源同字节。
 */
val syncTenantConfig by tasks.registering {
    group = "rider"
    description = "Copy config/{tenant}_{mode}.json → androidApp/src/main/assets/tenant.json"
    inputs.file(tenantConfigFile).optional()
    outputs.file(layout.projectDirectory.file("src/main/assets/tenant.json"))
    doLast {
        check(tenantConfigFile.exists()) {
            "Missing tenant config: ${tenantConfigFile.path}. " +
                "Set rider.tenant / rider.mode in local.properties."
        }
        val outDir = layout.projectDirectory.dir("src/main/assets").asFile
        outDir.mkdirs()
        tenantConfigFile.copyTo(outDir.resolve("tenant.json"), overwrite = true)
        logger.lifecycle("Synced ${tenantConfigFile.name} → assets/tenant.json (appId=$configApplicationId)")
    }
}

tasks.named("preBuild").configure {
    dependsOn(syncTenantConfig)
}

tasks.register("showApplicationId") {
    group = "rider"
    description = "Print the tenant applicationId used by :androidApp"
    doLast {
        logger.lifecycle("rider.tenant=$riderTenant rider.mode=$riderMode applicationId=$configApplicationId")
    }
}

dependencies {
    implementation(project(":shared"))
    implementation(project(":sharedUi"))
    implementation(libs.androidx.activity.compose)
    implementation(platform(libs.androidx.compose.bom))
    implementation(libs.androidx.compose.ui)
    implementation(libs.androidx.compose.ui.tooling.preview)
    implementation(libs.androidx.compose.material3)
    implementation(libs.androidx.compose.foundation)
    implementation(libs.kotlinx.coroutines.android)
    implementation(libs.androidx.camera.camera2)
    implementation(libs.androidx.camera.lifecycle)
    implementation(libs.androidx.camera.view)
    implementation(libs.mlkit.barcode.scanning)
    implementation(libs.tencent.map.vector.sdk)
    implementation(libs.tencent.map.sdk.utilities)
    implementation(libs.wechat.sdk)
    debugImplementation(libs.androidx.compose.ui.tooling)
}
