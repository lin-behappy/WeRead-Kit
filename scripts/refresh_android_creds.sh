#!/usr/bin/env bash
# 刷新安卓客户端凭据并导入 weread-kit。
#
# 背景: 微信读书的 access token 由客户端在启动/联网时续期，并延迟写回数据库。
# 直接读 WRAccount 常常拿到已过期的 token(-2012 登录超时)。所以本脚本会先
# 拉起客户端、等它写回新 token，再调用 extract_android_creds.sh 提取并导入。
# 客户端续期后旧 token 立即失效，而新 token 写库有延迟，因此导入后会用
# 周奖励接口校验通道，失败则重新提取重试。
#
# 用法:
#   scripts/refresh_android_creds.sh [main|dual|all] [--no-import]
#
# 环境变量:
#   ANDROID_SERIAL  adb 设备序列号(多设备时必须设置)
#   BASE            服务地址，默认 http://192.168.43.88:18081
#   WEREAD_USERNAME 登录用户名，默认 admin
#   WEREAD_PASSWORD 登录密码(未设置时只提取，不导入)
#   LAUNCH_WAIT     拉起客户端后的等待秒数，默认 15
#   RETRIES         导入校验的重试次数，默认 3
set -euo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
extract="$here/extract_android_creds.sh"
mode=${1:-all}
no_import=0
[[ ${2:-} == "--no-import" ]] && no_import=1

BASE=${BASE:-http://192.168.43.88:18081}
WEREAD_USERNAME=${WEREAD_USERNAME:-admin}
LAUNCH_WAIT=${LAUNCH_WAIT:-15}
RETRIES=${RETRIES:-3}

case "$mode" in
  main|dual|all) ;;
  *) echo "用法: $0 [main|dual|all] [--no-import]" >&2; exit 1 ;;
esac

if [[ -n ${ANDROID_SERIAL:-} ]]; then
  adb_args=(-s "$ANDROID_SERIAL")
else
  adb_args=()
  count=$(adb devices | awk 'NR>1 && $2=="device"' | wc -l)
  if (( count > 1 )); then
    echo "检测到多台设备，请设置 ANDROID_SERIAL" >&2
    exit 1
  fi
fi

tmp=$(mktemp -d "${TMPDIR:-/tmp}/refresh-android-XXXXXX")
trap 'rm -rf "$tmp"' EXIT

# vid 与提取模式一一对应: 主应用 384059045，应用双开 443553272。
vid_of() {
  case "$1" in
    main) echo 384059045 ;;
    dual) echo 443553272 ;;
  esac
}

launch() {
  local user_args=()
  [[ $1 == dual ]] && user_args=(--user 999)
  adb "${adb_args[@]}" shell am start "${user_args[@]}" \
    -n com.tencent.weread/com.tencent.weread.LauncherActivity >/dev/null
}

login() {
  curl -s -c "$1" -o /dev/null -X POST "$BASE/login" \
    --data-urlencode "username=$WEREAD_USERNAME" \
    --data-urlencode "password=$WEREAD_PASSWORD"
}

weekly_channel() {
  curl -s -b "$1" "$BASE/api/accounts/$2/weekly" |
    python3 -c 'import json,sys; print(json.load(sys.stdin).get("channel") or "?")'
}

modes=(main dual)
[[ $mode != all ]] && modes=("$mode")

for m in "${modes[@]}"; do
  vid=$(vid_of "$m")
  out="$tmp/android-$m.json"
  cookie="$tmp/cookies-$m"

  echo "[$m] 拉起微信读书客户端，等待 ${LAUNCH_WAIT}s 让 token 续期…"
  launch "$m"
  sleep "$LAUNCH_WAIT"

  if [[ $no_import -eq 1 || -z ${WEREAD_PASSWORD:-} ]]; then
    "$extract" "$m" "$out"
    echo "[$m] 已提取到 $out (未导入)"
    continue
  fi

  login "$cookie"
  ok=0
  for attempt in $(seq 1 "$RETRIES"); do
    "$extract" "$m" "$out" >/dev/null
    code=$(curl -s -b "$cookie" -o /dev/null -w '%{http_code}' -X POST \
      -H 'Content-Type: application/json' --data-binary @"$out" \
      "$BASE/api/accounts/$vid/android")
    if [[ $code != 200 ]]; then
      echo "[$m] 导入失败: HTTP $code" >&2
      exit 1
    fi
    channel=$(weekly_channel "$cookie" "$vid")
    if [[ $channel == android ]]; then
      ok=1
      break
    fi
    echo "[$m] 第 $attempt 次校验通道为 $channel，等待客户端写回新 token…"
    sleep 10
  done

  if [[ $ok -eq 1 ]]; then
    echo "[$m] 导入成功，通道: android"
  else
    echo "[$m] 导入后通道仍不是 android；请在手机上打开微信读书确认登录状态，再重试" >&2
    exit 1
  fi
done
