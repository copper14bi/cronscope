# cronscope

Lightweight daemon that monitors and alerts on missed or long-running cron jobs via webhook.

## Installation

```bash
go install github.com/cronscope/cronscope@latest
```

Or build from source:

```bash
git clone https://github.com/cronscope/cronscope.git && cd cronscope && go build ./...
```

## Usage

Create a `cronscope.yaml` config file:

```yaml
webhook: "https://hooks.slack.com/services/your/webhook/url"
jobs:
  - name: "daily-backup"
    schedule: "0 2 * * *"
    timeout: 30m
    grace: 5m
  - name: "hourly-sync"
    schedule: "0 * * * *"
    timeout: 10m
```

Run the daemon:

```bash
cronscope --config cronscope.yaml
```

Wrap your existing cron jobs to report start and finish:

```bash
# In your crontab
0 2 * * * cronscope exec --job daily-backup -- /usr/local/bin/backup.sh
```

cronscope will fire your webhook if a job fails to start within its grace period or exceeds its timeout.

### Webhook Payload

```json
{
  "job": "daily-backup",
  "status": "missed",
  "expected_at": "2024-01-15T02:00:00Z",
  "message": "Job did not start within grace period of 5m"
}
```

## License

MIT © cronscope contributors