# Rapper Profile Management Guide

## 🎯 Overview

Rapper now supports **multiple configuration profiles** with **flexible headers** and **hot-reload** capabilities!

## 📁 Profile System

### What is a Profile?

A profile is a YAML configuration file that defines:
- HTTP request settings (method, URL template, body template)
- **Flexible headers** (Authorization, Cookie, X-API-Key, custom headers)
- CSV field mappings
- Worker count

### Auto-Discovery

Rapper automatically discovers all `.yml` and `.yaml` files in the current directory:

```
/your-project/
├── rapper             # Binary
├── api1.yml           # Profile 1 - Development
├── production.yml     # Profile 2 - Production
├── staging.yml        # Profile 3 - Staging
└── data/
    └── users.csv
```

## 🆕 New Configuration Format

### Structure

```yaml
request:
  method: POST                        # HTTP method
  url_template: "https://..."         # URL with {{.variables}}
  body_template: |                    # JSON body with {{.variables}}
    {
      "field": "{{.value}}"
    }
  headers:                            # Flexible headers map
    Authorization: "Bearer token"
    Cookie: "session=abc"
    X-API-Key: "key123"
    Custom-Header: "value"

csv:
  fields: [id, name, email]           # CSV columns to extract
  separator: ","                      # CSV separator

workers: 4                            # Number of concurrent workers
```

### Header Flexibility

**Before (old format):**
```yaml
token: "JWT_TOKEN"                    # Only Bearer auth
```

**After (new format):**
```yaml
request:
  headers:
    Authorization: "Bearer JWT_TOKEN"  # Bearer auth
    Cookie: "session_id=abc123"        # Cookie auth
    X-API-Key: "my-api-key"           # API key
    X-Custom: "anything"               # Custom headers
```

## 📚 Example Profiles

### Development Profile (`api1.yml`)

```yaml
request:
  method: POST
  url_template: "http://localhost:8080/api/users/{{.id}}"
  body_template: |
    {
      "name": "{{.name}}",
      "email": "{{.email}}"
    }
  headers:
    Authorization: "Bearer dev-token-123"
    Content-Type: "application/json"

csv:
  fields: [id, name, email]
  separator: ","

workers: 2
```

### Production Profile (`production.yml`)

```yaml
request:
  method: PUT
  url_template: "https://api.production.com/users/{{.id}}"
  body_template: |
    {
      "name": "{{.name}}",
      "email": "{{.email}}",
      "phone": "{{.phone}}"
    }
  headers:
    Authorization: "Bearer prod-token-xyz"
    X-API-Key: "production-key-987"
    Cookie: "session=abc; auth=xyz"
    X-Request-ID: "{{.id}}"         # Dynamic header from CSV!

csv:
  fields: [id, name, email, phone]
  separator: ","

workers: 8
```

### Staging Profile (`staging.yml`)

```yaml
request:
  method: PATCH
  url_template: "https://staging.example.com/customers/{{.customer_id}}"
  body_template: |
    {
      "status": "{{.status}}"
    }
  headers:
    X-API-Key: "staging-key-123"      # API key instead of Bearer
    Content-Type: "application/json"

csv:
  fields: [customer_id, status]
  separator: ","

workers: 4
```

## 🔄 Backward Compatibility

The old format still works! Legacy configs are automatically converted:

**Old format:**
```yaml
token: "JWT_TOKEN"
path:
  method: PUT
  template: "https://api.com/users/{{.id}}"
payload:
  template: '{"name": "{{.name}}"}'
csv:
  fields: [id, name]
  separator: ","
workers: 2
```

**Auto-converted to:**
```yaml
request:
  method: PUT
  url_template: "https://api.com/users/{{.id}}"
  body_template: '{"name": "{{.name}}"}'
  headers:
    Authorization: "Bearer JWT_TOKEN"
    Content-Type: "application/json"
csv:
  fields: [id, name]
  separator: ","
workers: 2
```

## 🚀 Future Features (Coming Soon)

### Profile Switching in UI
```
Press Ctrl+P to open profile selector:
╔═══════════════════════════════════╗
║  Select Configuration Profile     ║
╠═══════════════════════════════════╣
║ ● api1           (./api1.yml)     ║
║   production     (./production.yml)║
║   staging        (./staging.yml)  ║
╚═══════════════════════════════════╝
```

### Settings View
```
Press F3 to edit configuration:
⚙️  Settings - Profile: production

Method: PUT
URL Template: https://api.production.com/users/{{.id}}
Body Template: {...}

Headers:
  Authorization: Bearer prod-token
  X-API-Key: key-123

Ctrl+P: Switch Profile | Ctrl+S: Save
```

### Dynamic Workers
```
Press F4 to adjust workers:
👷 Workers Control

Workers: 6 / 16
[████████░░░░░░░░]

📊 Metrics
  Active:     6 workers
  Throughput: 45.2 req/s
  Total:      1234 requests
```

## 💡 Tips

1. **Multiple Environments**: Create one profile per environment (dev, staging, prod)
2. **Header Templates**: Use `{{.variable}}` in header values for dynamic headers
3. **Security**: Don't commit sensitive tokens to git! Use environment variables or secrets management
4. **Testing**: Start with low worker count (1-2) for testing, scale up for production

## 📖 Migration Guide

### Step 1: Create New Profile

Copy your existing `config.yml` to `api1.yml`:
```bash
cp config.yml api1.yml
```

### Step 2: Update Format

Change from:
```yaml
token: "your-token"
path:
  method: POST
  template: "..."
```

To:
```yaml
request:
  method: POST
  url_template: "..."
  headers:
    Authorization: "Bearer your-token"
```

### Step 3: Test

Run rapper - it will auto-discover your profiles!

---

**Version:** 2.0.0 - Profile Management Edition
**Date:** 2026-01-14
**Author:** Rapper Team
