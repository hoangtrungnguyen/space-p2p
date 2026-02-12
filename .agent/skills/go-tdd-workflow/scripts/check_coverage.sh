#!/bin/bash
set -e

THRESHOLD=80
TARGET=${1:-./...}

echo "========================================"
echo "📊 Running Coverage Analysis for $TARGET"
echo "========================================"

# Run test with coverage profile
go test -coverprofile=coverage.out $TARGET

# Extract total coverage from go tool cover -func
# Output format: "total: (statements) 46.6%"
TOTAL_LINE=$(go tool cover -func=coverage.out | grep total:)
PERCENTAGE=$(echo $TOTAL_LINE | awk '{print $3}' | sed 's/%//')

echo "Total Coverage: $PERCENTAGE%"

# Convert to integer for comparison (e.g., 80.5 -> 80)
COVERAGE_INT=$(echo $PERCENTAGE | awk '{print int($1)}')

if [ "$COVERAGE_INT" -lt "$THRESHOLD" ]; then
    echo "❌ Coverage ($PERCENTAGE%) is BELOW threshold ($THRESHOLD%)"
    echo "Action Required: Add more tests. Check coverage.out for missed lines."
    exit 1
else
    echo "✅ Coverage ($PERCENTAGE%) MEETS threshold ($THRESHOLD%)"
    exit 0
fi
