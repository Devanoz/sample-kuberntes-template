#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== Full Stack Deployment ==="
echo ""




echo "Step 1: Deploying Grafana stack..."
"$SCRIPT_DIR/apply-grafana.sh"

echo ""
echo "Step 2: Deploying application services..."
"$SCRIPT_DIR/deploy-services.sh"

echo ""
echo "=== Full Stack Deployment Complete ==="
echo ""
echo "Access points:"
echo "  Grafana:         http://localhost:30030"
echo "  Product Service: http://localhost:30009"
