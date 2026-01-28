# Crosspay Server SDK for Go

A Go client library for the Crosspay API, providing easy access to subscription management, customer information, and webhook validation.

## Installation

```bash
go get github.com/maxint-app/crosspay-server-sdk-go
```

## Usage

### Initialize the Client

```go
import (
    "context"
    crosspay "github.com/maxint-app/crosspay-server-sdk-go/src"
)

client, err := crosspay.NewCrosspayServerClient("your_api_key_here")
if err != nil {
    log.Fatal(err)
}

// Or with a custom base URL
client, err := crosspay.NewCrosspayServerClient("your_api_key_here", "https://custom-api.example.com")
```

### List Products

```go
ctx := context.Background()
products, err := client.ListProducts(ctx)
if err != nil {
    log.Fatal(err)
}
```

### List Entitlements

```go
entitlements, err := client.ListEntitlements(ctx, "production") // or "sandbox"
if err != nil {
    log.Fatal(err)
}
```

### Get Active Subscription

```go
subscription, err := client.GetActiveSubscription(ctx, "customer@example.com")
if err != nil {
    log.Fatal(err)
}
```

### Get Active Product

```go
product, err := client.GetActiveProduct(ctx, "customer@example.com")
if err != nil {
    log.Fatal(err)
}
```

### Get Active Entitlement

```go
entitlement, err := client.GetActiveEntitlement(ctx, "customer@example.com", "production")
if err != nil {
    log.Fatal(err)
}
```

### List Customers

```go
limit := int64(20)
cursor := "optional_cursor_string"
customers, err := client.ListCustomers(ctx, &limit, &cursor)
if err != nil {
    log.Fatal(err)
}
```

### Get Customer Info

```go
customerInfo, err := client.GetCustomerInfo(ctx, "customer@example.com")
if err != nil {
    log.Fatal(err)
}
```

### Validate Webhook Events

```go
webhookPublicKey := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...
-----END PUBLIC KEY-----`

rawPayload := []byte(`{"email":"user@example.com","id":"123"}`)
signatureHeader := "base64_encoded_signature"
timestampHeader := "2026-01-06T10:00:00Z"

event, err := client.ConstructWebhookEvent(
    webhookPublicKey,
    rawPayload,
    signatureHeader,
    timestampHeader,
)
if err != nil {
    log.Printf("Webhook validation failed: %v", err)
} else {
    fmt.Printf("Valid webhook event: %+v\n", event)
}
```

## Types

The SDK exports the following types from the generated package:

- `TenantProduct` - Product information
- `TenantEntitlement` - Entitlement details
- `StorableSubscription` - Subscription data
- `GetCustomerExtendedInfoByEmailRow` - Extended customer information
- `ListCustomerResponseBody` - Paginated customer list response

## Error Handling

All methods return errors that should be checked. API errors are wrapped with descriptive messages.

```go
customers, err := client.ListCustomers(ctx, nil, nil)
if err != nil {
    log.Printf("Failed to list customers: %v", err)
    return
}
```

## Context Support

All API methods accept a `context.Context` parameter for cancellation and timeout support:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

products, err := client.ListProducts(ctx)
```

## Example

See the [example](./example/example.go) directory for a complete usage example.

## License

See LICENSE file for details.
