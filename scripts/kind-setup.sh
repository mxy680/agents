#!/bin/bash
# kind-setup.sh — Create a local Kubernetes cluster for development
set -e

CLUSTER_NAME="${KIND_CLUSTER_NAME:-agents-dev}"
NAMESPACE="agents"

echo "Creating kind cluster: $CLUSTER_NAME"
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    echo "Cluster $CLUSTER_NAME already exists"
else
    kind create cluster --name "$CLUSTER_NAME"
fi

echo "Creating namespace: $NAMESPACE"
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

echo "Creating ANTHROPIC_API_KEY secret"
if [ -z "$ANTHROPIC_API_KEY" ]; then
    echo "Warning: ANTHROPIC_API_KEY not set, creating placeholder secret"
    kubectl create secret generic anthropic-api-key \
        --namespace="$NAMESPACE" \
        --from-literal=api-key=placeholder \
        --dry-run=client -o yaml | kubectl apply -f -
else
    kubectl create secret generic anthropic-api-key \
        --namespace="$NAMESPACE" \
        --from-literal=api-key="$ANTHROPIC_API_KEY" \
        --dry-run=client -o yaml | kubectl apply -f -
fi

echo ""
echo "Kind cluster '$CLUSTER_NAME' is ready."
echo "  kubectl config current-context: $(kubectl config current-context)"
echo "  Namespace: $NAMESPACE"
