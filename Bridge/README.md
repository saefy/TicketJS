# thermalik-bridge

`thermalik-bridge` is the local companion service for the browser SDK served at `http://localhost:8000/thermalik.js`.

It runs on the user workstation, listens on `127.0.0.1:9100` by default, exposes a small HTTP API for the browser, and dispatches ESC/POS bytes to local printers.

## Installation

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/saefy/thermalik/main/scripts/install-bridge.sh | bash
```

### Windows (PowerShell)

```powershell
iwr https://raw.githubusercontent.com/saefy/thermalik/main/scripts/install-bridge.ps1 | iex
```

After installation run the interactive setup wizard:

```bash
thermalik-bridge setup
```

The wizard will:

1. Ask for your printer name, transport (network/usb/serial), address and paper width.
2. Save the config to the OS default path.
3. Install and start thermalik-bridge as a system service (LaunchAgent on macOS, systemd on Linux, Windows Service on Windows).

## Commands

```
thermalik-bridge <command> [flags]

Commands:
  setup      Interactive wizard: configure printers and install as a system service
  run        Run the bridge in the foreground (used internally by the service manager)
  start      Start the installed system service
  stop       Stop the installed system service
  restart    Restart the installed system service
  status     Show whether the system service is running
  update     Check for updates and apply if available
  uninstall  Remove the system service
  version    Print version information
```

### `run` flags

| Flag | Default | Description |
|---|---|---|
| `-config <path>` | OS default path | Path to the JSON config file |
| `-listen <addr>` | value from config | Override the listen address (e.g. `127.0.0.1:9200`) |

## Updates

### Check and apply manually

```bash
thermalik-bridge update
```

This command:

1. Queries the GitHub Releases API.
2. If a newer version is found, downloads the matching binary for your platform.
3. Verifies the SHA256 checksum from `checksums.txt`.
4. Atomically replaces the running binary.
5. Restarts the system service.

### Automatic update notifications

On every startup, thermalik-bridge performs a non-blocking background check.
If a new version is available it logs a notice:

```
update available: v0.3.0 → run 'thermalik-bridge update' to upgrade
```

The running service is never interrupted by an update check.

## Service lifecycle

```bash
thermalik-bridge start     # start the service
thermalik-bridge stop      # stop the service
thermalik-bridge restart   # restart the service
thermalik-bridge status    # check if running
thermalik-bridge uninstall # remove from system services
```

The service is registered under the name `thermalik-bridge`:

| OS | Mechanism |
|---|---|
| macOS | LaunchAgent at `~/Library/LaunchAgents/thermalik-bridge.plist` |
| Linux | systemd unit at `/etc/systemd/system/thermalik-bridge.service` |
| Windows | Windows Service named `thermalik-bridge` |

## Uninstall

**macOS / Linux:**

```bash
thermalik-bridge stop
thermalik-bridge uninstall
sudo rm $(which thermalik-bridge)
```

**Windows (PowerShell, as Administrator):**

```powershell
thermalik-bridge stop
thermalik-bridge uninstall
Remove-Item "$env:LOCALAPPDATA\thermalik-bridge" -Recurse -Force
```

## Configuration

### File locations

Config is saved at the OS default path during `setup`:

| OS | Config path |
|---|---|
| macOS | `~/Library/Application Support/thermalik-bridge/config.json` |
| Linux | `~/.config/thermalik-bridge/config.json` |
| Windows | `%APPDATA%\thermalik-bridge\config.json` |

Job state is persisted separately:

| OS | Jobs path |
|---|---|
| macOS | `~/Library/Caches/thermalik-bridge/jobs.json` |
| Linux | `~/.cache/thermalik-bridge/jobs.json` |
| Windows | `%LOCALAPPDATA%\thermalik-bridge\jobs.json` |

Use `configs/bridge.example.json` as a reference for manual edits.

### Full config schema

```json
{
  "server": {
    "address": "127.0.0.1:9100"
  },
  "storage": {
    "job_store_path": "/path/to/jobs.json"
  },
  "printers": []
}
```

`server.address` defaults to `127.0.0.1:9100`. The bridge always binds to loopback; do not expose it on a public interface.

### Network printer

```json
{
  "id": "caja-principal",
  "name": "Caja principal",
  "transport": "network",
  "default": true,
  "capabilities": {
    "paper_width_mm": 58,
    "encodings": ["cp437", "cp850", "latin1"],
    "job_tracking": true
  },
  "network": {
    "address": "192.168.1.50:9100",
    "timeout": "5s"
  }
}
```

`timeout` accepts Go duration strings: `5s`, `10s`, `1m`. Defaults to `5s`.

### USB printer (macOS / Linux via CUPS)

USB printing requires an installed print queue. To find the queue name:

```bash
lpstat -p
```

```json
{
  "id": "caja-usb",
  "name": "Caja USB",
  "transport": "usb",
  "default": false,
  "capabilities": {
    "paper_width_mm": 58,
    "encodings": ["cp437", "cp850"]
  },
  "usb": {
    "queue": "EPSON_TM_T20III",
    "vendor_id": "04b8",
    "product_id": "0e15"
  }
}
```

- `queue` is required and must match the queue name reported by `lpstat -p`.
- The bridge sends bytes with `lp -d <queue> -o raw`.
- `vendor_id` and `product_id` are optional metadata; they are not used for routing.
- Requires CUPS (`lp` command available). Common on macOS and Linux. Not supported on Windows.

### Serial printer (stub — not yet implemented)

The config key is accepted and validated but serial transport always returns an error.
Do not configure serial printers in production until this transport is implemented.

```json
{
  "id": "caja-serial",
  "name": "Caja serial",
  "transport": "serial",
  "serial": {
    "port": "/dev/ttyUSB0",
    "baud": 9600
  }
}
```

### Printer `capabilities`

`capabilities` is a free-form metadata map forwarded to the SDK via `GET /printers`.
No capabilities are enforced by the bridge itself.

| Key | Type | Example | Description |
|---|---|---|---|
| `paper_width_mm` | number | `58` | Nominal paper width |
| `encodings` | string[] | `["cp437","cp850"]` | Supported encodings |
| `job_tracking` | bool | `true` | Whether to persist job state |

### Multiple printers

You can define multiple printers. Only one may have `"default": true`.

```json
{
  "printers": [
    { "id": "caja-1", "default": true, ... },
    { "id": "cocina",  "default": false, ... }
  ]
}
```

## Endpoints

All endpoints listen on `127.0.0.1:9100` (loopback only).
CORS is restricted to `http://localhost:*` and `http://127.0.0.1:*` origins.

---

### `GET /health`

Returns the bridge status and number of configured printers.

**Response `200`:**

```json
{
  "status": "ok",
  "configured_printers": 2
}
```

---

### `GET /printers`

Returns all configured printers with their status and connection details.

**Response `200`:**

```json
{
  "data": [
    {
      "id": "caja-principal",
      "name": "Caja principal",
      "transport": "network",
      "status": "ready",
      "default": true,
      "capabilities": {
        "paper_width_mm": 58,
        "encodings": ["cp437", "cp850"]
      },
      "connection": {
        "address": "192.168.1.50:9100",
        "timeout": "5s"
      }
    }
  ]
}
```

Possible `status` values: `ready` (network/usb), `stub` (serial).

---

### `POST /print`

Enqueues a raw ESC/POS byte payload for a printer. Returns immediately with a job reference.

**Request body:**

```json
{
  "printer": "caja-principal",
  "bytes": [27, 64, 72, 101, 108, 108, 111, 10],
  "render": {
    "template": "ticket-venta",
    "order_id": "1234"
  }
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `printer` | string | ✓ | Printer `id` from config |
| `bytes` | number[] | ✓ | ESC/POS bytes (integers 0–255) |
| `render` | object | — | Arbitrary metadata attached to the job (passthrough, not interpreted) |

**Response `202`:**

```json
{
  "job_id": "job_3f8a1c2d4e5b6a7c",
  "status": "queued",
  "printer": "caja-principal",
  "created_at": "2026-05-30T17:00:00Z",
  "updated_at": "2026-05-30T17:00:00Z",
  "completed_at": null
}
```

**Error responses:**

| Status | Condition |
|---|---|
| `400` | Missing `printer`, empty `bytes`, or a byte value outside 0–255 |
| `404` | Printer ID not found in config |
| `500` | Internal error (job could not be persisted or enqueued) |

---

### `GET /jobs/{jobId}`

Returns the current state of a print job.

**Response `200`:**

```json
{
  "job_id": "job_3f8a1c2d4e5b6a7c",
  "status": "completed",
  "printer": "caja-principal",
  "created_at": "2026-05-30T17:00:00Z",
  "updated_at": "2026-05-30T17:00:01Z",
  "completed_at": "2026-05-30T17:00:01Z",
  "render": {
    "order_id": "1234"
  }
}
```

On failure:

```json
{
  "job_id": "job_3f8a1c2d4e5b6a7c",
  "status": "failed",
  "error": "dial tcp 192.168.1.50:9100: connect: connection refused",
  ...
}
```

**Job status lifecycle:**

```
queued → processing → completed
                    ↘ failed
```

| Status | Description |
|---|---|
| `queued` | Job accepted, waiting in the printer queue |
| `processing` | Bytes are being sent to the printer |
| `completed` | Bytes delivered successfully |
| `failed` | Delivery failed; see `error` field |

**Error responses:**

| Status | Condition |
|---|---|
| `400` | Missing job ID in path |
| `404` | Job ID not found |

---

## SDK compatibility

```js
thermalik.configure({
  apiKey: "trk_your_api_key",
  apiBaseUrl: "http://localhost:8000/api/v1",
  bridgeUrl: "http://localhost:9100",
  printer: "caja-principal"
});

thermalik.print({
  html: "<p>Hello</p>",
  printerWidth: 58,
  encoding: "cp437"
});
```

The SDK sends pre-rendered ESC/POS bytes to `POST /print`. The bridge does not parse HTML or perform encoding; that is handled by the Laravel render engine.

## Local development

```bash
cd thermalik-bridge

# Build and run without installing the service
go build -o thermalik-bridge ./cmd/thermalik-bridge

# Create a local config file (copy the example and edit as needed)
cp configs/bridge.example.json ./bridge.local.json

# Run the bridge in the foreground with the local config
go run ./cmd/thermalik-bridge run -config ./bridge.local.json
```

Test with:

```bash
curl http://127.0.0.1:9100/health
curl http://127.0.0.1:9100/printers
curl -X POST http://127.0.0.1:9100/print \
  -H "Content-Type: application/json" \
  -d '{"printer":"caja-principal","bytes":[27,64,72,101,108,108,111,10]}'
```

## Building a release

Releases are built automatically by goreleaser when a tag is pushed:

```bash
git tag v0.2.0
git push origin v0.2.0
```

The GitHub Actions workflow `.github/workflows/release-bridge.yml` runs goreleaser and publishes binaries for darwin/linux/windows × amd64/arm64 to the GitHub Release.

**Build a local snapshot (no tag required):**

```bash
# Install goreleaser once
brew install goreleaser   # macOS
# or: go install github.com/goreleaser/goreleaser/v2@latest

cd /path/to/repo
goreleaser release --snapshot --clean
# Binaries written to dist/
```

**Build a single binary for testing:**

```bash
cd thermalik-bridge
go build -ldflags "-X thermalik-bridge/internal/version.Version=dev" ./cmd/thermalik-bridge
./thermalik-bridge version
```

