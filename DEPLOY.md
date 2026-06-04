# Deployment & Setup Instructions

## Prerequisites
- **Go:** version 1.23 or later.
- **PostgreSQL:** version 13 or later.
- **Git:** for version control and submodule management.
- **GitHub Token:** A Personal Access Token (PAT) with `repo` permissions for autonomous PR management.

## Local Setup
1. **Clone the Repository:**
   ```bash
   git clone https://github.com/robertpelloni/enterprise_sales_bot.git
   cd enterprise_sales_bot
   ```
2. **Environment Variables:**
   Set up the following environment variables (or use a `.env` file):
   - `DATABASE_URL`: `postgres://user:password@localhost:5432/sales_bot?sslmode=disable`
   - `GITHUB_TOKEN`: Your GitHub PAT.
   - `GITHUB_REPOSITORY`: The `owner/repo` string for the main repository.
3. **Database Migrations:**
   Apply migrations using your preferred tool (e.g., `golang-migrate`):
   ```bash
   # Example using a tool that supports the migrations/ directory
   migrate -path migrations/ -database "$DATABASE_URL" up
   ```
   *Note: Ensure `000004_add_interaction_success.up.sql` is applied for the Self-Improving Prompts feature.*
4. **Initialize Submodules:**
   ```bash
   git submodule update --init --recursive
   ```

## Building the Application
Run the provided build script:
```batch
build.bat
```
This will run integrity tests and compile the binary to `bin/sales_bot.exe`.

## Self-Service Deployment Dashboard
The application includes a built-in dashboard for managing deployment tasks autonomously.
- **Sync Repository:** Triggers a fetch and merge from the remote origin and updates all submodules recursively, ensuring the bot is running the latest code.
- **Trigger Build:** Executes the project build process (`go build`) to recompile the system on the target environment.

### Automated Repository Synchronization
The bot can be configured to automatically sync with its repository using two methods:
1. **GitHub Webhooks:** Configure your repository to send push events to `http://<bot-ip>:8080/api/v1/webhook/github`. This will trigger an immediate sync and build.
2. **Background Polling:** Set the `DEPLOY_SYNC_INTERVAL` environment variable (e.g., `1h`, `15m`) to enable periodic background synchronization.

## Running the Application
Run the provided start script:
```batch
start.bat
```

## CI/CD
The project uses GitHub Actions for continuous integration and automated deployment:
- **CI/CD (`deploy.yml`):** A unified pipeline that manages testing, staging validation, and production deployment.
    - **Tests:** Runs unit and integration tests with a PostgreSQL service.
    - **Staging:** Automatically deploys to a staging environment (port 8081) on pull requests and runs smoke tests.
    - **Production:** Deploys to the production environment on pushes to `main` or version tags, gated by successful tests.

### Required Secrets
To enable automated deployment, ensure the following secrets are configured in GitHub:
- `DEPLOY_HOST`: The target server address.
- `DEPLOY_KEY`: SSH private key for server access.
- `DATABASE_URL`: Production PostgreSQL connection string.
