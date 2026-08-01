#!/bin/bash
# check-coverage.sh — verify per-package coverage floors
# Fails with exit code 1 if any package falls below its minimum.
#
# Uses parallel arrays instead of `declare -A` so the script works under
# macOS system bash 3.2 (which lacks associative arrays). Keep the two
# arrays index-aligned.

set -e

declare -a PKGS=(
    "github.com/lthiagol/lazybrew/internal/brew"
    "github.com/lthiagol/lazybrew/internal/gui/presentation"
    "github.com/lthiagol/lazybrew/internal/gui"
    "github.com/lthiagol/lazybrew/internal/gui/modal"
)
declare -a FLOORS=(62 91 36 41)

# Floors are set to current measured coverage rounded down.
# Raise only when tests land; do not lower without a decision log entry.

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
