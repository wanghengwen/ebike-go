#!/bin/bash
# 对齐 Android 的 syncTenantConfig：把 config/{tenant}_{mode}.json 拷进 iOS 宿主，
# 并改写 project.yml / Info.plist / project.pbxproj 里的 Bundle ID / 显示名 / 地图 Key。
# （工程已入库，日常不必再跑 xcodegen；换租户只跑本脚本即可。）
# local.properties 里的 rider.tenant / rider.mode / rider.map.tencentKey 与 Android 共用同一份。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
IOS="$ROOT/iosApp"
LOCAL="$ROOT/local.properties"
YML="$IOS/project.yml"
PLIST="$IOS/RiderAppHost/Info.plist"
PBX="$IOS/RiderAppHost.xcodeproj/project.pbxproj"

prop() {
  local key="$1" default="${2:-}"
  if [ -f "$LOCAL" ]; then
    local v
    v=$(grep -E "^${key}=" "$LOCAL" | tail -1 | cut -d= -f2- | tr -d '\r' || true)
    if [ -n "${v:-}" ]; then echo "$v"; return; fi
  fi
  echo "$default"
}

TENANT="$(prop rider.tenant demo)"
MODE="$(prop rider.mode release)"
# 优先直接用 Android 运行时那份 assets/tenant.json，保证两端连接信息字节级一致；
# 没有时再退回 config/{tenant}_{mode}.json（跟 Android gradle sync 同源）。
ANDROID_ASSET="$ROOT/androidApp/src/main/assets/tenant.json"
SRC="$ROOT/config/${TENANT}_${MODE}.json"
if [ -f "$ANDROID_ASSET" ]; then
  SRC="$ANDROID_ASSET"
elif [ ! -f "$SRC" ]; then
  echo "missing $SRC — falling back to demo_release.json" >&2
  SRC="$ROOT/config/demo_release.json"
fi

echo "source: $SRC"
mkdir -p "$IOS/RiderAppHost"
cp "$SRC" "$IOS/RiderAppHost/tenant.json"

OVERRIDE_KEY="$(prop rider.map.tencentKey)"

python3 - <<PY
import json, pathlib, re
src = pathlib.Path(r"""$SRC""")
yml_path = pathlib.Path(r"""$YML""")
plist_path = pathlib.Path(r"""$PLIST""")
pbx_path = pathlib.Path(r"""$PBX""")
cfg = json.loads(src.read_text(encoding="utf-8"))
app = cfg.get("app") or {}
mp = cfg.get("map") or {}
bundle = (app.get("iosBundleId") or "com.luopingtech.ebike.rider.demo").strip()
name = (app.get("displayName") or cfg.get("name") or "RiderApp").strip()
key = (mp.get("tencentKey") or "").strip()
override = r"""$OVERRIDE_KEY""".strip()
if override:
    key = override
key_plist = key if key else ""

text = yml_path.read_text(encoding="utf-8")

def replace_marked(text: str, marker: str, key_line: str, value: str) -> str:
    # 匹配「# SYNC:MARKER」下一行的「Key: ...」（兼容 CRLF）
    pattern = rf"(# SYNC:{re.escape(marker)}\r?\n\s*{re.escape(key_line)}:\s*).*"
    repl = rf"\g<1>{value}"
    new, n = re.subn(pattern, repl, text, count=1)
    if n != 1:
        raise SystemExit(f"failed to patch {marker} ({key_line}) in project.yml")
    return new

text = replace_marked(text, "RIDER_DISPLAY_NAME", "CFBundleDisplayName", name)
text = replace_marked(text, "RIDER_TENCENT_MAP_KEY", "RiderTencentMapKey", key if key else '""')
text = replace_marked(text, "RIDER_IOS_BUNDLE_ID", "PRODUCT_BUNDLE_IDENTIFIER", bundle)
yml_path.write_text(text, encoding="utf-8")

# 已入库的 Xcode 工程：同步改 Info.plist / pbxproj，免去 xcodegen。
if plist_path.is_file():
    plist = plist_path.read_text(encoding="utf-8")

    def replace_plist_string(xml: str, plist_key: str, value: str) -> str:
        pattern = rf"(<key>{re.escape(plist_key)}</key>\s*<string>)(.*?)(</string>)"
        new, n = re.subn(pattern, rf"\g<1>{value}\g<3>", xml, count=1, flags=re.S)
        if n != 1:
            raise SystemExit(f"failed to patch Info.plist key {plist_key}")
        return new

    plist = replace_plist_string(plist, "CFBundleDisplayName", name)
    plist = replace_plist_string(plist, "RiderTencentMapKey", key_plist)
    plist_path.write_text(plist, encoding="utf-8")

if pbx_path.is_file():
    pbx = pbx_path.read_text(encoding="utf-8")
    new, n = re.subn(
        r"PRODUCT_BUNDLE_IDENTIFIER = [^;]+;",
        f"PRODUCT_BUNDLE_IDENTIFIER = {bundle};",
        pbx,
    )
    if n < 1:
        raise SystemExit("failed to patch PRODUCT_BUNDLE_IDENTIFIER in project.pbxproj")
    pbx_path.write_text(new, encoding="utf-8")

print(f"synced tenant={cfg.get('alias', '?')} bundle={bundle} mapKey={'set' if key else 'empty'}")
PY
