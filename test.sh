#!/bin/bash
# test.sh

# Run tests with coverage
go test -coverprofile=coverage.out

# Display coverage report
go tool cover -func=coverage.out

# Check if coverage is above threshold (e.g., 80%)
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
THRESHOLD=80

echo "Coverage: $COVERAGE%"
if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
    echo "Coverage is below threshold of $THRESHOLD%"
    exit 1
else
    echo "Coverage is above threshold of $THRESHOLD%"
fi

# Optionally open HTML coverage report
go tool cover -html=coverage.out