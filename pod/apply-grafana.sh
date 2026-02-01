#!/bin/bash

set -e

NAMESPACE="staging"

echo "Creating namespace if not exists..."
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

echo "Creating ConfigMaps from config files..."

# Delete existing ConfigMaps (ignore errors if they don't exist)
kubectl delete configmap loki-config -n $NAMESPACE 2>/dev/null || true
kubectl delete configmap tempo-config -n $NAMESPACE 2>/dev/null || true
kubectl delete configmap mimir-config -n $NAMESPACE 2>/dev/null || true
kubectl delete configmap alloy-config -n $NAMESPACE 2>/dev/null || true

# Create ConfigMaps from files
kubectl create configmap loki-config \
  --from-file=loki.yaml=../loki/loki.yaml \
  -n $NAMESPACE

kubectl create configmap tempo-config \
  --from-file=tempo.yaml=../tempo/tempo.yaml \
  -n $NAMESPACE

kubectl create configmap mimir-config \
  --from-file=mimir.yaml=../mimir/mimir.yaml \
  -n $NAMESPACE

kubectl create configmap alloy-config \
  --from-file=config.alloy=../alloy/config.alloy \
  --from-file=endpoints.json=../alloy/endpoints.json \
  -n $NAMESPACE

echo "Applying Grafana stack manifest..."
kubectl apply -f grafana.yaml

echo ""
echo "Grafana stack deployed!"
echo "Access Grafana at: http://localhost:30030"
echo ""
echo "To check status:"
echo "  kubectl get pods -n $NAMESPACE"
echo "  kubectl get svc -n $NAMESPACE"