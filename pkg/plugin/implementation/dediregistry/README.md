# DeDi Registry Plugin

A **registry type plugin** for Beckn-ONIX that integrates with DeDi (Decentralized Digital Infrastructure) registry services. This plugin implements the `RegistryLookup` interface, making it a specialized type of registry plugin.

## Overview

The DeDi Registry plugin is a **registry implementation** that enables Beckn-ONIX to lookup participant records from DeDi registries. It converts DeDi API responses to standard Beckn Subscription format, allowing it to work interchangeably with the standard registry plugin through the same `RegistryLookup` interface.

## Plugin Type Classification

**Registry Type Plugin**: This plugin is a **type of registry plugin**, not a standalone plugin category.

- **Interface**: Implements `RegistryLookup` interface (same as standard registry plugin)
- **Interchangeable**: Can replace or work alongside standard registry plugin
- **Manager Access**: Available via `manager.Registry()` method
- **Plugin Category**: Registry

## Features

- **Standard Registry Interface**: Implements `RegistryLookup` interface for seamless integration
- **DeDi API Integration**: GET requests to DeDi registry endpoints with Bearer authentication
- **Dynamic Participant Lookup**: Uses subscriber IDs from request context (not static configuration)
- **Data Conversion**: Converts DeDi responses to standard Beckn Subscription format
- **HTTP Retry Logic**: Built-in retry mechanism using retryablehttp client
- **Timeout Control**: Configurable request timeouts


## Configuration

```yaml
plugins:
  dediRegistry:
    id: dediregistry
    config:
      baseURL: "https://dedi-registry.example.com"
      apiKey: "your-api-key"
      namespaceID: "beckn-network"
      registryName: "participants"
      timeout: "30"  # seconds
```

### Configuration Parameters

| Parameter | Required | Description | Default |
|-----------|----------|-------------|---------|
| `baseURL` | Yes | DeDi registry API base URL | - |
| `apiKey` | Yes | API key for authentication | - |
| `namespaceID` | Yes | DeDi namespace identifier | - |
| `registryName` | Yes | Registry name to query | - |
| `timeout` | No | Request timeout in seconds | 30 |

## Usage

### In Module Configuration

```yaml
modules:
  - name: bapTxnReceiver
    handler:
      plugins:
        registry:
          id: dediregistry
          config:
            baseURL: "https://dedi-registry.example.com"
            apiKey: "your-api-key"
            namespaceID: "beckn-network"
            registryName: "participants"
```

### In Code

```go
// Load DeDi registry plugin (same as any registry plugin)
dediRegistry, err := manager.Registry(ctx, &plugin.Config{
    ID: "dediregistry",  // Plugin ID specifies DeDi implementation
    Config: map[string]string{
        "baseURL": "https://dedi-registry.example.com",
        "apiKey": "your-api-key",
        "namespaceID": "beckn-network",
        "registryName": "participants",
    },
})

// Lookup participant with dynamic subscriber ID (from request context)
subscription := &model.Subscription{
    Subscriber: model.Subscriber{
        SubscriberID: "bap-network", // Extracted from Authorization header or request body
    },
}
results, err := dediRegistry.Lookup(ctx, subscription)
if err != nil {
    return err
}

// Extract public key from result (standard Beckn format)
if len(results) > 0 {
    publicKey := results[0].SigningPublicKey
    subscriberID := results[0].SubscriberID
}
```

## API Integration

### DeDi API URL Pattern
```
{baseURL}/dedi/lookup/{namespaceID}/{registryName}/{subscriberID}
```

**Example**: `https://dedi-registry.com/dedi/lookup/beckn-network/participants/bap-network`

### Expected DeDi Response Format

```json
{
  "message": "Resource retrieved successfully",
  "data": {
    "record_name": "bap.example.com",
    "details": {
      "entity_name": "BAP Example Provider",
      "entity_url": "https://bap.example.com",
      "publicKey": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...",
      "keyType": "RSA",
      "keyFormat": "PEM"
    },
    "state": "live",
    "created_at": "2025-09-23T07:45:10.357Z",
    "updated_at": "2025-09-23T07:51:39.923Z"
  }
}
```

### Converted to Standard Beckn Format

The plugin converts DeDi responses to standard Beckn Subscription format:

```json
{
  "subscriber_id": "bap.example.com",
  "url": "https://bap.example.com",
  "signing_public_key": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...",
  "status": "live",
  "created": "2025-09-23T07:45:10.357Z",
  "updated": "2025-09-23T07:51:39.923Z"
}
```

## Testing

Run plugin tests:

```bash
go test ./pkg/plugin/implementation/dediregistry -v
```

## Dependencies

- `github.com/hashicorp/go-retryablehttp`: HTTP client with retry logic
- Standard Go libraries for HTTP and JSON handling

## Plugin Architecture

### Registry Type Plugin Classification

```
Plugin Manager
├── Registry Plugins (RegistryLookup interface)
│   ├── registry (standard YAML-based registry)
│   └── dediregistry (DeDi API-based registry) ← This plugin
└── Other Plugin Types...
```

### Integration Notes

- **Plugin Type**: Registry implementation
- **Interface**: Implements `RegistryLookup` interface with `Lookup(ctx, *model.Subscription) ([]model.Subscription, error)`
- **Interchangeable**: Drop-in replacement for standard registry plugin
- **Manager Access**: Available via `manager.Registry()` method (same as standard registry)
- **Dynamic Lookup**: Uses `req.SubscriberID` from request context, not static configuration
- **Data Conversion**: Automatically converts DeDi API format to Beckn Subscription format
- **Build Integration**: Included in `build-plugins.sh`, compiles to `dediregistry.so`
- **Usage Pattern**: Configure with `id: dediregistry` in registry plugin configuration