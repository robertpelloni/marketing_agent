# TormentNexus Marketing Agent Helm Chart

This Helm chart provides a configurable, one-command deployment method for the TormentNexus Marketing Agent onto Kubernetes clusters.

## Prerequisites

- Kubernetes cluster 1.20+
- Helm 3.0+

## Installing the Chart

To install the chart with the release name `marketing-agent` in the `tormentnexus` namespace:

```bash
kubectl create namespace tormentnexus
helm install marketing-agent ./deploy/helm/marketing-agent --namespace tormentnexus
```

## Configuration

The `values.yaml` file exposes the configurable parameters of the marketing agent. You can override these during installation:

```bash
helm install marketing-agent ./deploy/helm/marketing-agent \
  --namespace tormentnexus \
  --set secrets.databaseUrl="postgres://user:password@host/db" \
  --set secrets.hermesApiKey="your-key"
```

## Important Considerations

- **Secrets Management:** The generated `secret.yaml` handles base64 encoding from stringData automatically. For production, it is highly recommended to manage secrets externally (e.g., using External-Secrets operator) instead of passing them directly through Helm `values.yaml`.
- **Scaling:** The worker architecture operates sequentially to avoid state-locking issues. Keep `replicaCount` at `1` unless horizontal scaling (via Postgres advisory locks or Redis) is fully implemented in the background workers.
