#!/usr/bin/env bash
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

GO_MINOR=24
GO_PATCH=2

pass=0
fail=0

check() {
  local dep=$1
  local category=$2
  local found=$3
  local version=$4

  if [ "$found" = true ]; then
    echo -e "  ${GREEN}✓${NC} $dep ($version)"
    pass=$((pass + 1))
  else
    echo -e "  ${RED}✗${NC} $dep — not found"
    fail=$((fail + 1))
  fi
}

check_cmd() {
  local dep=$1
  local category=$2
  if command -v "$dep" &>/dev/null; then
    local ver
    ver=$("$dep" version 2>/dev/null | head -1 || "$dep" --version 2>/dev/null | head -1 || echo "found")
    check "$dep" "$category" true "$ver"
  else
    check "$dep" "$category" false ""
  fi
}

echo "── Build Dependencies ──"
echo ""

check_cmd "go" "build"
check_cmd "git" "build"
check_cmd "make" "build"

echo ""
echo "── Run Dependencies ──"
echo ""

check_brew() {
  if command -v brew &>/dev/null; then
    local ver
    ver=$(brew --version 2>/dev/null | head -1 || echo "found")
    check "brew (Homebrew)" "run" true "$ver"
  else
    local candidates=(
      "/opt/homebrew/bin/brew"
      "/usr/local/bin/brew"
      "/home/linuxbrew/.linuxbrew/bin/brew"
    )
    local found=false
    for c in "${candidates[@]}"; do
      if [ -x "$c" ]; then
        local ver
        ver=$("$c" --version 2>/dev/null | head -1 || echo "found at $c")
        check "brew (Homebrew)" "run" true "$ver"
        found=true
        break
      fi
    done
    if [ "$found" = false ]; then
      check "brew (Homebrew)" "run" false ""
    fi
  fi
}
check_brew

echo ""
echo "── Development Tools ──"
echo ""
check_cmd "gofmt" "dev"
check_cmd "goreleaser" "dev"

echo ""
total=$((pass + fail))
echo ""
echo "─── Result ───"
echo -e "  ${GREEN}${pass} met${NC}, ${RED}${fail} unmet${NC} (${total} total)"
echo ""

if [ "$fail" -gt 0 ]; then
  exit 1
fi
