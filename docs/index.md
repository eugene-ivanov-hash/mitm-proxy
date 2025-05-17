# MITM Proxy Documentation

Welcome to the MITM Proxy documentation. This guide will help you understand, install, and use the MITM Proxy utility.

## Table of Contents

1. [Overview](overview.md) - Introduction and key features
2. [Installation](installation.md) - How to install and set up the proxy
3. [Usage](usage.md) - Basic and advanced usage instructions
4. [Rules System](rules.md) - Understanding and creating rules
5. [Examples](examples.md) - Example rule configurations
6. [Troubleshooting](troubleshooting.md) - Common issues and solutions

## Quick Start

1. Generate CA certificate and key:

   **Option 1: Using OpenSSL**
   
   *For macOS/Linux:*
   ```bash
   openssl genrsa -out ca.key 2048
   openssl req -new -x509 -days 3650 -key ca.key -out ca.crt -subj "/CN=MITM Proxy CA"
   ```
   
   *For Windows:*
   ```powershell
   # Using OpenSSL in Windows (requires OpenSSL installation)
   openssl genrsa -out ca.key 2048
   openssl req -new -x509 -days 3650 -key ca.key -out ca.crt -subj "/CN=MITM Proxy CA"
   ```
   
   **Option 2: Using mkcert (Recommended for development)**
   
   *For macOS/Linux:*
   ```bash
   # Install mkcert
   # macOS: brew install mkcert
   # Linux: apt install mkcert
   
   # Generate and install certificate
   mkcert -install
   mkcert -cert-file ca.crt -key-file ca.key localhost 127.0.0.1 ::1
   ```
   
   *For Windows:*
   ```powershell
   # Install mkcert (requires Chocolatey)
   choco install mkcert
   
   # Generate and install certificate
   mkcert -install
   mkcert -cert-file ca.crt -key-file ca.key localhost 127.0.0.1 ::1
   ```

2. Get the MITM Proxy binary:

   **Option 1: Download from GitHub Releases**
   ```bash
   # For macOS Intel (AMD64)
   curl -L https://github.com/eugene-ivanov-hash/mitm-proxy/releases/latest/download/mitm-proxy-darwin-amd64 -o mitm-proxy
   chmod +x mitm-proxy
   
   # For macOS Apple Silicon (ARM64)
   curl -L https://github.com/eugene-ivanov-hash/mitm-proxy/releases/latest/download/mitm-proxy-darwin-arm64 -o mitm-proxy
   chmod +x mitm-proxy
   
   # For Linux
   curl -L https://github.com/eugene-ivanov-hash/mitm-proxy/releases/latest/download/mitm-proxy-linux-amd64 -o mitm-proxy
   chmod +x mitm-proxy
   
   # For Windows (using PowerShell)
   Invoke-WebRequest -Uri https://github.com/eugene-ivanov-hash/mitm-proxy/releases/latest/download/mitm-proxy-windows-amd64.exe -OutFile mitm-proxy.exe
   ```

   **Option 2: Build from source**
   ```bash
   go build -o mitm-proxy
   ```

3. Run the proxy:

   **For macOS/Linux:**
   ```bash
   # Option 1: Using generated certificate files
   ./mitm-proxy -cacertfile ca.crt -cakeyfile ca.key
   
   # Option 2: Using mkcert's root CA (if you used mkcert)
   ./mitm-proxy -cacertfile "$(mkcert -CAROOT)/rootCA.pem" -cakeyfile "$(mkcert -CAROOT)/rootCA-key.pem"
   ```
   
   **For Windows:**
   ```powershell
   # Option 1: Using generated certificate files
   .\mitm-proxy.exe -cacertfile ca.crt -cakeyfile ca.key
   
   # Option 2: Using mkcert's root CA (if you used mkcert)
   $CAROOT = mkcert -CAROOT
   .\mitm-proxy.exe -cacertfile "$CAROOT\rootCA.pem" -cakeyfile "$CAROOT\rootCA-key.pem"
   ```

3. Configure your client to use the proxy (default: `127.0.0.1:9999`)

4. Create rules in the `proxy_rules` directory

See the [Usage Guide](usage.md) for more detailed instructions.
