# DeDi Registry Plugin

A **registry type plugin** for Beckn-ONIX that integrates with DeDi (Decentralized Digital Infrastructure) registry services.

## Overview

The DeDi Registry plugin is a **registry implementation** that enables Beckn-ONIX to lookup participant records from remote DeDi registries via REST API calls.

## Configuration

```yaml
plugins:
  registry:
    id: dediregistry
    config:
      baseURL: "https://dedi-api.example.com"
      apiKey: "your-bearer-token"
      namespaceID: "76EU8BF1gzRGGatgw7wZZb7nEVx77XSwkKDv4UDLdxh8ztty4zmbYU"
      registryName: "dedi_registry"
      timeout: "30"  # seconds
```

### Configuration Parameters

| Parameter | Required | Description | Default |
|-----------|----------|-------------|---------|
| `baseURL` | Yes | DeDi registry API base URL | - |
| `apiKey` | Yes | Bearer token for API authentication | - |
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

**Example**: `https://dedi-api.com/dedi/lookup/76EU8BF1gzRGGatgw7wZZb7nEVx77XSwkKDv4UDLdxh8ztty4zmbYU/dedi_registry/bap-network`

### Authentication
```
Authorization: Bearer {apiKey}
```

### Expected DeDi Response Format

```json
{
  "message": "Resource retrieved successfully",
  "data": {
    "namespace": "dediregistry",
    "namespace_id": "76EU8BF1gzRGGatgw7wZZb7nEVx77XSwkKDv4UDLdxh8ztty4zmbYU",
    "registry_name": "dedi_registry",
    "record_name": "bap-network",
    "details": {
      "key_id": "b692d295-5425-40f5-af77-d62646841dca",
      "signing_public_key": "YK3Xqc83Bpobc1UT0ObAe6mBJMiAOkceIsNtmph9WTc=",
      "encr_public_key": "YK3Xqc83Bpobc1UT0ObAe6mBJMiAOkceIsNtmph9WTc=",
      "status": "SUBSCRIBED",
      "created": "2024-01-15T10:00:00Z",
      "updated": "2024-01-15T10:00:00Z",
      "valid_from": "2024-01-01T00:00:00Z",
      "valid_until": "2025-12-31T23:59:59Z"
    },
    "state": "live",
    "created_at": "2025-10-09T06:09:48.295Z"
  }
}
```

### Field Mapping to Beckn Subscription

| DeDi Field | Beckn Field | Description |
|------------|-------------|-------------|
| `data.record_name` | `subscriber_id` | Participant identifier |
| `data.details.key_id` | `key_id` | Unique key identifier |
| `data.details.signing_public_key` | `signing_public_key` | Public key for signature verification |
| `data.details.encr_public_key` | `encr_public_key` | Public key for encryption |
| `data.details.status` | `status` | Subscription status |
| `data.details.created` | `created` | Creation timestamp |
| `data.details.updated` | `updated` | Last update timestamp |
| `data.details.valid_from` | `valid_from` | Key validity start |
| `data.details.valid_until` | `valid_until` | Key validity end |

### Lookup Flow

1. **Request Processing**: Plugin extracts `subscriber_id` from incoming request
2. **API Call**: Makes GET request to DeDi API with static config + dynamic subscriber ID
3. **Response Parsing**: Extracts data from `data.details` object
4. **Format Conversion**: Maps DeDi fields to Beckn Subscription format
5. **Return**: Returns array of Subscription objects

## Testing

Run plugin tests:

```bash
go test ./pkg/plugin/implementation/dediregistry -v
```

Test coverage includes:
- Configuration validation
- Successful API responses
- HTTP error handling
- Network failures
- Invalid JSON responses
- Missing required fields

## Dependencies

- `github.com/hashicorp/go-retryablehttp`: HTTP client with retry logic
- Standard Go libraries for HTTP and JSON handling

## Error Handling

- **Configuration Errors**: Missing required config parameters
- **Network Errors**: Connection failures, timeouts
- **HTTP Errors**: Non-200 status codes from DeDi API
- **Data Errors**: Missing required fields in response
- **Validation Errors**: Empty subscriber ID in request


### Integration Notes

- **Plugin Type**: Registry implementation
- **Interface**: Implements `RegistryLookup` interface with `Lookup(ctx, *model.Subscription) ([]model.Subscription, error)`
- **Manager Access**: Available via `manager.Registry()` method (same as standard registry)
- **Dynamic Lookup**: Uses `req.SubscriberID` from request context, not static configuration
- **Data Conversion**: Automatically converts DeDi API format to Beckn Subscription format
- **Build Integration**: Included in `build-plugins.sh`, compiles to `dediregistry.so`
- **Usage Pattern**: Configure with `id: dediregistry` in registry plugin configuration