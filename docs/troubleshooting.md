# Troubleshooting

This guide covers common issues and their solutions when using the MITM Proxy.

## Certificate Issues

### Problem: Browser Shows Certificate Error

**Symptoms**: Browser displays "Your connection is not private" or similar warnings.

**Solutions**:
1. Verify that you've installed the CA certificate correctly in your system or browser's trust store.
2. Check that you're using the correct CA certificate and key files with the proxy.
3. Restart your browser after installing the certificate.

### Problem: Certificate Generation Fails

**Symptoms**: Error messages about certificate creation in the proxy logs.

**Solutions**:
1. Ensure the CA key file has the correct permissions.
2. Verify that the CA certificate is valid and properly formatted.
3. Check that the OpenSSL version is compatible.

## Connection Issues

### Problem: Cannot Connect to Proxy

**Symptoms**: Connection timeouts or refused connections.

**Solutions**:
1. Verify the proxy is running with `ps aux | grep mitm-proxy`.
2. Check that you're using the correct proxy address and port.
3. Ensure no firewall is blocking the connection.
4. Verify the proxy is listening on the correct interface (use `0.0.0.0` to listen on all interfaces).

### Problem: Proxy Starts but No Traffic Flows

**Symptoms**: Proxy starts successfully but doesn't intercept any traffic.

**Solutions**:
1. Verify your client is configured to use the proxy.
2. Check that the proxy address is correctly specified in your client settings.
3. Try a simple test with curl: `curl -x http://127.0.0.1:9999 http://example.com`.

## Rule Issues

### Problem: Rules Not Applied

**Symptoms**: Traffic passes through the proxy but rules aren't being applied.

**Solutions**:
1. Check that the rule file is in the correct directory.
2. Verify the rule file has the correct extension (`.yaml` or `.yml`).
3. Ensure both the rule file and the specific rule are enabled.
4. Check the rule condition to make sure it matches your traffic.
5. Run with `-debug` flag to see detailed rule evaluation logs.

### Problem: Rule Compilation Errors

**Symptoms**: Errors about rule compilation when starting the proxy.

**Solutions**:
1. Check the CEL expression syntax in your rule.
2. Verify that any Go code in the script section is valid.
3. Ensure all required imports are specified.
4. Check for typos in property names.

## Performance Issues

### Problem: Slow Proxy Performance

**Symptoms**: Noticeably slower browsing or API responses when using the proxy.

**Solutions**:
1. Minimize the number of rules to only those necessary.
2. Optimize rule conditions to avoid expensive operations.
3. Avoid reading and writing large request/response bodies when not necessary.
4. Consider running the proxy on a more powerful machine.

## Environment Variable Issues

### Problem: Environment Variables Not Available in Rules

**Symptoms**: Rules using environment variables don't work as expected.

**Solutions**:
1. Verify the environment file exists and is correctly formatted.
2. Check that the environment variable is defined in the file.
3. Use the `-env` flag to explicitly specify the environment file.
4. Check the template syntax in your rule: `{{ .Envs.VARIABLE_NAME }}`.

## Debugging

For advanced troubleshooting, use the `-debug` flag to enable detailed logging:

```bash
./mitm-proxy -cacertfile ca.crt -cakeyfile ca.key -debug
```

This will show:
- Rule evaluation results
- Request and response details
- Certificate generation information
- Connection handling

You can also use the `-test` flag to test your rules without starting the proxy:

```bash
./mitm-proxy -cacertfile ca.crt -cakeyfile ca.key -rulesdir ./my_rules -test
```
