#!/bin/bash
# check-coverage.sh — verify per-package coverage floors
# Fails with exit code 1 if any package falls below its minimum.
#
# Uses parallel arrays instead of `declare -A` so the script works under
# macOS system bash 3.2 (which lacks associative arrays). Keep the two
# arrays index-aligned.

set -e

declare -a PKGS=(
    "github.com/lthiagol/lazybrew/cmd/lazybrew"
    "github.com/lthiagol/lazybrew/internal/app"
    "github.com/lthiagol/lazybrew/internal/brew"
    "github.com/lthiagol/lazybrew/internal/config"
    "github.com/lthiagol/lazybrew/internal/gui"
    "github.com/lthiagol/lazybrew/internal/gui/anim"
    "github.com/lthiagol/lazybrew/internal/gui/modal"
    "github.com/lthiagol/lazybrew/internal/gui/presentation"
    "github.com/lthiagol/lazybrew/internal/gui/style"
    "github.com/lthiagol/lazybrew/internal/gui/task"
    "github.com/lthiagol/lazybrew/internal/gui/testutil"
    "github.com/lthiagol/lazybrew/internal/gui/textutil"
    "github.com/lthiagol/lazybrew/internal/logtail"
)
declare -a FLOORS=(42 68 90 91 90 86 90 98 90 94 43 98 91)

# Floors are the measured coverage rounded down, less 2 points of headroom
# for small drift (Q-cover-floors, 2026-11-20). Every package with tests has
# one; a new package gets one in the slice that adds it. internal/gui/flows is
# tests only (no statements to measure), so it has none. Raise only when
# tests land; do not lower without a decision log entry.

FAILED=0
for i in "${!PKGS[@]}"; do
    pkg="${PKGS[$i]}"
    floor="${FLOORS[$i]}"
    output=$(go test -cover -count=1 "$pkg" 2>/dev/null)
    cov=$(echo "$output" | sed -n 's/.*coverage: \([0-9.]*\)%.*/\1/p')
    if [ -z "$cov" ]; then
        echo "FAIL: could not get coverage for $pkg"
        FAILED=1
        continue
    fi
    int_cov=$(echo "$cov" | cut -d. -f1)
    if [ "$int_cov" -lt "$floor" ] 2>/dev/null; then
        echo "FAIL: $pkg coverage ${cov}% < minimum ${floor}%"
        FAILED=1
    else
        echo "OK:   $pkg coverage ${cov}% >= ${floor}%"
    fi
done

if [ "$FAILED" -eq 0 ]; then
    echo "All coverage floors met"
fi
exit "$FAILED"
