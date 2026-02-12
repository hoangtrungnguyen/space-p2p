#!/bin/bash
set -e

# TDD Watcher Script
# Usage: ./test_watcher.sh [package]
# Returns 0 if tests pass, non-zero if they fail.

TARGET=${1:-./...}

echo "========================================"
echo "👀 TDD Watcher: Running tests for $TARGET"
echo "========================================"

if command -v gotestsum &> /dev/null; then
    gotestsum --format short-verbose -- $TARGET
else
    go test -v $TARGET
fi

EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ Tests Passed!"
else
    echo "❌ Tests Failed!"
fi

exit $EXIT_CODE
