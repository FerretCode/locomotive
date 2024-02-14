# locomotive

A railway sidecar service for sending webhook events when new logs are received

# Configuration

Configuration is done through environment variables.

# Log Format

```json
{
  "Message": "log message",
  "Severity": "INFO",
  "Embed": true
}
```

All variables:

- RAILWAY_API_KEY - Your Railway API key
- ENVIRONMENT_ID - The environment ID your service is in (Ctrl+K -> Copy environment ID)
- TRAIN - The ID of the service you want to monitor (found in URL on the dashboard)
- LOGS_FILTER - Which logs you want sent (either ALL, INFO, ERROR, WARN or any custom severity)
- DISCORD_WEBHOOK_URL - The Discord webhook URL to send logs to (optional)
- INGEST_URL - The URL to send a generic request to
- ADDITIONAL_HEADERS - Any additional headers to be sent with the generic request (like auth). In the format of a cookie (e.g. Authorization=key;Content-Type=application/json)
