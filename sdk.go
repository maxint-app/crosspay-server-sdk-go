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

func ptr[T any](v T) *T {
	return &v
}

// CrosspayServerClient is a high-level client for the Crosspay API
type CrosspayServerClient struct {
	client *Client
	apiKey string
}

type Environment string

const (
	// EnvironmentProduction is the production environment
	EnvironmentProduction Environment = "prod"
	// EnvironmentSandbox is the sandbox environment
	EnvironmentSandbox Environment = "sandbox"
)

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
func (c *CrosspayServerClient) ListEntitlements(ctx context.Context, environment Environment) ([]TenantEntitlement, error) {
	resp, err := c.client.GetTenantEntitlementsByEnvironment(ctx, string(environment))
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

// GetActiveProduct retrieves the active product for a customer
func (c *CrosspayServerClient) GetActiveProducts(ctx context.Context, customerEmail string, environment Environment) ([]TenantProduct, error) {
	activeEntitlements, err := c.GetActiveEntitlements(ctx, customerEmail, environment)
	if err != nil {
		return nil, err
	}
	if activeEntitlements == nil {
		return nil, nil
	}

	products, err := c.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	activeProducts := []TenantProduct{}

	for _, product := range products {
		for _, activeEntitlement := range activeEntitlements {
			if product.EntitlementId == activeEntitlement.Id {
				activeProducts = append(activeProducts, product)
			}
		}
	}

	return activeProducts, nil
}

// GetActiveEntitlements retrieves the active entitlement for a customer
func (c *CrosspayServerClient) GetActiveEntitlements(ctx context.Context, customerEmail string, environment Environment) ([]StorableEntitlement, error) {
	body := TenantActiveEntitlementsInputBody{
		CustomerEmail: customerEmail,
		Environment:   ptr(string(environment)),
	}
	resp, err := c.client.PostTenantEntitlementsActive(ctx, body)
	if err != nil {
		return nil, err
	}
	result, err := ParsePostTenantEntitlementsActiveResponse(resp)
	if err != nil {
		return nil, err
	}
	if result.JSON200.Error != nil && *result.JSON200.Error != "" {
		return nil, errors.New(*result.JSON200.Error)
	}

	return *result.JSON200.Data, nil
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
func (c *CrosspayServerClient) GetCustomerInfo(ctx context.Context, customerEmail string, environment Environment) (*CustomerEntitlements, error) {
	body := TenantServerGetCustomerInputBody{
		CustomerEmail: customerEmail,
		Environment:   ptr(string(environment)),
	}

	resp, err := c.client.PostTenantServerV2Customer(ctx, body)
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

	var result TenantServerGetCustomerV2ResponseBody
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.Error != nil && *result.Error != "" {
		return nil, errors.New(*result.Error)
	}

	return result.Data, nil
}

func (c *CrosspayServerClient) CancelStripeSubscription(
	ctx context.Context,
	environment Environment,
	entitlementId,
	customerEmail string,
) error {
	resp, err := c.client.PostTenantServerStripeCancelByEnvironment(ctx, string(environment), TenantCancelStripeSubscriptionInputBody{
		CustomerEmail: customerEmail,
		EntitlementId: entitlementId,
	})

	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *CrosspayServerClient) CancelGocardlessSubscription(
	ctx context.Context,
	environment Environment,
	entitlementId,
	customerEmail string,
) error {
	resp, err := c.client.PostTenantServerGocardlessCancelByEnvironment(ctx, string(environment), TenantCancelGoCardlessSubscriptionInputBody{
		CustomerEmail: customerEmail,
		EntitlementId: entitlementId,
	})

	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// ConstructWebhookEvent validates and parses a webhook event
func (c *CrosspayServerClient) ConstructWebhookEvent(
	webhookPublicKey string,
	rawPayload []byte,
	signatureHeader string,
	timestampHeader string,
) (*CustomerEntitlements, error) {
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
	var event CustomerEntitlements
	if err := json.Unmarshal(rawPayload, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	return &event, nil
}
