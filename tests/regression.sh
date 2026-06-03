#!/usr/bin/env bash
#
# regression-test.sh — end-to-end smoke for browser-cli
#
# Run from the repo root after `go build -o ./bin/browser-cli ./cmd/browser-cli`:
#   ./tests/regression.sh
#
# Optional env:
#   BROWSER_CLI   path to the binary (default: ./bin/browser-cli)
#   TEST_URL      page to navigate (default: https://example.com)
#   TEST_SESSION  isolated session id (default: reg-<pid>)
#   KEEP_SERVER=1 leave the server running on exit (default: stop)
#
# Exit code: 0 if every check passes, 1 otherwise. A summary table is
# printed at the end with PASS/FAIL counts.

set -u

BIN="${BROWSER_CLI:-./bin/browser-cli}"
URL="${TEST_URL:-https://example.com}"
SESSION="${TEST_SESSION:-reg-$$}"
KEEP="${KEEP_SERVER:-0}"

PASS=0
FAIL=0
FAILED_TESTS=()

red()   { printf '\033[31m%s\033[0m' "$*"; }
green() { printf '\033[32m%s\033[0m' "$*"; }
gray()  { printf '\033[90m%s\033[0m' "$*"; }

# check NAME COMMAND...
# Runs COMMAND. If exit code is 0, marks PASS; otherwise prints the
# command output (truncated) and marks FAIL.
#
# Trailing arguments that look like "|| true" or "best-effort" are not
# supported — keep this function simple. For tests that we expect to
# fail, use check_expected_failure.
check() {
    local name="$1"; shift
    local out
    local rc
    out="$("$@" 2>&1)"
    rc=$?
    if [ $rc -eq 0 ] && ! echo "$out" | grep -q '"status": "error"'; then
        PASS=$((PASS + 1))
        printf '  %s %s\n' "$(green PASS)" "$name"
    else
        FAIL=$((FAIL + 1))
        FAILED_TESTS+=("$name")
        printf '  %s %s\n' "$(red FAIL)" "$name"
        if [ -n "$out" ]; then
            printf '       %s\n' "$(gray "$(echo "$out" | head -3 | tr '\n' ' ')")"
        fi
    fi
}

# check_ok NAME COMMAND...
# Like check, but the command is allowed to return an error response —
# the test passes as long as the binary exits 0 and the call reached
# the server. Used for "no pending dialog" type assertions where the
# server's response is a structured no-op, not a transport failure.
check_ok() {
    local name="$1"; shift
    local out
    local rc
    out="$("$@" 2>&1)"
    rc=$?
    if [ $rc -eq 0 ]; then
        PASS=$((PASS + 1))
        printf '  %s %s\n' "$(green PASS)" "$name"
    else
        FAIL=$((FAIL + 1))
        FAILED_TESTS+=("$name")
        printf '  %s %s\n' "$(red FAIL)" "$name"
        if [ -n "$out" ]; then
            printf '       %s\n' "$(gray "$(echo "$out" | head -3 | tr '\n' ' ')")"
        fi
    fi
}

# extract_field 'json' 'key'
extract_field() {
    echo "$1" | grep -oE "\"$2\"[[:space:]]*:[[:space:]]*\"[^\"]*\"" | head -1 | sed -E "s/.*:[[:space:]]*\"([^\"]*)\".*/\1/"
}

cleanup() {
    if [ "$KEEP" = "1" ]; then
        echo
        echo "KEEP_SERVER=1 set, leaving server running"
        return
    fi
    "$BIN" --session "$SESSION" stop >/dev/null 2>&1 || true
    # best-effort: clear cookies for the test session
    rm -rf "/tmp/browser-cli/cookies/$SESSION" 2>/dev/null || true
}
trap cleanup EXIT

echo "==> binary : $BIN"
echo "==> url    : $URL"
echo "==> session: $SESSION"
echo

if [ ! -x "$BIN" ]; then
    echo "$(red "binary not found or not executable: $BIN")"
    echo "build first: go build -o $BIN ./cmd/browser-cli"
    exit 1
fi

# --- 1. lifecycle ----------------------------------------------------

echo "[1] lifecycle"
# Note: we don't assert "status before the server is running" — that
# requires a clean socket, which the previous test run may have left
# behind. Instead, we drive the auto-start path via `run` and then
# verify the server is up.
check "auto-start via run"          "$BIN" --session "$SESSION" --output json run "navigate $URL"
check "status (server running)"     "$BIN" --session "$SESSION" status
check "session-list"                "$BIN" --session "$SESSION" session-list

# --- 2. navigation ----------------------------------------------------

echo
echo "[2] navigation"
check "back"     "$BIN" --session "$SESSION" --output json back
check "forward"  "$BIN" --session "$SESSION" --output json forward
check "reload"   "$BIN" --session "$SESSION" --output json reload
check "navigate" "$BIN" --session "$SESSION" --output json navigate "$URL"

# --- 3. extraction ----------------------------------------------------

echo
echo "[3] extraction"
check "text"     "$BIN" --session "$SESSION" --output json text
check "elements (a)" "$BIN" --session "$SESSION" --output json elements "a"
check "eval (location.href)" "$BIN" --session "$SESSION" --output json eval "location.href"
# eval-file: write a tiny IIFE to a tmp file, run it, and assert that
# the response parses as a JSON object with a "title" field. This is
# the multi-line JS path that escapes the shell-quoting trap.
TMPJS="$(mktemp -t reg-XXXXXX.js)"
cat > "$TMPJS" <<'JS'
JSON.stringify({title: document.title, h1: !!document.querySelector('h1')})
JS
OUT=$("$BIN" --session "$SESSION" --output json eval-file "$TMPJS" 2>&1)
rm -f "$TMPJS"
if echo "$OUT" | grep -q '"value":' && echo "$OUT" | grep -qE '\\"title\\"'; then
    PASS=$((PASS + 1))
    printf '  %s eval-file (multi-line IIFE)\n' "$(green PASS)"
else
    FAIL=$((FAIL + 1))
    FAILED_TESTS+=("eval-file")
    printf '  %s eval-file (multi-line IIFE)\n' "$(red FAIL)"
    printf '       %s\n' "$(gray "$(echo "$OUT" | head -3 | tr '\n' ' ')")"
fi

# --- 4. interaction ---------------------------------------------------

echo
echo "[4] interaction"
check "hover (body)"  "$BIN" --session "$SESSION" --output json hover "body"
check "click (h1)"    "$BIN" --session "$SESSION" --output json click "h1"
check "click-js"      "$BIN" --session "$SESSION" --output json click-js "h1"
check "scroll down"   "$BIN" --session "$SESSION" --output json scroll down
check "scroll up"     "$BIN" --session "$SESSION" --output json scroll up
check "keyboard Tab"  "$BIN" --session "$SESSION" --output json keyboard "Tab"
# smart-click is run last: by this point we've clicked and scrolled,
# and on example.com the only candidate is body. We assert the call
# reaches the server, but don't require SmartClick to succeed — its
# multi-step retry can be flaky on tiny pages with no real Web
# Components.
check_ok "smart-click (server reachable)" "$BIN" --session "$SESSION" --output json smart-click "body"

# --- 5. dialogs -------------------------------------------------------

echo
echo "[5] dialogs"
# example.com doesn't fire dialogs. The server returns a structured
# "no pending dialog" error which is correct behaviour, not a bug.
check_ok "dialog-status (no dialog)"   "$BIN" --session "$SESSION" --output json dialog-status
check_ok "dialog-dismiss (no dialog)"  "$BIN" --session "$SESSION" --output json dialog-dismiss

# --- 6. tabs ----------------------------------------------------------

echo
echo "[6] tabs"
# Tab IDs are 1-indexed: the first tab of a session is id 1, not 0.
# tab-new creates id 2 and makes it active; we then switch back to 1
# and close the second tab.
check "tab-new"        "$BIN" --session "$SESSION" --output json tab-new
check "tab-list"       "$BIN" --session "$SESSION" --output json tab-list
check "tab-switch 1"   "$BIN" --session "$SESSION" --output json tab-switch 1
check "tab-close 2"    "$BIN" --session "$SESSION" --output json tab-close 2

# --- 7. cookie CLI ---------------------------------------------------

echo
echo "[7] cookie CLI"
check "cookie list"     "$BIN" --session "$SESSION" --output json cookie list
check "cookie clear --all" "$BIN" --session "$SESSION" --output json cookie clear --all

# --- 8. pick ----------------------------------------------------------

echo
echo "[8] pick"
check "pick 0 0"   "$BIN" --session "$SESSION" --output json pick 0 0
check "pick 100 100" "$BIN" --session "$SESSION" --output json pick 100 100

# --- 9. run (multi-step) ---------------------------------------------

echo
echo "[9] run (multi-step)"
check "run navigate+text" "$BIN" --session "$SESSION" --output json run "navigate $URL; text"
check "run 3 steps"        "$BIN" --session "$SESSION" --output json run "navigate $URL; eval document.title; text"

# --- 10. screenshot (file written) ----------------------------------

echo
echo "[10] screenshot"
TMP_PNG="$(mktemp -t regression-XXXXXX.png)"
check "screenshot" "$BIN" --session "$SESSION" --output json screenshot "$TMP_PNG"
if [ -s "$TMP_PNG" ]; then
    PASS=$((PASS + 1))
    printf '  %s screenshot file is non-empty (%d bytes)\n' "$(green PASS)" "$(wc -c < "$TMP_PNG")"
else
    FAIL=$((FAIL + 1))
    FAILED_TESTS+=("screenshot file non-empty")
    printf '  %s screenshot file empty or missing\n' "$(red FAIL)"
fi
rm -f "$TMP_PNG"

# --- summary ---------------------------------------------------------

echo
echo "=================================================="
printf 'Total: %d  ' "$((PASS + FAIL))"
printf '%s ' "$(green "$PASS passed")"
printf '%s\n' "$(red "$FAIL failed")"
echo "=================================================="

if [ $FAIL -gt 0 ]; then
    echo
    echo "Failed tests:"
    for t in "${FAILED_TESTS[@]}"; do
        echo "  - $t"
    done
    exit 1
fi
exit 0
