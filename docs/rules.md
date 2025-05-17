# Rules System

## Overview

The MITM Proxy uses a powerful rule-based system to intercept and modify HTTP/HTTPS traffic. Rules are defined in YAML files and placed in the rules directory (default: `proxy_rules`).

## Rule File Structure

Rule files must have a `.yaml` or `.yml` extension and follow this structure:

```yaml
enabled: true  # Enable/disable all rules in this file
rules:
  - name: "Rule Name"
    enabled: true  # Enable/disable this specific rule
    change: "request"  # "request" or "response"
    rule: "req.URL.Host == 'example.com'"  # CEL expression
    action: "script"  # "script" or "reject"
    import: |
      "strings"
      "io/ioutil"
    script: |
      // Go code to modify request or response
      body, _ := io.ReadAll(req.Body)
      req.Body = io.NopCloser(strings.NewReader(string(body) + " modified"))
```

## Rule Properties

| Property | Description | Required |
|----------|-------------|----------|
| `name` | Descriptive name for the rule | Yes |
| `enabled` | Whether the rule is active | Yes |
| `change` | Whether to modify request or response (`request` or `response`) | Yes |
| `rule` | CEL expression that determines when the rule applies | Yes |
| `action` | Action to take when rule matches (`script` or `reject`) | Yes |
| `import` | Go package imports for the script | No |
| `script` | Go code to execute when rule matches | Yes (if action is `script`) |

## CEL Expressions

Rules use [Common Expression Language (CEL)](https://github.com/google/cel-spec) to determine when they should be applied. CEL expressions have access to:

- `req`: The HTTP request object
- `resp`: The HTTP response object (only for response rules)

### Available Methods

#### Header Methods
Both `req.Header` and `resp.Header` provide the following method:
- `Get(headerName string) string` - Returns the value of the specified header

#### Body Methods
Both request and response objects provide:
- `getBody() string` - Returns the body content as a string

### Examples

```yaml
# Match requests to example.com
rule: "req.URL.Host == 'example.com'"

# Match POST requests
rule: "req.Method == 'POST'"

# Match responses with status code 200
rule: "resp.StatusCode == 200"

# Match requests with specific header
rule: "req.Header.Get('Content-Type').contains('application/json')"

# Match requests with specific path
rule: "req.URL.Path.startsWith('/api/')"

# Match requests with specific body content
rule: "req.getBody().contains('search_term')"

# Match responses with specific body content
rule: "resp.getBody().contains('error')"
```

## Script Actions

When a rule matches and the action is `script`, the Go code in the `script` property is executed. The script has access to:

- `req`: The HTTP request object (can be modified)
- `resp`: The HTTP response object (can be modified, null for request rules)

### Script Compilation

Scripts from rule.yaml files are compiled into Go functions at runtime. Each script is wrapped in a function with the following structure:

```go
package rule{N} // Where N is an auto-generated index number

import (
    "net/http"
    "fmt"
    // Additional imports from the rule's import section
)

func Modify(req *http.Request, resp *http.Response) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("rule{N}.Modify err: %v", r)
        }
    }()
    
    // The script content from your rule.yaml file is inserted here
    
    // The function implicitly returns nil if no error is returned
}
```

This means that your script code becomes the body of the `Modify` function, and you don't need to explicitly return `nil` at the end of successful scripts.

### Example Scripts

```go
// Modify request body
body, _ := io.ReadAll(req.Body)
newBody := strings.Replace(string(body), "original", "modified", -1)
req.Body = io.NopCloser(strings.NewReader(newBody))

// Add a header to the response
resp.Header.Add("X-Modified-By", "MITM-Proxy")

// Modify response body
body, _ := io.ReadAll(resp.Body)
newBody := strings.Replace(string(body), "original", "modified", -1)
resp.Body = io.NopCloser(strings.NewReader(newBody))
```

## Reject Action

When a rule matches and the action is `reject`, the request or response is rejected, and the connection is closed.

## Environment Variables

Rules can access environment variables using Go templates:

```yaml
rule: "req.URL.Host == '{{ .Envs.TARGET_HOST }}'"
```

Environment variables can be loaded from:

1. The system environment
2. A `.env` file in the current directory
3. A file specified with the `-env` flag

## Rule Processing Order

Rules are processed in the order they are loaded from the rules directory. For each request and response:

1. All matching request rules are applied before sending the request
2. All matching response rules are applied before returning the response to the client

If any rule returns an error or rejects the request/response, processing stops, and the error is returned.
