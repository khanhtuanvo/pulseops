#!/usr/bin/env bash
set -euo pipefail

# Synthetic alert generator for PulseOps.
# Used for manual demo setup and load testing of the webhook pipeline.
#
# Usage:
#   ./scripts/load-test/generate-alerts.sh --api-key KEY [options]
#
# Options:
#   --count N        number of alerts to send        (default 100)
#   --rate N         alerts per second                (default 10)
#   --api-key KEY    team API key (required)
#   --url URL        webhook URL                      (default http://localhost:8080/webhooks/alerts)
#   --scenario NAME  'random' (varied fingerprints) | 'duplicate' (same fingerprint)  (default random)

COUNT=100
RATE=10
API_KEY=""
URL="http://localhost:8080/webhooks/alerts"
SCENARIO="random"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --count) COUNT="$2"; shift 2 ;;
    --rate) RATE="$2"; shift 2 ;;
    --api-key) API_KEY="$2"; shift 2 ;;
    --url) URL="$2"; shift 2 ;;
    --scenario) SCENARIO="$2"; shift 2 ;;
    -h|--help) sed -n '3,16p' "$0"; exit 0 ;;
    *) echo "unknown argument: $1" >&2; exit 1 ;;
  esac
done

if [[ -z "$API_KEY" ]]; then
  echo "error: --api-key is required" >&2
  exit 1
fi
if [[ "$SCENARIO" != "random" && "$SCENARIO" != "duplicate" ]]; then
  echo "error: --scenario must be 'random' or 'duplicate'" >&2
  exit 1
fi
if ! [[ "$COUNT" =~ ^[0-9]+$ && "$RATE" =~ ^[0-9]+$ && "$RATE" -gt 0 ]]; then
  echo "error: --count and --rate must be positive integers" >&2
  exit 1
fi

SOURCES=(prometheus datadog grafana cloudwatch sentry)
ALERT_NAMES=(HighCPU HighMemory DiskFull LatencySpike ErrorRateHigh PodCrashLoop)
SEVERITIES=(CRITICAL HIGH MEDIUM LOW)
ENVIRONMENTS=(prod staging)

interval="$(awk "BEGIN { print 1 / $RATE }")"
declare -A status_counts=()
sent=0
start_ts="$(date +%s)"

pick() {
  local -n arr="$1"
  echo "${arr[$((RANDOM % ${#arr[@]}))]}"
}

print_progress() {
  local elapsed effective breakdown=""
  elapsed=$(( $(date +%s) - start_ts ))
  [[ "$elapsed" -eq 0 ]] && elapsed=1
  effective="$(awk "BEGIN { printf \"%.1f\", $sent / $elapsed }")"
  for code in $(printf '%s\n' "${!status_counts[@]}" | sort); do
    breakdown+=" ${code}=${status_counts[$code]}"
  done
  printf 'sent=%d/%d  rate=%s/s  status:%s\n' "$sent" "$COUNT" "$effective" "$breakdown"
}

for ((i = 1; i <= COUNT; i++)); do
  if [[ "$SCENARIO" == "duplicate" ]]; then
    source="prometheus"; alert_name="HighCPU"; severity="CRITICAL"; environment="prod"
  else
    source="$(pick SOURCES)"; alert_name="$(pick ALERT_NAMES)"
    severity="$(pick SEVERITIES)"; environment="$(pick ENVIRONMENTS)"
  fi

  payload="$(printf '{"source":"%s","alertName":"%s","severity":"%s","environment":"%s"}' \
    "$source" "$alert_name" "$severity" "$environment")"

  code="$(curl -s -o /dev/null -w '%{http_code}' \
    -X POST "$URL" \
    -H "X-API-Key: $API_KEY" \
    -H "Content-Type: application/json" \
    -d "$payload" || echo "000")"
  status_counts["$code"]=$(( ${status_counts["$code"]:-0} + 1 ))
  sent=$((sent + 1))

  if (( sent % 10 == 0 )); then
    print_progress
  fi

  if (( i < COUNT )); then
    sleep "$interval"
  fi
done

echo "---"
print_progress
echo "done."
