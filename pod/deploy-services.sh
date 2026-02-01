#!/bin/bash

set -e

NAMESPACE="staging"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "Deploying Services to Kubernetes"
echo ""

echo "Building Docker images..."

echo "  Building order-service..."
docker build --platform linux/amd64 -t order-service:latest "$PROJECT_ROOT/order service"

echo "  Building product-service..."
docker build --platform linux/amd64 -t product-service:latest "$PROJECT_ROOT/product service"

echo ""
echo "Applying Kubernetes manifests..."

kubectl apply -f "$SCRIPT_DIR/order-service.yaml"
kubectl apply -f "$SCRIPT_DIR/product-service.yaml"

echo ""
echo "Restarting deployments..."

kubectl rollout restart deployment/order-service -n $NAMESPACE
kubectl rollout restart deployment/product-service -n $NAMESPACE

echo ""
echo "Waiting for rollout to complete..."

kubectl rollout status deployment/order-service -n $NAMESPACE --timeout=120s
kubectl rollout status deployment/product-service -n $NAMESPACE --timeout=120s

echo ""
echo "=== Deployment Complete ==="
echo ""
echo "Services:"
kubectl get pods -n $NAMESPACE -l 'app in (order-service, product-service)'
echo ""
echo "Access product-service at: http://localhost:30009"
echo ""
echo "Test commands:"
echo "  curl http://localhost:30009/products"
echo "  curl -X POST http://localhost:30009/products/123/orders -H 'Content-Type: application/json' -d '{\"quantity\": 5}'"
