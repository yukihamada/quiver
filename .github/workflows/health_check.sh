#!/bin/bash
set -e

echo "Checking health endpoints..."

# Gateway health
curl -f http://localhost:8080/health || echo "Gateway health check failed"

# Aggregator health  
curl -f http://localhost:8081/health || echo "Aggregator health check failed"

echo "Health checks complete"