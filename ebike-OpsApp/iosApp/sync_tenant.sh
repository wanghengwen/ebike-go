#!/bin/bash
# 对齐 Android 的 syncTenantConfig：把 config/{tenant}_{mode}.json 拷进 iOS 宿主，
# 并改写 project.yml 里的 Bundle ID / 显示名 / 地图 Key（字面量，避开 CocoaPods
# 对 xcodegen configFiles 的一致性问题）。
# local.properties 里的 ops.tenant / ops.mode / ops.map.tencentKey 与 Android 共用同一份。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
IOS="$ROOT/iosApp"
LOCAL="$ROOT/local.properties"
YML="$IOS/project.yml"

prop() {
  local key="$1" default="${2:-}"
  if [ -f "$LOCAL" ]; then
    local v
    v=$(grep -E "^${key}=" "$LOCAL" | tail -1 | cut -d= -f2- | tr -d '\r' || true)
    if [ -n "${v:-}" ]; then echo "$v"; return; fi
  fi
  echo "$default"
}

TENANT="$(prop ops.tenant demo)"
MODE="$(prop ops.mode release)"
# 优先直接用 Android 运行时那份 assets/tenant.json，保证两端连接信息字节级一致；
# 没有时再退回 config/{tenant}_{mode}.json（跟 Android gradle sync 同源）。
ANDROID_ASSET="$ROOT/androidApp/src/main/assets/tenant.json"
SRC="$ROOT/config/${TENANT}_${MODE}.json"
if [ -f "$ANDROID_ASSET" ]; then
  SRC="$ANDROID_ASSET"
elif [ ! -f "$SRC" ]; then
  echo "missing $SRC — falling back to demo_release.json" >&2
  SRC="$ROOT/config/demo_release.json"
  TENANT=demo
  MODE=release
fi

echo "source: $SRC"
cp "$SRC" "$IOS/OpsAppHost/tenant.json"

OVERRIDE_KEY="$(prop ops.map.tencentKey)"

python3 - <<PY
import json, pathlib, re
src = pathlib.Path(r"""$SRC""")
yml_path = pathlib.Path(r"""$YML""")
cfg = json.loads(src.read_text(encoding="utf-8"))
app = cfg.get("app") or {}
mp = cfg.get("map") or {}
bundle = (app.get("iosBundleId") or "com.luopingtech.ebike.ops.demo").strip()
name = (app.get("displayName") or cfg.get("name") or "OpsApp").strip()
key = (mp.get("tencentKey") or "").strip()
override = r"""$OVERRIDE_KEY""".strip()
if override:
    key = override

text = yml_path.read_text(encoding="utf-8")

def replace_marked(text: str, marker: str, key_line: str, value: str) -> str:
    # 匹配「# SYNC:MARKER」下一行的「Key: ...」（兼容 CRLF）
    pattern = rf"(# SYNC:{re.escape(marker)}\r?\n\s*{re.escape(key_line)}:\s*).*"
    repl = rf"\g<1>{value}"
    new, n = re.subn(pattern, repl, text, count=1)
    if n != 1:
        raise SystemExit(f"failed to patch {marker} ({key_line}) in project.yml")
    return new

text = replace_marked(text, "OPS_DISPLAY_NAME", "CFBundleDisplayName", name)
text = replace_marked(text, "OPS_TENCENT_MAP_KEY", "OpsTencentMapKey", key if key else '""')
text = replace_marked(text, "OPS_IOS_BUNDLE_ID", "PRODUCT_BUNDLE_IDENTIFIER", bundle)
yml_path.write_text(text, encoding="utf-8")
print(f"synced tenant={cfg.get('alias', '?')} bundle={bundle} mapKey={'set' if key else 'empty'}")
PY
