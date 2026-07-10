#!/bin/bash

# Rolls back traffic to the other deployment version instantly if health checks fail

CURRENT_VERSION=$(kubectl get service marketing-agent-service -n tormentnexus -o=jsonpath='{.spec.selector.version}')

if [[ "$CURRENT_VERSION" == "blue" ]]; then
    TARGET_VERSION="green"
else
    TARGET_VERSION="blue"
fi

echo "Rolling back traffic to $TARGET_VERSION..."

kubectl scale deployment marketing-agent-$TARGET_VERSION --replicas=1 -n tormentnexus
kubectl rollout status deployment/marketing-agent-$TARGET_VERSION -n tormentnexus
kubectl patch service marketing-agent-service -n tormentnexus -p "{\"spec\":{\"selector\":{\"version\":\"$TARGET_VERSION\"}}}"

echo "Rollback to $TARGET_VERSION complete."
