# RockPanel

Lightweight all-in-one server management panel for Linux.

## Features

- **Server Management** — Process/service management with resource limits
- **Application Management** — Node.js, Python, static sites, custom apps with Git deployment
- **Docker** — Containers, images, volumes, networks, Compose support
- **Minecraft** — Vanilla, Paper, Purpur, Fabric, Forge with live console
- **Websites** — Reverse proxy, custom domains, automatic HTTPS
- **Databases** — PostgreSQL, MySQL, MariaDB, SQLite management
- **File Manager** — Browse, upload, edit, archive with security controls
- **Backups** — Scheduled backups for servers, apps, databases, files
- **Scheduler** — Lightweight cron-like task scheduler
- **Monitoring** — CPU, RAM, disk, network, process/container state
- **Alerts** — Disk, RAM, crash, backup, certificate notifications
- **CLI** — Full command-line interface mirroring web functionality
- **REST API** — Documented API with token authentication
- **Multi-user** — Admin/user roles with granular permissions

## Architecture

```
RockPanel
├── Web UI (single-page application)
├── REST API
├── Core
│   ├── Process Manager
│   ├── File Manager
│   ├── Scheduler
│   ├── Metrics
│   ├── Authentication
│   └── Permissions
├── Minecraft
├── Docker
├── Applications
├── Websites
├── Databases
└── Backups
```

Single binary, SQLite database, no external dependencies.

## Installation

```bash
curl -fsSL https://github.com/rockpanel/rockpanel/releases/latest/download/install.sh | sh
rockpanel init
```

Or download the binary directly from [GitHub Releases](https://github.com/rockpanel/rockpanel/releases).

### Manual Installation

```bash
wget https://github.com/rockpanel/rockpanel/releases/latest/download/rockpanel-linux-amd64
chmod +x rockpanel-linux-amd64
sudo mv rockpanel-linux-amd64 /usr/local/bin/rockpanel
rockpanel init
```

## Quick Start

```bash
rockpanel init
rockpanel start
```

Open http://localhost:8080 — default login: `admin` / `changeme`

## CLI Reference

```
rockpanel init                 # Initialize configuration and database
rockpanel start                # Start the panel
rockpanel stop                 # Stop the panel
rockpanel restart              # Restart the panel
rockpanel status               # Show panel status
rockpanel logs                 # Show panel logs
rockpanel update               # Update to latest version

rockpanel server list          # List managed servers
rockpanel server start <name>  # Start a server
rockpanel server stop <name>   # Stop a server
rockpanel server restart <name> # Restart a server

rockpanel app list             # List applications
rockpanel app start <name>     # Start an application
rockpanel app stop <name>      # Stop an application
rockpanel app restart <name>   # Restart an application

rockpanel docker list          # List Docker containers

rockpanel backup create <name> # Create a backup

rockpanel token create         # Create API token
rockpanel token list           # List API tokens
rockpanel token revoke <id>    # Revoke API token
```

## Minecraft

First-class Minecraft server management:

- Server types: Vanilla, Paper, Purpur, Fabric, Forge
- Per-server Java version selection (8, 17, 21, 25)
- Live console streaming with command input
- server.properties, EULA, memory, port management
- Player list, whitelist, ban, op/deop

## Docker

Native Docker API integration:

- Container lifecycle (start, stop, restart, pause, kill, remove)
- Image management
- Volume and network listing
- Resource metrics (CPU, RAM, network)
- Docker Compose project support

## Applications

Git-based deployment workflow:

```
Git Repository → Clone → Install → Build → Start → Monitor
```

Supports GitHub, GitLab, generic Git. Optional webhook auto-deploy.

## Websites

Reverse proxy with automatic HTTPS:

```
example.com → RockPanel → localhost:3000
```

WebSocket support, custom domains, Let's Encrypt integration.

## Databases

Basic management for PostgreSQL, MySQL, MariaDB, SQLite:

- Create/delete databases
- User and credential management
- Connection information

## Backups

Scheduled backups for all resource types with create/restore/delete/download.

## API

REST API at `/api/v1` with token authentication. All web dashboard functionality available via API.

```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/servers
```

## Configuration

`/etc/rockpanel/config.yaml`:

```yaml
port: 8080
data_dir: /var/lib/rockpanel
log_dir: /var/log/rockpanel
database: sqlite
tls:
  enabled: false
  cert_file: ""
  key_file: ""
modules:
  minecraft: true
  docker: true
  apps: true
  websites: true
  databases: true
```

## Development

```bash
git clone https://github.com/rockpanel/rockpanel
cd rockpanel
go build -o rockpanel ./cmd/rockpanel
./rockpanel init
./rockpanel start
```

### Requirements

- Go 1.22+
- Linux (systemd for service installation)

### Testing

```bash
go test ./...
```

## Security

- No plaintext passwords (bcrypt)
- Secure session management
- Path traversal prevention
- Command injection protection
- SSRF/XSS/CSRF protection
- Privilege separation for managed processes
- Secure archive extraction
- Input validation on all endpoints

See [SECURITY.md](SECURITY.md) for vulnerability reporting.

## Resource Usage

| Scenario | RAM | CPU (idle) |
|----------|-----|------------|
| Panel only | ~15 MB | <0.5% |
| + Application | ~25 MB | <1% |
| + Minecraft | ~30 MB | <1% |
| + Docker | ~25 MB | <1% |

Startup: <500ms on modern hardware.

## License

MIT — see [LICENSE](LICENSE).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Changelog

See [CHANGELOG.md](CHANGELOG.md).