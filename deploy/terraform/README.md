# TormentNexus Infrastructure as Code (Terraform)

This module provisions the foundational cloud infrastructure for the TormentNexus autonomous agent.

It provides:
- An isolated `aws_vpc` setup for the background workers.
- An `aws_db_instance` (PostgreSQL 16) designed to host the `pgvector` memory vault and pipeline state management.
- Hardened `aws_security_group` configurations exposing only necessary ports.

## Usage

1. Initialize Terraform:
   ```bash
   terraform init
   ```

2. Set your database credentials securely (do NOT commit these to git):
   ```bash
   export TF_VAR_db_username="admin"
   export TF_VAR_db_password="SuperSecretPassword123!"
   ```

3. Review the infrastructure plan:
   ```bash
   terraform plan
   ```

4. Apply the configuration:
   ```bash
   terraform apply
   ```

## Next Steps
Once the infrastructure is successfully built, use the DB endpoint provided in the `outputs` to configure your `DATABASE_URL` via the `.env` or Kubernetes ConfigMap prior to deploying the agent binary/pods.
