#!/bin/bash
# check-coverage.sh — verify per-package coverage floors
# Fails with exit code 1 if any package falls below its minimum.

set -e

declare -A FLOORS
FLOORS["github.com/thiago/lazybrew/internal/brew"]=65
FLOORS["github.com/thiago/lazybrew/internal/gui/presentation"]=90
FLOORS["github.com/thiago/lazybrew/internal/gui"]=30
FLOORS["github.com/thiago/lazybrew/internal/gui/modal"]=40

FAILED=0
for pkg in "${!FLOORS[@]}"; do
    floor=${FLOORS[$pkg]}
    output=$(go test -cover -count=1 "$pkg" 2>/dev/null)
    cov=$(echo "$output" | grep -oP 'coverage: \K[0-9.]+(?=% of statements)')
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
