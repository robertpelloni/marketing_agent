const https = require('https');

const JULES_API_KEY = process.env.JULES_API_KEY;
const SOURCE = "sources/github/robertpelloni/enterprise_sales_bot";

const subtasks = [
  {
    title: "Live CRM Integration Test Suite",
    prompt: "Implement a comprehensive suite of integration tests in Go that verify boundary cases, rate limiting, and data integrity for CRM connectivity. Use the existing CRMClient interface and expand RestCRMClient to support more detailed error reporting. Target: internal/crm package."
  },
  {
    title: "Staging Environment Orchestration",
    prompt: "Orchestrate a staging environment for the Enterprise Sales Bot using Docker Compose. Ensure it includes automatic database migrations for a separate staging DB, live secret injection for CRM_BASE_URL and CRM_API_KEY, and a health check sequence that verifies connectivity before starting workers."
  },
  {
    title: "Production-grade Observability",
    prompt: "Migrate the entire application from log.Printf to structured JSON logging using the Go 'slog' package. Implement a centralized logger configuration and add basic Prometheus metrics to track background worker success/failure rates and processing latency."
  }
];

async function spawnSubtask(task) {
  const data = JSON.stringify({
    title: task.title,
    sourceContext: {
      source: SOURCE,
      githubRepoContext: {
        startingBranch: "main"
      }
    },
    prompt: task.prompt,
    automationMode: "AUTO_CREATE_PR"
  });

  const options = {
    hostname: 'jules.googleapis.com',
    port: 443,
    path: '/v1alpha/sessions',
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Goog-Api-Key': JULES_API_KEY,
      'Content-Length': Buffer.byteLength(data)
    }
  };

  return new Promise((resolve, reject) => {
    const req = https.request(options, (res) => {
      let body = '';
      res.on('data', (chunk) => body += chunk);
      res.on('end', () => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(JSON.parse(body));
        } else {
          reject(new Error(`Failed to spawn subtask: ${res.statusCode} ${body}`));
        }
      });
    });

    req.on('error', (e) => reject(e));
    req.write(data);
    req.end();
  });
}

async function main() {
  console.log("Spawning modular subagents...");
  for (const task of subtasks) {
    try {
      const session = await spawnSubtask(task);
      console.log(`Spawned: ${task.title} (Session ID: ${session.id})`);
    } catch (err) {
      console.error(err.message);
    }
  }
}

main();
