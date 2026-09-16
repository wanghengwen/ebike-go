rootProject.name = "ebike-RiderApp"

pluginManagement {
    repositories {
        // 国内直连 Maven Central / Google 经常超时；阿里云镜像优先。
        maven(url = "https://maven.aliyun.com/repository/google")
        maven(url = "https://maven.aliyun.com/repository/central")
        maven(url = "https://maven.aliyun.com/repository/gradle-plugin")
        google()
        mavenCentral()
        gradlePluginPortal()
    }
}

dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
        maven(url = "https://maven.aliyun.com/repository/google")
        maven(url = "https://maven.aliyun.com/repository/central")
        maven(url = "https://maven.aliyun.com/repository/public")
        google()
        mavenCentral()
        // Tencent Map / Location SDK
        maven(url = "https://mirrors.tencent.com/nexus/repository/maven-public/")
        maven(url = "https://mirrors.tencent.com/repository/maven/tencent_public/")
    }
}

include(":shared")
include(":sharedUi")
include(":androidApp")
