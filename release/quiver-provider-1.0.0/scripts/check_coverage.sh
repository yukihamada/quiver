#!/bin/bash
set -e

echo "Checking test coverage..."

THRESHOLD=70
FAILED=0

for f in coverage/*.out; do
  if [ -f "$f" ]; then
    coverage=$(go tool cover -func="$f" | grep total | awk '{print $3}' | sed 's/%//')
    name=$(basename "$f" .out)
    
    echo -n "$name: $coverage% "
    
    if (( $(echo "$coverage >= $THRESHOLD" | bc -l) )); then
      echo "✓"
    else
      echo "✗ (below $THRESHOLD%)"
      FAILED=1
    fi
  fi
done

if [ $FAILED -eq 1 ]; then
  echo "Coverage check FAILED"
  exit 1
else
  echo "Coverage check PASSED"
fi