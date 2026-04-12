# portwatch

A lightweight CLI daemon that monitors and logs port binding changes on a host with alerting hooks.

---

## Installation

```bash
go install github.com/yourusername/portwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/portwatch.git && cd portwatch && go build -o portwatch .
```

---

## Usage

Start the daemon with default settings:

```bash
portwatch start
```

Watch specific ports and send alerts to a webhook on changes:

```bash
portwatch start --ports 80,443,8080 --webhook https://hooks.example.com/alert
```

Run a one-time snapshot of current port bindings:

```bash
portwatch scan
```

### Example Output

```
2024/01/15 10:32:01 [OPEN]   0.0.0.0:8080  pid=12345  proc=myapp
2024/01/15 10:33:47 [CLOSED] 0.0.0.0:8080  pid=12345  proc=myapp
2024/01/15 10:34:02 [OPEN]   127.0.0.1:5432 pid=67890 proc=postgres
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--interval` | `5s` | Polling interval |
| `--ports` | all | Comma-separated list of ports to watch |
| `--webhook` | — | URL to POST change events to |
| `--log` | stdout | Log output file path |

---

## License

MIT © 2024 yourusername