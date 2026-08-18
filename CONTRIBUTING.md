# Contributing to RockPanel

Thank you for contributing.

## Development Setup

### Requirements

- Go 1.22+
- Linux (for full functionality)
- Docker (for Docker integration testing)
- Java 8/17/21 (for Minecraft testing)

### Building

```bash
git clone https://github.com/rockpanel/rockpanel
cd rockpanel
go build -o rockpanel ./cmd/rockpanel
```

### Running Locally

```bash
./rockpanel init
./rockpanel start
```

Panel runs on http://localhost:8080

### Running Tests

```bash
go test ./...
go test -race ./...
go test -cover ./...
```

## Repository Structure

```
cmd/rockpanel          # CLI entry point
internal/
  core/                # Process manager, file manager, scheduler, metrics, auth, permissions
  api/                 # REST API handlers
  web/                 # Web dashboard handlers and templates
  minecraft/           # Minecraft server management
  docker/              # Docker management
  apps/                # Application management
  files/               # File manager
  backups/             # Backup system
  scheduler/           # Task scheduler
  websites/            # Website/reverse proxy management
  databases/           # Database management
  auth/                # Authentication
  metrics/             # System metrics collection
  logs/                # Log streaming
  config/              # Configuration management
pkg/
  types/               # Shared types
  utils/               # Utility functions
web/
  static/              # Static assets (CSS, JS)
  templates/           # HTML templates
tests/                 # Integration tests
```

## Pull Requests

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Run `go test ./...` and `go vet ./...`
5. Submit PR with clear description

## Coding Standards

- **Zero comments in source code** — use descriptive names, clean architecture, small functions, clear modules, good types
- Run `gofmt` on all Go files
- Follow Go idioms and best practices
- Keep functions small and focused
- Use interfaces for testability
- No global state
- Errors as values, not panics
- Structured logging

## Testing Requirements

- Unit tests for all core logic
- Integration tests for API endpoints
- Security tests for path traversal, command injection, symlink attacks, malicious archives, SSRF, XSS, CSRF, authentication bypass, privilege escalation
- CLI tests for all commands
- Benchmark tests for performance-critical paths

## Reporting Issues

Use GitHub Issues. Include:

- RockPanel version
- OS and architecture
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs