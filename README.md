# driftwatch

Lightweight daemon that detects configuration drift between running services and their declared state in version control.

---

## Installation

```bash
go install github.com/yourorg/driftwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourorg/driftwatch.git && cd driftwatch && go build -o driftwatch .
```

---

## Usage

Create a config file pointing to your services and their source-of-truth repository:

```yaml
# driftwatch.yaml
interval: 60s
repo: https://github.com/yourorg/infra-config
services:
  - name: api-server
    path: services/api-server/config.yaml
    endpoint: http://api-server:8080/config
  - name: worker
    path: services/worker/config.yaml
    endpoint: http://worker:9090/config
```

Run the daemon:

```bash
driftwatch --config driftwatch.yaml
```

When drift is detected, driftwatch logs a structured alert and optionally sends a webhook notification:

```
{"level":"warn","service":"api-server","drift":true,"fields":["timeout","max_retries"],"timestamp":"2024-11-01T12:00:00Z"}
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `driftwatch.yaml` | Path to config file |
| `--interval` | `60s` | Poll interval |
| `--dry-run` | `false` | Log drift without alerting |
| `--log-format` | `json` | Output format (`json` or `text`) |

---

## License

MIT © yourorg