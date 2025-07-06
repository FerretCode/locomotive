# locomotive

A Railway sidecar service for sending webhook events when new logs are received. Supports Discord, Datadog, Axiom, BetterStack and more!

## Configuration

Configuration is done through environment variables. See explanation and examples below.

## Webhook Log Format Examples

**These are examples of what the body would look like in the POST request done by locomotive**
<details>
<summary>Deploy Logs</summary>
<details>
<summary>Plaintext Deploy Logs</summary>

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
</details>

<details>
<summary>Structured JSON Deploy Logs</summary>

For Structured JSON Deploy Logs

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
</details>

<details>
<summary>Grafana Loki Plaintext Deploy Logs Example</summary>

Grafana Loki Plaintext Deploy Logs Example

```json
{
    "streams": [
        {
            "stream": {
                "deployment_id": "fb8172c8-a65d-48a4-9d1e-9d5ef986c9c3",
                "deployment_instance_id": "25dfeb9b-0097-4f91-820b-5dccc5009b1d",
                "project_id": "dce92382-c4e4-4923-bacd-3a5f7bcab337",
                "project_name": "Union Pacific Freight",
                "environment_id": "57d88ccb-8db9-4aef-957e-ecd94c41fdf8",
                "environment_name": "production",
                "service_id": "aa8ce660-dad0-4f7d-8921-46295d180c09",
                "service_name": "Dash 8",
                "severity": "error",
                "level": "error"
            },
            "values": [["1590182853000000000", "a plaintext message", {}]]
        }
    ]
}
```
</details>

<details>
<summary>Grafana Loki Structured Deploy Logs Example</summary>

Grafana Loki Structured Deploy Logs Example

```json
{
    "streams": [
        {
            "stream": {
                "deployment_id": "fb8172c8-a65d-48a4-9d1e-9d5ef986c9c3",
                "deployment_instance_id": "25dfeb9b-0097-4f91-820b-5dccc5009b1d",
                "project_id": "dce92382-c4e4-4923-bacd-3a5f7bcab337",
                "project_name": "Union Pacific Freight",
                "environment_id": "57d88ccb-8db9-4aef-957e-ecd94c41fdf8",
                "environment_name": "production",
                "service_id": "aa8ce660-dad0-4f7d-8921-46295d180c09",
                "service_name": "Dash 8",
                "severity": "error",
                "level": "error"
            },
            "values": [
                [
                    "1590182853000000000",
                    "hello, world",
                    {
                        "float": "10.51",
                        "number": "10",
                        "string_value": "hello world",
                        "user": "null"
                    }
                ]
            ]
        }
    ]
}
```
</details>
</details>

<details>
<summary>HTTP Logs</summary>

<details>
<summary>HTTP Logs For Generic Webhooks</summary>

```json
[
   {
      "_metadata":{
         "projectId":"bbd37ec6-1a5f-41bc-8461-910dffb30b1e",
         "projectName":"Railway",
         "environmentId":"681108cd-bbc6-49ac-a571-446cbbc2c6fe",
         "environmentName":"production",
         "serviceId":"f14f6e00-4d4e-448c-b526-79c358fc6ac0",
         "serviceName":"Frontend Railpack",
         "deploymentId":"ed5c3ddf-c333-4407-858d-58e6c5765066"
      },
      "clientUa":"Mozilla/5.0 (Macintosh; Intel Mac OS X 15_5) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.4 Safari/605.1.15",
      "downstreamProto":"HTTP/2.0",
      "edgeRegion":"us-east4-eqdc4a",
      "host":"railway.com",
      "httpStatus":200,
      "method":"GET",
      "path":"/dashboard",
      "requestId":"SMy7Drs-RcGXiaSg4a9AQ",
      "responseDetails":"",
      "rxBytes":4176,
      "srcIp":"66.33.22.11",
      "totalDuration":477,
      "txBytes":22453,
      "upstreamAddress":"http://[fd12:74d7:7e85:0:1000:34:be32:e1aa]:8080",
      "upstreamProto":"HTTP/1.1",
      "upstreamRqDuration":420,
      "message":"/dashboard",
      "timestamp":"2020-05-22T21:27:33Z",
      "time":"2020-05-22T21:27:33Z",
      "_time":"2020-05-22T21:27:33Z",
      "ts":"2020-05-22T21:27:33Z",
      "datetime":"2020-05-22T21:27:33Z",
      "dt":"2020-05-22T21:27:33Z"
   }
]
```

</details>

<details>
<summary>HTTP Logs For Loki</summary>

```json
{
   "streams":[
      {
         "stream":{
            "project_name":"Railway",
            "environment_id":"5cd7a403-45d9-4303-9de4-71bcfc7d2bf2",
            "environment_name":"production",
            "service_id":"3100de87-d044-4991-9c18-7a23e49c3927",
            "service_name":"Frontend Railpack",
            "deployment_id":"7d7426b1-0bd6-4b5e-8193-7f7a67160798",
            "project_id":"8ab20430-761a-4d60-9b39-772a514d928a"
         },
         "values":[
            [
               "1590182853000000000",
               "/dashboard",
               {
                  "clientUa":"Mozilla/5.0 (Macintosh; Intel Mac OS X 15_5) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.4 Safari/605.1.15",
                  "downstreamProto":"HTTP/2.0",
                  "edgeRegion":"us-east4-eqdc4a",
                  "host":"railway.com",
                  "httpStatus":404,
                  "method":"GET",
                  "requestId":"SMy7Drs-RcGXiaSg4a9AQ",
                  "responseDetails":"",
                  "rxBytes":4302,
                  "srcIp":"66.33.22.11",
                  "totalDuration":242,
                  "txBytes":19,
                  "upstreamAddress":"http://[fd12:74d7:7e85:0:1000:34:be32:e1aa]:8080",
                  "upstreamProto":"HTTP/1.1",
                  "upstreamRqDuration":185
               }
            ]
         ]
      }
   ]
}
```

</details>
</details>

**Notes:**

-   Metadata is gathered approximately every 10 to 20 minutes. If a project/service/environment name has changed, the name in the metadata will not be correct until the locomotive refreshes its metadata.

-   The body will always be a JSON array containing one or more log objects.

-   Various common timestamp attributes are included in every log object to increase compatibility with external logging services. [ref 1](https://axiom.co/docs/send-data/ingest#timestamp-field), [ref 2](https://betterstack.com/docs/logs/http-rest-api/#sending-timestamps)

-   The default `Content-Type` for these POST requests is set to `application/json`

-   Structured log attributes sent to Grafana Loki must always be a string

- The root attributes in the HTTP logs are subject to change as Railway adds or removes information.

All variables:

-   `RAILWAY_API_KEY` - Your Railway API key.

    -   Project level keys do not work.

-   `ENVIRONMENT_ID` - The environment ID your service is in.

    -   Auto-filled to the current environment ID.

-   `TRAIN` - The ID of the service you want to monitor.

    -   Supports multiple service Ids, separated with a comma.

-   `DISCORD_WEBHOOK_URL` - The Discord webhook URL to send logs to.

    -   Optional.

-   `DISCORD_PRETTY_JSON` - Pretty print the RAW JSON object in Discord embeds.

-   `SLACK_WEBHOOK_URL` - The Slack webhook URL to send logs to.

    -   Optional.

-   `SLACK_PRETTY_JSON` - Pretty print the RAW JSON object in Slack embeds.

-   `SLACK_TAGS` - Tags to add to the Slack message.

    -   Supports multiple tags, separated with a comma.
    -   Optional.

-   `LOKI_INGEST_URL` - The Loki ingest URL to send logs to.

    -   Example with no authentication: `https://loki-instance.up.railway.app/loki/api/v1/push`
    -   Example with username/password authentication: `https://user:pass@loki-instance.up.railway.app/loki/api/v1/push`
    -   Optional.

-   `INGEST_URL` - The URL to send a generic request to.

    -   Example for Datadog: `INGEST_URL=https://http-intake.logs.datadoghq.com/api/v2/logs`
    -   Example for Axiom: `INGEST_URL=https://api.axiom.co/v1/datasets/DATASET_NAME/ingest`
    -   Example for BetterStack: `INGEST_URL=https://in.logs.betterstack.com`
    -   Optional.

-   `ADDITIONAL_HEADERS` - Any additional headers to be sent with the generic request.

    -   Useful for auth. In the format of a cookie. meaning each key value pair is split by a semi-colon and each key value is split by an equals sign.
    -   Example for Datadog: `ADDITIONAL_HEADERS=DD-API-KEY=<DD_API_KEY>;DD-APPLICATION-KEY=<DD_APP_KEY>`
    -   Example for Axiom/BetterStack: `ADDITIONAL_HEADERS=Authorization=Bearer API_TOKEN`

-   `REPORT_STATUS_EVERY` - Reports the status of the locomotive every 5 seconds.

    -   Default: 5s.
    -   Format must be in the Golang time.DurationParse format
        -   E.g. 10h, 5h, 10m, 5m 5s

-   `ENABLE_HTTP_LOGS` - Enable shipping HTTP logs.

    -   Default: false.
    -   If enabled, locomotive will send logs to the HTTP endpoint specified in the `INGEST_URL` and `LOKI_INGEST_URL` environment variables.
    -   Discord and Slack will not receive HTTP logs.
    -   Optional.

-   `ENABLE_DEPLOY_LOGS` - Enable shipping deploy logs.

    -   Default: true.
    -   If enabled, locomotive will send logs to all the configured outputs.
    -   Optional.

-   `LOGS_FILTER` - Global log filter.

    -   Either ALL, INFO, ERROR, WARN or any custom combination of severity / level.
    -   Accepts multiple values, separated with a comma.
    -   Defaults to allowing all log levels.
    -   Optional.

-   `LOGS_FILTER_DISCORD` - Discord specific log filter.

    -   Same options and behavior as the global log filter.

-   `LOGS_FILTER_SLACK` - Slack specific log filter.

    -   Same options and behavior as the global log filter.

-   `LOGS_FILTER_LOKI` - Slack specific log filter.

    -   Same options and behavior as the global log filter.

-   `LOGS_FILTER_WEBHOOK` - Ingest URL specific log filter.

    -   Same options and behavior as the global log filter.

## Log Filtering

You can filter logs by severity level and content using the following environment variables:

### Level Filters

-   `LOGS_FILTER`: Global level filter applied to all outputs
-   `LOGS_FILTER_DISCORD`: Level filter applied to Discord output
-   `LOGS_FILTER_SLACK`: Level filter applied to Slack output
-   `LOGS_FILTER_LOKI`: Level filter applied to Loki output
-   `LOGS_FILTER_WEBHOOK`: Level filter applied to webhook output

Level filter options: ALL, INFO, ERROR, WARN, or any custom combination of severity / level.

### Content Filters

-   `LOGS_CONTENT_FILTER`: Global content filter applied to all outputs
-   `LOGS_CONTENT_FILTER_DISCORD`: Content filter applied to Discord output
-   `LOGS_CONTENT_FILTER_SLACK`: Content filter applied to Slack output
-   `LOGS_CONTENT_FILTER_LOKI`: Content filter applied to Loki output
-   `LOGS_CONTENT_FILTER_WEBHOOK`: Content filter applied to webhook output

Content filters support regular expressions or plain text searches.

Examples:

-   "hello"
-   "[A-za-z]ello"
