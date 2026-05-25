# Deployment & Setup Instructions

## Prerequisites
- **Go:** version 1.24 or later.
- **PostgreSQL:** version 13 or later.
- **Git:** for version control and submodule management.

## Local Setup
1. **Clone the Repository:**
   ```bash
   git clone https://github.com/robertpelloni/enterprise_sales_bot.git
   cd enterprise_sales_bot
   ```
2. **Environment Variables:**
   Set up the following environment variables (or use a `.env` file):
   - `DATABASE_URL`: `postgres://user:password@localhost:5432/sales_bot?sslmode=disable`
3. **Database Migrations:**
   Apply migrations using your preferred tool (e.g., `golang-migrate`):
   ```bash
   # Example using a tool that supports the migrations/ directory
   migrate -path migrations/ -database "$DATABASE_URL" up
   ```
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

## Running the Application
Run the provided start script:
```batch
start.bat
```

## CI/CD
The project uses GitHub Actions for continuous integration and automated deployment:
- **CI (`ci.yml`):** Runs on every push and PR to `main`. It verifies submodule integrity, version consistency between `VERSION` and `VERSION.md`, and executes all integrity and project tests.
- **CD (`deploy.yml`):** Automatically triggers on version tags (e.g., `v0.2.0`). It builds the bot and executes provisioning logic to update the target environment.

### Required Secrets
To enable automated deployment, ensure the following secrets are configured in GitHub:
- `DEPLOY_HOST`: The target server address.
- `DEPLOY_KEY`: SSH private key for server access.
- `DATABASE_URL`: Production PostgreSQL connection string.
