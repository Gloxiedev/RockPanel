# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| Latest release | Yes |
| Previous release | Security fixes only |

## Reporting a Vulnerability

Report security vulnerabilities privately via GitHub Security Advisories.

Do not create public issues for security vulnerabilities.

Include:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix if available

We aim to respond within 72 hours and release a fix within 30 days for critical vulnerabilities.

## Security Expectations

RockPanel implements:

- **Authentication** — bcrypt password hashing, secure session cookies, session expiration
- **Authorization** — role-based access control (admin/user), resource-level permissions
- **Input Validation** — all API inputs validated, path traversal prevented
- **Command Execution** — structured process execution, no shell interpolation
- **File Operations** — sandboxed to configured data directory, symlink protection
- **Archive Extraction** — safe extraction with path validation, size limits
- **Network Security** — SSRF protection for outbound requests, HTTPS enforcement option
- **Web Security** — CSP headers, CSRF tokens, XSS prevention in templates
- **Privilege Separation** — managed processes run as dedicated users, not root
- **Secrets Management** — API tokens hashed in database, never logged

## Threat Model

RockPanel assumes:

- Trusted administrator installs and configures the panel
- Panel runs on a dedicated or shared Linux server
- Panel web interface exposed to trusted networks or with TLS
- Managed applications may be untrusted user code
- Docker daemon access implies container escape risk
- File manager access constrained to panel data directory

## Out of Scope

- Vulnerabilities in managed applications (Minecraft servers, Docker containers, user apps)
- Vulnerabilities in underlying system (kernel, Docker, Java runtime)
- Social engineering or physical access attacks
- Denial of service via resource exhaustion (mitigated via resource limits)

## Security Hardening

Recommended for production:

- Enable TLS with valid certificates
- Restrict panel port with firewall
- Use strong admin password
- Create non-admin users for daily operations
- Enable automatic updates
- Monitor audit logs
- Backup configuration and database regularly