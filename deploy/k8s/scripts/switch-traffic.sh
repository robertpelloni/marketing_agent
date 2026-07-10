#!/bin/bash

# Usage: ./switch-traffic.sh [blue|green]

TARGET_VERSION=$1

echo "Switching traffic to $TARGET_VERSION deployment..."

# Scale up target
kubectl scale deployment marketing-agent-$TARGET_VERSION --replicas=1 -n tormentnexus

# Wait for rollout
kubectl rollout status deployment/marketing-agent-$TARGET_VERSION -n tormentnexus

# Update Service to point to the new version
kubectl patch service marketing-agent-service -n tormentnexus -p "{\"spec\":{\"selector\":{\"version\":\"$TARGET_VERSION\"}}}"

echo "Traffic switched to $TARGET_VERSION."

# Scale down the old deployment
if [[ "$TARGET_VERSION" == "blue" ]]; then
    kubectl scale deployment marketing-agent-green --replicas=0 -n tormentnexus
else
    kubectl scale deployment marketing-agent-blue --replicas=0 -n tormentnexus
fi

echo "Deployment complete."
