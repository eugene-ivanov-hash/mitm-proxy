# Usage Guide

## Basic Usage

To start the MITM Proxy with default settings:

```bash
./mitm-proxy -cacertfile ca.crt -cakeyfile ca.key
```

This will start the proxy on the default address `127.0.0.1:9999`.

## Command Line Options

The MITM Proxy supports several command line options:

| Option | Description | Default |
|--------|-------------|---------|
| `-addr` | Address to listen on | `127.0.0.1:9999` |
| `-cacertfile` | Path to CA certificate file | Required |
| `-cakeyfile` | Path to CA key file | Required |
| `-debug` | Enable debug logging | `false` |
| `-rulesdir` | Directory containing rule files | `proxy_rules` |
| `-env` | Path to environment file | `.env` (optional) |
| `-test` | Test rules without starting proxy | `false` |

Example with all options:

```bash
./mitm-proxy -addr 0.0.0.0:8080 -cacertfile ca.crt -cakeyfile ca.key -debug -rulesdir ./my_rules -env ./config.env
```

## Configuring Your Client

To use the proxy, you need to configure your client (browser, application, etc.) to use it:

### Browser Configuration

1. Set the proxy settings to point to the MITM Proxy address (e.g., `127.0.0.1:9999`)
2. Make sure the CA certificate is trusted (see Installation Guide)

### Command Line Tools

For tools like `curl`, you can use the proxy with:

```bash
http_proxy=127.0.0.1:9999 https_proxy=127.0.0.1:9999 curl https://example.com
```

### System-Wide Proxy (macOS)

```bash
networksetup -setwebproxy "Wi-Fi" 127.0.0.1 9999
networksetup -setsecurewebproxy "Wi-Fi" 127.0.0.1 9999
```

To disable the proxy:

```bash
networksetup -setwebproxystate "Wi-Fi" off
networksetup -setsecurewebproxystate "Wi-Fi" off
```

## Testing Rules

You can test your rules without starting the proxy server using the `-test` flag:

```bash
./mitm-proxy -cacertfile ca.crt -cakeyfile ca.key -rulesdir ./my_rules -test
```

This will run your rules against a test request and response to verify they work correctly.
