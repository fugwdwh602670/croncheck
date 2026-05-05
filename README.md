# croncheck

Lightweight daemon that monitors cron job execution and sends alerts on missed or failed runs.

## Installation

```bash
go install github.com/yourname/croncheck@latest
```

Or build from source:

```bash
git clone https://github.com/yourname/croncheck.git && cd croncheck && go build ./...
```

## Usage

Define your monitored jobs in a `croncheck.yaml` config file:

```yaml
jobs:
  - name: daily-backup
    schedule: "0 2 * * *"
    timeout: 30m
    alert:
      email: ops@example.com

  - name: hourly-sync
    schedule: "0 * * * *"
    timeout: 5m
    alert:
      slack: "#alerts"
```

Start the daemon:

```bash
croncheck --config croncheck.yaml
```

Wrap an existing cron job to report its status:

```bash
# In your crontab
0 2 * * * croncheck exec --job daily-backup -- /usr/local/bin/backup.sh
```

Check job status at any time:

```bash
croncheck status
```

```
JOB             LAST RUN              STATUS   NEXT RUN
daily-backup    2024-01-15 02:00:01   OK       2024-01-16 02:00:00
hourly-sync     2024-01-15 09:00:03   MISSED   2024-01-15 10:00:00
```

## License

MIT