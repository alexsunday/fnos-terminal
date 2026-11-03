package main

import (
	"os"
)

const scriptContent = `#!/bin/bash

# Ensure runtime data directory exists
mkdir -p "${TRIM_PKGVAR}"

LOG_FILE="${TRIM_PKGVAR}/app.log"
PID_FILE="${TRIM_PKGVAR}/app.pid"
SOCKET_PATH="${TRIM_APPDEST}/app.sock"
GATEWAY_PREFIX="/app/terminal"

# Locate binary in target/bin/terminal or target/app/bin/terminal
if [ -f "${TRIM_APPDEST}/bin/terminal" ]; then
  APP_BIN="${TRIM_APPDEST}/bin/terminal"
elif [ -f "${TRIM_APPDEST}/app/bin/terminal" ]; then
  APP_BIN="${TRIM_APPDEST}/app/bin/terminal"
else
  APP_BIN="/var/apps/terminal/target/bin/terminal"
fi

log_msg() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

check_process() {
  local pid="$1"
  kill -0 "$pid" 2>/dev/null
}

status_app() {
  if [ -f "$PID_FILE" ]; then
    local pid
    pid="$(head -n 1 "$PID_FILE" | tr -d '[:space:]')"
    if [ -n "$pid" ] && check_process "$pid"; then
      return 0
    fi
    rm -f "$PID_FILE"
  fi
  return 1
}

start_app() {
  if status_app; then
    return 0
  fi

  log_msg "Starting fnOS Terminal daemon..."
  chmod +x "$APP_BIN"
  rm -f "$SOCKET_PATH"

  if [ ! -f "$APP_BIN" ]; then
    log_msg "ERROR: APP_BIN not found at $APP_BIN"
    if [ -n "$TRIM_TEMP_LOGFILE" ]; then
      echo "Binary not found at $APP_BIN" > "$TRIM_TEMP_LOGFILE"
    fi
    exit 1
  fi

  "$APP_BIN" -socket "$SOCKET_PATH" -prefix "$GATEWAY_PREFIX" >> "$LOG_FILE" 2>&1 &
  local new_pid="$!"
  echo "$new_pid" > "$PID_FILE"
  log_msg "fnOS Terminal started with PID $new_pid"

  # Brief wait to ensure the process did not exit immediately
  sleep 1
  if check_process "$new_pid"; then
    return 0
  else
    log_msg "ERROR: fnOS Terminal process exited immediately. Check $LOG_FILE"
    if [ -n "$TRIM_TEMP_LOGFILE" ]; then
      tail -n 10 "$LOG_FILE" > "$TRIM_TEMP_LOGFILE"
    fi
    rm -f "$PID_FILE"
    exit 1
  fi
}

stop_app() {
  log_msg "Stopping fnOS Terminal daemon..."

  if [ ! -f "$PID_FILE" ]; then
    rm -f "$SOCKET_PATH"
    return 0
  fi

  local pid
  pid="$(head -n 1 "$PID_FILE" | tr -d '[:space:]')"

  if [ -z "$pid" ] || ! [[ "$pid" =~ ^[0-9]+$ ]] || ! check_process "$pid"; then
    rm -f "$PID_FILE"
    rm -f "$SOCKET_PATH"
    return 0
  fi

  kill -TERM "$pid" 2>> "$LOG_FILE" || true

  local count=0
  while check_process "$pid" && [ "$count" -lt 10 ]; do
    sleep 1
    count=$((count + 1))
  done

  if check_process "$pid"; then
    kill -KILL "$pid" 2>> "$LOG_FILE" || true
  fi

  rm -f "$PID_FILE"
  rm -f "$SOCKET_PATH"
  log_msg "fnOS Terminal stopped."
}

case "$1" in
  start)
    start_app
    ;;
  stop)
    stop_app
    ;;
  status)
    if status_app; then
      exit 0
    fi
    exit 3
    ;;
  *)
    echo "Usage: $0 {start|stop|status}" >&2
    exit 1
    ;;
esac
`

func main() {
	// Write with strict LF (no CRLF)
	content := []byte(scriptContent)
	if err := os.WriteFile("package/cmd/main", content, 0755); err != nil {
		panic(err)
	}
}
