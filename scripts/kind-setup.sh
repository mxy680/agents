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

echo "Creating CLAUDE_CODE_OAUTH_TOKEN secret"
if [ -z "$CLAUDE_CODE_OAUTH_TOKEN" ]; then
    echo "Warning: CLAUDE_CODE_OAUTH_TOKEN not set, creating placeholder secret"
    kubectl create secret generic claude-session-token \
        --namespace="$NAMESPACE" \
        --from-literal=session-token=placeholder \
        --dry-run=client -o yaml | kubectl apply -f -
else
    kubectl create secret generic claude-session-token \
        --namespace="$NAMESPACE" \
        --from-literal=session-token="$CLAUDE_CODE_OAUTH_TOKEN" \
        --dry-run=client -o yaml | kubectl apply -f -
fi

echo ""
echo "Kind cluster '$CLUSTER_NAME' is ready."
echo "  kubectl config current-context: $(kubectl config current-context)"
echo "  Namespace: $NAMESPACE"
