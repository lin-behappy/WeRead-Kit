#!/usr/bin/env bash
# 从已 root 的安卓设备提取微信读书安卓客户端凭据（7 天档 6 书币用）。
#
# 用法:
#   scripts/extract_android_creds.sh main <output.json>   # 主应用
#   scripts/extract_android_creds.sh dual <output.json>   # MIUI 应用双开 (XSpace user 999)
#
# 依赖: adb、设备已 root、设备上已登录对应微信读书账号。
# 输出文件权限为 0600，内含 accessToken/refreshToken，勿提交到仓库。
set -euo pipefail

mode=${1:-main}
out=${2:-}
if [[ -z "$out" ]]; then
  echo "用法: $0 [main|dual] <output.json>" >&2
  exit 1
fi
case "$mode" in
  main) pkgdir=/data/data/com.tencent.weread ;;
  dual) pkgdir=/data/user/999/com.tencent.weread ;;
  *) echo "mode 必须是 main 或 dual" >&2; exit 1 ;;
esac

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

# adb exec-out cat 会破坏二进制；统一用 base64 传输。
adb exec-out su -c "base64 $pkgdir/databases/WRAccount" | tr -d '\r' | base64 -d > "$tmp/WRAccount.db"
if adb shell su -c "test -f $pkgdir/databases/WRAccount-wal" >/dev/null 2>&1; then
  adb exec-out su -c "base64 $pkgdir/databases/WRAccount-wal" | tr -d '\r' | base64 -d > "$tmp/WRAccount.db-wal"
fi
adb exec-out su -c "cat $pkgdir/shared_prefs/device.xml" > "$tmp/device.xml"

brand=$(adb shell getprop ro.product.brand | tr -d '\r' | tr '[:upper:]' '[:lower:]')
model=$(adb shell getprop ro.product.model | tr -d '\r')
release=$(adb shell getprop ro.build.version.release | tr -d '\r')
sdk=$(adb shell getprop ro.build.version.sdk | tr -d '\r')
build=$(adb shell getprop ro.build.id | tr -d '\r')
pkg=$(adb shell dumpsys package com.tencent.weread | tr -d '\r')
version_name=$(printf '%s\n' "$pkg" | sed -n 's/.*versionName=\([^ ]*\).*/\1/p' | head -1)
version_code=$(printf '%s\n' "$pkg" | sed -n 's/.*versionCode=\([0-9]*\).*/\1/p' | head -1)
if [[ -z "$version_name" || -z "$version_code" ]]; then
  echo "无法从 dumpsys package 读取微信读书版本号" >&2
  exit 1
fi
appver="$version_name.$version_code"
ua="WeRead/$version_name WRBrand/$brand Dalvik/2.1.0 (Linux; U; Android $release; $model Build/$build)"

python3 - "$tmp" "$out" "$model" "$release" "$sdk" "$appver" "$ua" <<'PY'
import json, os, sqlite3, sys, xml.etree.ElementTree as ET

tmp, out, model, release, sdk, appver, ua = sys.argv[1:]
db = sqlite3.connect(os.path.join(tmp, "WRAccount.db"))
cols = [r[1] for r in db.execute("pragma table_info(Account)")]
row = None
try:
    setting_cols = [r[1] for r in db.execute("pragma table_info(Setting)")]
    value_col = "val" if "val" in setting_cols else "value"
    cur = db.execute(f"select {value_col} from Setting where key='current_login_account'").fetchone()
except sqlite3.Error:
    cur = None
if cur:
    row = db.execute("select * from Account where id=?", (int(cur[0]),)).fetchone()
if row is None:
    row = db.execute("select * from Account order by id limit 1").fetchone()
if row is None:
    sys.exit("Account 表为空: 请先在安卓客户端登录")
acc = dict(zip(cols, row))

root = ET.parse(os.path.join(tmp, "device.xml")).getroot()
dev = {}
for el in root:
    name = el.get("name")
    if not name:
        continue
    # <int ... value="..."/> 用属性；<string ...>文本</string> 用 text。
    dev[name] = el.get("value") if el.get("value") is not None else (el.text or "")
cred = {
    "access_token": acc["accessToken"],
    "refresh_token": acc["refreshToken"],
    "device_id": dev.get("deviceid", ""),
    "device_name": model,
    "install_id": dev.get("installId", ""),
    "user_agent": ua,
    "app_version": appver,
    "baseapi": sdk,
    "osver": release,
    "channel_id": "1",
}
fd = os.open(out, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
with os.fdopen(fd, "w") as f:
    json.dump(cred, f, ensure_ascii=False, indent=2)
print(f"已写入 {out} (vid={acc['vid']})")
PY
