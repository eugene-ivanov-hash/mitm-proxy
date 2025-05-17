# MITM Proxy

A powerful Man-In-The-Middle proxy utility for intercepting, inspecting, and modifying HTTP/HTTPS traffic.

## Features

- Intercepts and modifies both HTTP and HTTPS traffic
- Dynamic TLS certificate generation for HTTPS connections
- Rule-based system for conditional traffic modification
- Support for custom Go scripts to modify requests and responses
- WebSocket connection handling
- Environment variable integration

## Quick Start

### Prerequisites

- Go 1.18 or higher
- OpenSSL (for generating CA certificates)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/eugene-ivanov-hash/mitm-proxy.git
   cd mitm-proxy
   ```

2. Generate CA certificate and key:
   ```bash
   openssl genrsa -out ca.key 2048
   openssl req -new -x509 -days 3650 -key ca.key -out ca.crt -subj "/CN=MITM Proxy CA"
   ```

3. Build the application:
   ```bash
   go build -o mitm-proxy
   ```

4. Add the CA certificate to your system's trusted certificates (see [Installation Guide](docs/installation.md))

### Usage

Start the proxy:
```bash
./mitm-proxy -cacertfile ca.crt -cakeyfile ca.key
```

Configure your client to use the proxy (default: `127.0.0.1:9999`).

For more options:
```bash
./mitm-proxy -help
```

### Creating Rules

Rules are defined in YAML files in the `proxy_rules` directory:

```yaml
enabled: true
rules:
  - name: "Add Custom Header"
    enabled: true
    change: "request"
    rule: "req.URL.Host == 'example.com'"
    action: "script"
    script: |
      req.Header.Add("X-Custom-Header", "CustomValue")
      return nil
```

## Documentation

For detailed documentation, see the [docs](docs/) directory:

- [Overview](docs/overview.md)
- [Installation Guide](docs/installation.md)
- [Usage Guide](docs/usage.md)
- [Rules System](docs/rules.md)
- [Example Rules](docs/examples.md)
- [Troubleshooting](docs/troubleshooting.md)

## License

See the [LICENSE](LICENSE) file for details.
