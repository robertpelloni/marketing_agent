# TormentNexus Marketing Agent Kubernetes Deployment

These manifests provide a basic deployment structure for running the marketing agent within a Kubernetes cluster.

## Deployment Steps

1. **Create Namespace**
   ```bash
   kubectl apply -f namespace.yaml
   ```

2. **Configure Secrets**
   Before deploying, edit `secret.yaml` to include your actual API keys, database credentials, and passwords. It is highly recommended to manage these secrets externally using tools like External-Secrets operator or HashiCorp Vault in a production environment.

3. **Apply Configurations**
   ```bash
   kubectl apply -f configmap.yaml
   kubectl apply -f secret.yaml
   ```

4. **Deploy Application**
   ```bash
   kubectl apply -f deployment.yaml
   kubectl apply -f service.yaml
   ```

## Note on Scaling

By default, `replicas` is set to `1`. The current worker architecture relies heavily on PostgreSQL polling for background jobs. While PostgreSQL connection limits are configured, concurrent execution of multiple marketing agent pods could result in duplicate messaging unless a distributed lock (e.g., Redis or PG advisory locks) is fully implemented for the worker queues. Proceed with caution when scaling beyond 1 replica.
