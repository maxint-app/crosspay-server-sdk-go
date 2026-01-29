package crosspay

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CrosspayServerClient is a high-level client for the Crosspay API
type CrosspayServerClient struct {
	client *Client
	apiKey string
}

// NewCrosspayServerClient creates a new Crosspay server client
func NewCrosspayServerClient(apiKey string, baseURL ...string) (*CrosspayServerClient, error) {
	url := "https://api.crosspay.dev"
	if len(baseURL) > 0 && baseURL[0] != "" {
		url = baseURL[0]
	}

	client, err := NewClient(url, WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("api-key", apiKey)
		return nil
	}))
	if err != nil {
		return nil, err
	}

	return &CrosspayServerClient{
		client: client,
		apiKey: apiKey,
	}, nil
}

// ListProducts retrieves all tenant products
func (c *CrosspayServerClient) ListProducts(ctx context.Context) ([]TenantProduct, error) {
	resp, err := c.client.GetTenantProducts(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result TenantListProductsResponseBody
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Error != nil && *result.Error != "" {
		return nil, errors.New(*result.Error)
	}

	if result.Data == nil {
		return []TenantProduct{}, nil
	}

	return *result.Data, nil
}

// ListEntitlements retrieves all tenant entitlements for the specified environment
func (c *CrosspayServerClient) ListEntitlements(ctx context.Context, environment string) ([]TenantEntitlement, error) {
	resp, err := c.client.GetTenantEntitlementsByEnvironment(ctx, environment)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result TenantListEntitlementsResponseBody
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Error != nil && *result.Error != "" {
		return nil, errors.New(*result.Error)
	}

	if result.Data == nil {
		return []TenantEntitlement{}, nil
	}

	return *result.Data, nil
}

// GetActiveSubscription retrieves the active subscription for a customer
func (c *CrosspayServerClient) GetActiveSubscription(ctx context.Context, customerEmail string) (*StorableSubscription, error) {
	body := TenantActiveSubscriptionInputBody{
		CustomerEmail: customerEmail,
	}

	resp, err := c.client.PostTenantSubscriptionsActive(ctx, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result TenantActiveSubscriptionResponseBody
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.Error != nil && *result.Error != "" {
		return nil, errors.New(*result.Error)
	}

	return result.Data, nil
}

// GetActiveProduct retrieves the active product for a customer
func (c *CrosspayServerClient) GetActiveProduct(ctx context.Context, customerEmail string) (*TenantProduct, error) {
	activeSubscription, err := c.GetActiveSubscription(ctx, customerEmail)
	if err != nil {
		return nil, err
	}
	if activeSubscription == nil {
		return nil, nil
	}

	products, err := c.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	for _, product := range products {
		if product.ProductId == activeSubscription.ProductId {
			return &product, nil
		}
	}

	return nil, nil
}

// GetActiveEntitlement retrieves the active entitlement for a customer
func (c *CrosspayServerClient) GetActiveEntitlement(ctx context.Context, customerEmail, environment string) (*TenantEntitlement, error) {
	activeProduct, err := c.GetActiveProduct(ctx, customerEmail)
	if err != nil {
		return nil, err
	}
	if activeProduct == nil {
		return nil, nil
	}

	entitlements, err := c.ListEntitlements(ctx, environment)
	if err != nil {
		return nil, err
	}

	for _, entitlement := range entitlements {
		if entitlement.Id == activeProduct.EntitlementId {
			return &entitlement, nil
		}
	}

	return nil, nil
}

// ListCustomers retrieves a paginated list of customers
func (c *CrosspayServerClient) ListCustomers(ctx context.Context, limit *int64, cursor *string) (*ListCustomerResponseBody, error) {
	params := &GetTenantServerCustomersParams{
		Limit:  limit,
		Cursor: cursor,
	}

	resp, err := c.client.GetTenantServerCustomers(ctx, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result ListCustomerResponseBody
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Error != nil && *result.Error != "" {
		return nil, errors.New(*result.Error)
	}

	return &result, nil
}

// GetCustomerInfo retrieves extended customer information
func (c *CrosspayServerClient) GetCustomerInfo(ctx context.Context, customerEmail string) (*GetCustomerExtendedInfoByEmailRow, error) {
	body := TenantServerGetCustomerInputBody{
		CustomerEmail: customerEmail,
	}

	resp, err := c.client.PostTenantServerCustomer(ctx, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result TenantServerGetCustomerResponseBody
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.Error != nil && *result.Error != "" {
		return nil, errors.New(*result.Error)
	}

	return result.Data, nil
}

// ConstructWebhookEvent validates and parses a webhook event
func (c *CrosspayServerClient) ConstructWebhookEvent(
	webhookPublicKey string,
	rawPayload []byte,
	signatureHeader string,
	timestampHeader string,
) (*GetCustomerExtendedInfoByEmailRow, error) {
	// Parse the timestamp
	timestampDate, err := time.Parse(time.RFC3339, timestampHeader)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp header: %w", err)
	}

	// Check if within 5-minute window
	timeDiff := time.Since(timestampDate)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}
	if timeDiff > 5*time.Minute {
		return nil, errors.New("timestamp is outside the 5-minute window")
	}

	// Construct the data to verify
	data := append([]byte(timestampHeader), '.')
	data = append(data, rawPayload...)

	// Decode the signature
	signature, err := base64.StdEncoding.DecodeString(signatureHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	// Parse the public key
	block, _ := pem.Decode([]byte(webhookPublicKey))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	// Verify the signature based on key type
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not RSA")
	}

	// Create a hash of the data
	hash := sha256.Sum256(data)

	// Verify using RSA PSS or PKCS1v15
	err = rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, hash[:], signature)
	if err != nil {
		return nil, errors.New("signature verification failed")
	}

	// Parse and return the payload
	var event GetCustomerExtendedInfoByEmailRow
	if err := json.Unmarshal(rawPayload, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	return &event, nil
}
