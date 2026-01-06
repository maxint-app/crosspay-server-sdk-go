package main

import (
	"context"
	"fmt"
	"log"

	crosspay "crosspay-server-sdk-go/src"
)

func main() {
	// Create a new Crosspay server client
	client, err := crosspay.NewCrosspayServerClient("your_api_key_here")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// List all products
	products, err := client.ListProducts(ctx)
	if err != nil {
		log.Printf("Error listing products: %v", err)
	} else {
		fmt.Println("Products:", products)
	}

	// List entitlements for production environment
	entitlements, err := client.ListEntitlements(ctx, "production")
	if err != nil {
		log.Printf("Error listing entitlements: %v", err)
	} else {
		fmt.Println("Entitlements:", entitlements)
	}

	// Get active subscription for a customer
	activeSubscription, err := client.GetActiveSubscription(ctx, "customer_email@example.com")
	if err != nil {
		log.Printf("Error getting active subscription: %v", err)
	} else {
		fmt.Println("Active Subscription:", activeSubscription)
	}

	// Get active product for a customer
	activeProduct, err := client.GetActiveProduct(ctx, "customer_email@example.com")
	if err != nil {
		log.Printf("Error getting active product: %v", err)
	} else {
		fmt.Println("Active Product:", activeProduct)
	}

	// Get active entitlement for a customer
	activeEntitlement, err := client.GetActiveEntitlement(ctx, "customer_email@example.com", "sandbox")
	if err != nil {
		log.Printf("Error getting active entitlement: %v", err)
	} else {
		fmt.Println("Active Entitlement:", activeEntitlement)
	}

	// Get customer info
	customerInfo, err := client.GetCustomerInfo(ctx, "customer_email@example.com")
	if err != nil {
		log.Printf("Error getting customer info: %v", err)
	} else {
		fmt.Println("Customer Info:", customerInfo)
	}

	// List customers with pagination
	limit := int64(20)
	customers, err := client.ListCustomers(ctx, &limit, nil)
	if err != nil {
		log.Printf("Error listing customers: %v", err)
	} else {
		fmt.Println("Customers:", customers)
	}

	// Example of webhook validation (commented out)
	/*
		webhookPublicKey := "-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----"
		rawPayload := []byte(`{"email":"user@example.com","id":"123"}`)
		signatureHeader := "base64_encoded_signature"
		timestampHeader := "2026-01-06T10:00:00Z"

		event, err := client.ConstructWebhookEvent(webhookPublicKey, rawPayload, signatureHeader, timestampHeader)
		if err != nil {
			log.Printf("Webhook validation failed: %v", err)
		} else {
			fmt.Println("Webhook Event:", event)
		}
	*/
}
