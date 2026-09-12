import java.util.Properties
import groovy.json.JsonSlurper

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

val opsTenant: String = localProps.getProperty("ops.tenant", "demo").orEmpty()
val opsMode: String = localProps.getProperty("ops.mode", "release").orEmpty()
val tenantConfigFile = rootProject.file("config/${opsTenant}_${opsMode}.json")
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

val configApplicationId: String = tenantString("app", "androidApplicationId", default = "com.luopingtech.ebike.ops.demo")
val configDisplayName: String = tenantString("app", "displayName", default = "Demo Ops")
val configTencentKey: String = tenantString("map", "tencentKey")
val tencentMapKey: String = localProps.getProperty("ops.map.tencentKey", configTencentKey).orEmpty()
    .ifBlank { configTencentKey }

android {
    namespace = "com.luopingtech.ebike.ops"
    compileSdk = libs.versions.android.compileSdk.get().toInt()

    defaultConfig {
        applicationId = configApplicationId
        minSdk = libs.versions.android.minSdk.get().toInt()
        targetSdk = libs.versions.android.targetSdk.get().toInt()
        versionCode = 1
        versionName = "1.0.0"
        manifestPlaceholders["TENCENT_MAP_KEY"] = tencentMapKey
        buildConfigField("String", "TENCENT_MAP_KEY", "\"${tencentMapKey.replace("\"", "\\\"")}\"")
        buildConfigField("String", "OPS_TENANT", "\"${opsTenant.replace("\"", "\\\"")}\"")
        buildConfigField("String", "OPS_MODE", "\"${opsMode.replace("\"", "\\\"")}\"")
        buildConfigField("String", "OPS_DISPLAY_NAME", "\"${configDisplayName.replace("\"", "\\\"")}\"")
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

val syncTenantConfig by tasks.registering {
    group = "ops"
    description = "Copy config/{tenant}_{mode}.json → androidApp/src/main/assets/tenant.json"
    inputs.file(tenantConfigFile).optional()
    outputs.file(layout.projectDirectory.file("src/main/assets/tenant.json"))
    doLast {
        check(tenantConfigFile.exists()) {
            "Missing tenant config: ${tenantConfigFile.path}. Set ops.tenant / ops.mode in local.properties."
        }
        val outDir = layout.projectDirectory.dir("src/main/assets").asFile
        outDir.mkdirs()
        val outFile = outDir.resolve("tenant.json")
        tenantConfigFile.copyTo(outFile, overwrite = true)
        logger.lifecycle("Synced ${tenantConfigFile.name} → assets/tenant.json (appId=$configApplicationId)")
    }
}

tasks.named("preBuild").configure {
    dependsOn(syncTenantConfig)
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
    debugImplementation(libs.androidx.compose.ui.tooling)
}
