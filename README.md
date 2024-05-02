# locomotive

A Railway sidecar service for sending webhook events when new logs are received. Supports Discord, Axiom, BetterStack and more! 

## Configuration

Configuration is done through environment variables. See explanation and examples below.

## Generic Webhook Log Formats

**These are examples of what the body would look like in the POST request done by locomotive**

For Plaintext logs
```json
[
  {
    "_metadata": {
      "deploymentId": "577e2cf2-a1fc-4e0f-b352-2780bca73a94",
      "deploymentInstanceId": "bbdc4e76-7600-415f-9f6b-e425311cec51",
      "environmentId": "b5ce7ab5-96f1-4fa3-929b-fc883f89cbd1",
      "environmentName": "production",
      "projectId": "8a6502bf-6479-440c-a14f-78ecd52abf09",
      "projectName": "Railyard",
      "serviceId": "24335e07-e68b-498f-bc9e-1b9146436867",
      "serviceName": "Autorack"
    },
    "message": "Hello, World!",
    "level": "info",
    "severity": "info",
    "time": "2020-05-22T21:27:33Z",
    "_time": "2020-05-22T21:27:33Z",
    "dt": "2020-05-22T21:27:33Z",
    "datetime": "2020-05-22T21:27:33Z",
    "ts": "2020-05-22T21:27:33Z",
    "timestamp": "2020-05-22T21:27:33Z"
  }
]
```

For Structured JSON logs
```json
[
  {
    "_metadata": {
      "deploymentId": "5b7c81b35-1578-4eb8-8498-44f4f517b263",
      "deploymentInstanceId": "46cda6d4-f76c-45cb-8642-c7265949e497",
      "environmentId": "b5ce7ab5-96f1-4fa3-929b-fc883f89cbd1",
      "environmentName": "production",
      "projectId": "8a6502bf-6479-440c-a14f-78ecd52abf09",
      "projectName": "Railyard",
      "serviceId": "55b1755f-f2c6-4f24-8d51-0ed3754b253e",
      "serviceName": "Superliner"
    },
    "level": "info",
    "severity": "info",
    "message": "Hello, World!",
    "example_string": "foo bar",
    "example_int": 12345678,
    "example_float": 1.2345678,
    "example_int_slice": [123, 456, 789],
    "example_string_slice": ["hello", "world"],
    "example_group": {
      "example_grouped_int": 12345678,
      "example_grouped_string": "Hello, World!"
    },
    "time": "2020-05-22T21:27:33Z",
    "_time": "2020-05-22T21:27:33Z",
    "dt": "2020-05-22T21:27:33Z",
    "datetime": "2020-05-22T21:27:33Z",
    "ts": "2020-05-22T21:27:33Z",
    "timestamp": "2020-05-22T21:27:33Z"
  }
]

```

**Notes:**
- Metadata is gathered once when the locomotive starts, If a project/service/environment name has changed, the name in the metadata will not be correct until the locomotive is restarted.

- The body will always be a JSON array containing one or more log objects.

- Various common timestamp attributes are included in every log object to increase compatibility with external logging services. [ref 1](https://axiom.co/docs/send-data/ingest#timestamp-field), [ref 2](https://betterstack.com/docs/logs/http-rest-api/#sending-timestamps)

- The default `Content-Type` for these POST requests is set to `application/json`

All variables:

- `RAILWAY_API_KEY` - Your Railway API key.
  - Project level keys do not work.

- `ENVIRONMENT_ID` - The environment ID your service is in.
  - Auto-filled to the current environment ID.

- `TRAIN` - The ID of the service you want to monitor.
  - Supports multiple service Ids, separated with a comma.

- `DISCORD_WEBHOOK_URL` - The Discord webhook URL to send logs to.
  - Optional.

- `DISCORD_PRETTY_JSON` - Pretty print the RAW JSON object in Discord embeds.

- `INGEST_URL` - The URL to send a generic request to.
  - Example for Axiom: `INGEST_URL=https://api.axiom.co/v1/datasets/DATASET_NAME/ingest`
  - Example for BetterStack: `INGEST_URL=https://in.logs.betterstack.com`
  - Optional.

- `ADDITIONAL_HEADERS` - Any additional headers to be sent with the generic request.
  - Useful for auth. In the format of a cookie. meaning each key value pair is split by a semi-colon and each key value is split by an equals sign.
  - Example for Axiom/BetterStack: `ADDITIONAL_HEADERS=Authorization=Bearer API_TOKEN`

- `REPORT_STATUS_EVERY` - Reports the status of the locomotive every 5 seconds.
  - Default: 5s.
  - Format must be in the Golang time.DurationParse format
      - E.g. 10h, 5h, 10m, 5m 5s

- `MAX_ERR_ACCUMULATIONS` - The maximum number of errors to occur before exiting.
  - Default: 10.

- `LOGS_FILTER` - Global log filter.
  - Either ALL, INFO, ERROR, WARN or any custom combination of severity / level.
  - Accepts multiple values, separated with a comma.
  - Defaults to allowing all log levels.
  - Optional.

- `LOGS_FILTER_DISCORD` - Discord specific log filter.
  - Same options and behavior as the global log filter.

- `LOGS_FILTER_WEBHOOK` - Ingest URL specific log filter.
  - Same options and behavior as the global log filter.
