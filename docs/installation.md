# Installation Guide

## Prerequisites

- Go 1.24 or higher (only if building from source)
- OpenSSL (for generating CA certificates using Option 1)
- mkcert (for generating CA certificates using Option 2)

## Installation Steps

### 1. Get the MITM Proxy

**Option 1: Download from GitHub Releases**

*For macOS Intel (AMD64):*
```bash
curl -L https://github.com/eugene-ivanov-hash/mitm-proxy/releases/latest/download/mitm-proxy-darwin-amd64 -o mitm-proxy
chmod +x mitm-proxy
```

*For macOS Apple Silicon (ARM64):*
```bash
curl -L https://github.com/eugene-ivanov-hash/mitm-proxy/releases/latest/download/mitm-proxy-darwin-arm64 -o mitm-proxy
chmod +x mitm-proxy
```

*For Linux:*
```bash
curl -L https://github.com/eugene-ivanov-hash/mitm-proxy/releases/latest/download/mitm-proxy-linux-amd64 -o mitm-proxy
chmod +x mitm-proxy
```

*For Windows (using PowerShell):*
```powershell
Invoke-WebRequest -Uri https://github.com/eugene-ivanov-hash/mitm-proxy/releases/latest/download/mitm-proxy-windows-amd64.exe -OutFile mitm-proxy.exe
```

**Option 2: Build from Source**

1. Clone the repository:
```bash
git clone https://github.com/eugene-ivanov-hash/mitm-proxy.git
cd mitm-proxy
```

2. Build the application:
```bash
go build -o mitm-proxy
```

### 2. Generate CA Certificate and Key

To intercept HTTPS traffic, the proxy needs a CA certificate that will be used to sign dynamically generated certificates.

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

### 3. Run the Proxy

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

### 4. Configure Your Browser or System to Trust the CA

For the proxy to work with HTTPS connections, you need to add the generated CA certificate to your system or browser's trusted certificate authorities.

#### On macOS:

1. Add the certificate to the Keychain:
   ```bash
   sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ca.crt
   ```
   
   Note: If you used mkcert, this step is already done for you by the `mkcert -install` command.

#### On Windows:

1. Double-click the CA certificate file
2. Click "Install Certificate"
3. Select "Place all certificates in the following store"
4. Browse and select "Trusted Root Certification Authorities"
5. Click "Next" and "Finish"

   Note: If you used mkcert, this step is already done for you by the `mkcert -install` command.

#### On Linux:

1. Copy the certificate to the trusted CA directory:
   ```bash
   sudo cp ca.crt /usr/local/share/ca-certificates/
   sudo update-ca-certificates
   ```

   Note: If you used mkcert, this step is already done for you by the `mkcert -install` command.

#### For Firefox:

Firefox uses its own certificate store:
1. Open Firefox Preferences
2. Go to Privacy & Security
3. Scroll down to Certificates and click "View Certificates"
4. In the Authorities tab, click "Import" and select your CA certificate
5. Check "Trust this CA to identify websites" and click "OK"
