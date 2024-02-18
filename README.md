# locomotive

A Railway sidecar service for sending webhook events when new logs are received. Supports Discord, Axiom, BetterStack and more! 

## Configuration

Configuration is done through environment variables. See explanation and examples below.

## Generic Webhook Log Formats

**These are examples of what the body would look like in the POST request done by locomotive**

For Plaintext logs
```json
{
   "_metadata":{
      "environmentId":"b5ce7ab5-96f1-4fa3-929b-fc883f89cbd1",
      "environmentName":"production",
      "projectId":"8a6502bf-6479-440c-a14f-78ecd52abf09",
      "projectName":"Railyard",
      "serviceId":"24335e07-e68b-498f-bc9e-1b9146436867",
      "serviceName":"Autorack"
   },
   "time":"2020-05-22T21:27:33Z",
   "severity":"info",
   "message":"Hello, World!"
}
```

For Structured JSON logs
```json
{
   "_metadata":{
      "environmentId":"b5ce7ab5-96f1-4fa3-929b-fc883f89cbd1",
      "environmentName":"production",
      "projectId":"8a6502bf-6479-440c-a14f-78ecd52abf09",
      "projectName":"Railyard",
      "serviceId":"55b1755f-f2c6-4f24-8d51-0ed3754b253e",
      "serviceName":"Superliner"
   },
   "time":"2020-05-22T21:27:33Z",
   "level":"INFO",
   "message":"Hello, World!",
   "example_attribute_string":"testing testing",
   "example_attribute_float":1.2345678,
   "example_attribute_int":12345678,
   "example_grouped_attribute":{
      "grouped_int":12345678,
      "grouped_string":"Hello, World!"
   }
}

```

**Notes:**
- Metadata is gathered once when the locomotive starts, If a project/service/environment name has changed, the name in the metadata will not be correct until the locomotive is restarted.

- The default `Content-Type` for these POST requests is set to `application/x-ndjson` but this can be overridden by adding `Content-Type=application/json` onto the end of the `ADDITIONAL_HEADERS` environment variable incase the service expects `application/json` instead

All variables:

- RAILWAY_API_KEY - Your Railway API key.
  - Project level keys do not work.
- ENVIRONMENT_ID - The environment ID your service is in.
  - Auto-filled to the current environment ID.
- TRAIN - The ID of the service you want to monitor.
  - Supports multiple service Ids, separated with a comma.
- LOGS_FILTER - Which logs you want sent.
  - Either ALL, INFO, ERROR, WARN or any custom severity.
- DISCORD_WEBHOOK_URL - The Discord webhook URL to send logs to.
  - Optional.
- INGEST_URL - The URL to send a generic request to.
  - Example for Axiom: `INGEST_URL=https://api.axiom.co/v1/datasets/DATASET_NAME/ingest`
  - Example for BetterStack: `INGEST_URL=https://in.logs.betterstack.com`
- ADDITIONAL_HEADERS - Any additional headers to be sent with the generic request.
  - Useful for auth. In the format of a cookie. meaning each key value pair is split by a semi-colon and each key value is split by an equals sign.
  - Example for Axiom/BetterStack: `ADDITIONAL_HEADERS=Authorization=Bearer API_TOKEN`
- REPORT_STATUS_EVERY - Reports the status of the locomotive every X log lines shipped.
  - Default: 50.
