# locomotive

A railway sidecar service for sending webhook events when new logs are received

# Configuration

Configuration is done through environment variables.

All variables:

- RAILWAY_API_KEY - Your Railway API key
- RAILWAY_PROJECT_ID - Your project ID (automatically set by Railway)
- TRAIN - The ID of the service you want to monitor (found in URL on the dashboard)
- LOGS_FILTER - Which logs you want sent (either ALL, INFO, or ERROR)
- POLLING_RATE_SECONDS - The number of seconds between checking for logs
- DISCORD_WEBHOOK_URL - The Discord webhook URL to send logs to
