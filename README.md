## Railway Log Downloader

### Usage 

```bash
go mod download

export RAILWAY_ACCOUNT_TOKEN=<your-railway-account-token>

# Download logs for a deployment
go run . --deployment <deploymentId>

# Download logs for a service (requires both service and environment)
go run . --service <serviceId> --environment <environmentId>
```

This will download all the logs for the given deployment or service until one of the following conditions is met:

- You reach the deployment/service's creation date
- You reach the log retention limit (7/30/90 days [depending on your account's plan](https://docs.railway.com/reference/logging#log-retention))
- You hit the [API rate limit](https://docs.railway.com/reference/public-api#rate-limits)
- You cancel the operation (Ctrl/Cmd + C)

In any case, all the logs that have been downloaded will be saved to a file called `deployment-<deploymentId>.jsonl` or `service-<serviceId>.jsonl`.

### Configuration

The application can be configured using command-line flags or environment variables:

| Option         | Flag           | Environment Variable     | Description                                            | Required                              | Validation           |
|----------------|----------------|--------------------------|--------------------------------------------------------|---------------------------------------|----------------------|
| Deployment ID  | `--deployment` | `RAILWAY_DEPLOYMENT_ID`  | The deployment ID to download logs for                 | Yes (if not using `--service`)        | Must be a valid UUID |
| Service ID     | `--service`    | `RAILWAY_SERVICE_ID`     | The service ID to download logs for                    | Yes (if not using `--deployment`)     | Must be a valid UUID |
| Environment ID | `--environment`| `RAILWAY_ENVIRONMENT_ID` | The environment ID to download logs for                | Yes (required when using `--service`) | Must be a valid UUID |
| Filter         | `--filter`     | `RAILWAY_LOG_FILTER`     | Filter to apply to logs                                | No                                    | -                    |
| Overwrite File | `--overwrite`  | `RAILWAY_OVERWRITE_FILE` | Overwrite existing logs file                           | No                                    | Any boolean value    |
| Resume         | `--resume`     | `RAILWAY_RESUME`         | Resume downloading logs from the oldest downloaded log | No                                    | Any boolean value    |
| Account Token  | -              | `RAILWAY_ACCOUNT_TOKEN`  | Railway account token for authentication               | Yes                                   | Must be a valid UUID |

**Examples:**

Download all error logs:
```bash
go run . --deployment <deploymentId> --filter "@level:error"
```

Download all error logs with a specific message:
```bash
go run . --deployment <deploymentId> --filter "@level:error \"failed to prepare batch\""
```

Download all logs for a specific service:
```bash
go run . --service <serviceId> --environment <environmentId>
```

Download all logs for a specific service with a specific message:
```bash
go run . --service <serviceId> --environment <environmentId> --filter "@level:error \"failed to prepare batch\""
```

Download all logs for a specific service with a specific message and resume from the last downloaded log:

```bash
go run . --service <serviceId> --environment <environmentId> --filter "@level:error \"failed to prepare batch\"" --resume
```

Download all logs for a specific service with a specific message and overwrite the existing log file:

```bash
go run . --service <serviceId> --environment <environmentId> --filter "@level:error \"failed to prepare batch\"" --overwrite
```

See Railway's documentation on [logging](https://docs.railway.com/guides/logs#filtering-logs) for more information on the filter syntax.

### Notes

- The log downloader will only download deployment logs, it will not download HTTP logs.