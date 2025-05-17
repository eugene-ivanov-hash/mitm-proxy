# Example Rules

This document provides practical examples of rules for the MITM Proxy.

## Basic Request Modification

```yaml
enabled: true
rules:
  - name: "Add Custom Header to All Requests"
    enabled: true
    change: "request"
    rule: "true"  # Apply to all requests
    action: "script"
    script: |
      req.Header.Add("X-Custom-Header", "CustomValue")
```

## Response Modification

```yaml
enabled: true
rules:
  - name: "Modify JSON Response"
    enabled: true
    change: "response"
    rule: "resp.Header.Get('Content-Type').contains('application/json')"
    action: "script"
    import: |
      "io"
      "bytes"
      "encoding/json"
    script: |
      body, err := io.ReadAll(resp.Body)
      if err != nil {
        return err
      }
      
      // Parse JSON
      var data map[string]interface{}
      err = json.Unmarshal(body, &data)
      if err != nil {
        return err
      }
      
      // Modify data
      data["modified_by"] = "MITM Proxy"
      
      // Convert back to JSON
      newBody, err := json.Marshal(data)
      if err != nil {
        return err
      }
      
      resp.Body = io.NopCloser(bytes.NewReader(newBody))
```

## Conditional Request Blocking

```yaml
enabled: true
rules:
  - name: "Block Specific User Agent"
    enabled: true
    change: "request"
    rule: "req.Header.Get('User-Agent').contains('BadBot')"
    action: "reject"
```

## URL Rewriting

```yaml
enabled: true
rules:
  - name: "Redirect API Requests"
    enabled: true
    change: "request"
    rule: "req.URL.Host == 'api.example.com'"
    action: "script"
    script: |
      // Change the host to a different API endpoint
      req.URL.Host = "api-test.example.com"
      req.Host = "api-test.example.com"
```

## Using Environment Variables

```yaml
enabled: true
rules:
  - name: "Modify Requests Based on Environment"
    enabled: true
    change: "request"
    rule: "req.URL.Host == '{{ .Envs.TARGET_HOST }}'"
    action: "script"
    script: |
      req.Header.Add("X-Environment", "{{ .Envs.ENVIRONMENT }}")
```

## Response Delay Simulation

```yaml
enabled: true
rules:
  - name: "Simulate Slow Response"
    enabled: true
    change: "response"
    rule: "req.URL.Path.startsWith('/api/slow')"
    action: "script"
    import: |
      "time"
    script: |
      // Sleep for 2 seconds to simulate slow response
      time.Sleep(2 * time.Second)
```

## Body Content Replacement

```yaml
enabled: true
rules:
  - name: "Replace Text in HTML Responses"
    enabled: true
    change: "response"
    rule: "resp != null && resp.Header.Get('Content-Type').contains('text/html')"
    action: "script"
    import: |
      "io"
      "strings"
    script: |
      body, err := io.ReadAll(resp.Body)
      if err != nil {
        return err
      }
      
      // Replace text in the HTML
      newBody := strings.Replace(string(body), "Original Text", "Modified Text", -1)
      
      resp.Body = io.NopCloser(strings.NewReader(newBody))
```

## WebSocket Modification

```yaml
enabled: true
rules:
  - name: "Modify WebSocket Upgrade Request"
    enabled: true
    change: "request"
    rule: "req.Header.Get('Upgrade') == 'websocket'"
    action: "script"
    script: |
      // Add custom header to WebSocket upgrade request
      req.Header.Add("X-WebSocket-Custom", "CustomValue")
```
