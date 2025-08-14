#!/bin/bash
set -e

# Vegeta load test script
DURATION="30s"
RATE="2/s"  # 2 requests per second (2 workers)
TARGET_URL="http://localhost:8080/generate"

# Create target file
cat > targets.txt << EOF
POST ${TARGET_URL}
Content-Type: application/json
@payload1.json

POST ${TARGET_URL}
Content-Type: application/json
@payload2.json
EOF

# Create payloads
echo '{"prompt":"What is 2+2?","model":"llama2","token":"vegeta-1"}' > payload1.json
echo '{"prompt":"Name the days of the week","model":"llama2","token":"vegeta-2"}' > payload2.json

# Run attack
echo "Running Vegeta attack: ${RATE} for ${DURATION}"
vegeta attack -duration=${DURATION} -rate=${RATE} -targets=targets.txt | \
  vegeta report -type=text

# Generate detailed report
vegeta attack -duration=${DURATION} -rate=${RATE} -targets=targets.txt | \
  vegeta encode | \
  vegeta report -type=json > vegeta_report.json

# Check thresholds
echo "Checking CI thresholds..."
p95=$(cat vegeta_report.json | jq '.latencies.p95' | cut -d. -f1)
error_rate=$(cat vegeta_report.json | jq '.errors | length / .requests * 100')

echo "P95 latency: ${p95}ns"
echo "Error rate: ${error_rate}%"

# Convert nanoseconds to milliseconds
p95_ms=$((p95 / 1000000))

if [ $p95_ms -gt 2500 ]; then
  echo "FAIL: P95 latency ${p95_ms}ms exceeds 2500ms"
  exit 1
fi

if (( $(echo "$error_rate > 1" | bc -l) )); then
  echo "FAIL: Error rate ${error_rate}% exceeds 1%"
  exit 1
fi

echo "PASS: All thresholds met"

# Cleanup
rm -f targets.txt payload*.json