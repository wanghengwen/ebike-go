#!/bin/bash
# iOS 开发者首次 / 日常启动前准备：租户资源 + SharedUi.framework + CocoaPods。
# 工程文件（.xcodeproj / .xcworkspace）已入库，一般不必再跑 xcodegen。
set -euo pipefail
export PATH="/usr/bin:/bin:/usr/sbin:/sbin:/opt/homebrew/bin:${PATH:-}"
export LANG="${LANG:-en_US.UTF-8}"
export LC_ALL="${LC_ALL:-en_US.UTF-8}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

MODE="${1:-sim}" # sim | device

echo "### 0/3 sync tenant"
bash iosApp/sync_tenant.sh

case "$MODE" in
  device)
    TASK=":sharedUi:linkDebugFrameworkIosArm64"
    ;;
  sim|simulator)
    TASK=":sharedUi:linkDebugFrameworkIosSimulatorArm64"
    ;;
  *)
    echo "usage: $0 [sim|device]" >&2
    exit 1
    ;;
esac

echo "### 1/3 gradle $TASK"
./gradlew "$TASK"

echo "### 2/3 pod install"
cd iosApp
if ! command -v pod >/dev/null 2>&1; then
  echo "CocoaPods missing: brew install cocoapods" >&2
  exit 1
fi
pod install

echo "### done"
echo "open: $ROOT/iosApp/RiderAppHost.xcworkspace"
echo "提示：腾讯地图仅真机可用；模拟器自动 MapKit。"
