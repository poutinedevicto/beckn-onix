package dediregistry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beckn-one/beckn-onix/pkg/model"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "empty baseURL",
			config: &Config{
				BaseURL:      "",
				ApiKey:       "test-key",
				NamespaceID:  "test-namespace",
				RegistryName: "test-registry",
			},
			wantErr: true,
		},
		{
			name: "empty apiKey",
			config: &Config{
				BaseURL:      "https://test.com",
				ApiKey:       "",
				NamespaceID:  "test-namespace",
				RegistryName: "test-registry",
			},
			wantErr: true,
		},
		{
			name: "empty namespaceID",
			config: &Config{
				BaseURL:      "https://test.com",
				ApiKey:       "test-key",
				NamespaceID:  "",
				RegistryName: "test-registry",
			},
			wantErr: true,
		},
		{
			name: "empty registryName",
			config: &Config{
				BaseURL:      "https://test.com",
				ApiKey:       "test-key",
				NamespaceID:  "test-namespace",
				RegistryName: "",
			},
			wantErr: true,
		},
		{
			name: "valid config",
			config: &Config{
				BaseURL:      "https://test.com",
				ApiKey:       "test-key",
				NamespaceID:  "test-namespace",
				RegistryName: "test-registry",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNew(t *testing.T) {
	ctx := context.Background()

	validConfig := &Config{
		BaseURL:      "https://test.com",
		ApiKey:       "test-key",
		NamespaceID:  "test-namespace",
		RegistryName: "test-registry",
		Timeout:      30,
	}

	client, closer, err := New(ctx, validConfig)
	if err != nil {
		t.Errorf("New() error = %v", err)
		return
	}

	if client == nil {
		t.Error("New() returned nil client")
	}

	if closer == nil {
		t.Error("New() returned nil closer")
	}

	// Test cleanup
	if err := closer(); err != nil {
		t.Errorf("closer() error = %v", err)
	}
}

func TestLookup(t *testing.T) {
	ctx := context.Background()

	// Test successful lookup
	t.Run("successful lookup", func(t *testing.T) {
		// Mock server with successful response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request method and path
			if r.Method != "GET" {
				t.Errorf("Expected GET request, got %s", r.Method)
			}
			if r.URL.Path != "/dedi/lookup/test-namespace/test-registry/bap-network" {
				t.Errorf("Unexpected path: %s", r.URL.Path)
			}
			// Verify Authorization header
			if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
				t.Errorf("Expected Bearer test-key, got %s", auth)
			}

			// Return mock response using actual DeDI format
			response := map[string]interface{}{
				"message": "Resource retrieved successfully",
				"data": map[string]interface{}{
					"record_name": "bap-network",
					"details": map[string]interface{}{
						"entity_name": "BAP Network Provider",
						"entity_url":  "https://bap-network.example.com",
						"publicKey":   "test-public-key",
						"keyType":     "ed25519",
						"keyFormat":   "base64",
					},
					"state":      "live",
					"created_at": "2023-01-01T00:00:00Z",
					"updated_at": "2023-01-01T00:00:00Z",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		config := &Config{
			BaseURL:      server.URL,
			ApiKey:       "test-key",
			NamespaceID:  "test-namespace",
			RegistryName: "test-registry",
			Timeout:      30,
		}

		client, closer, err := New(ctx, config)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		defer closer()

		req := &model.Subscription{
			Subscriber: model.Subscriber{
				SubscriberID: "bap-network",
			},
		}
		results, err := client.Lookup(ctx, req)
		if err != nil {
			t.Errorf("Lookup() error = %v", err)
			return
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
			return
		}

		subscription := results[0]
		if subscription.Subscriber.SubscriberID != "bap-network" {
			t.Errorf("Expected subscriber_id bap-network, got %s", subscription.Subscriber.SubscriberID)
		}
		if subscription.SigningPublicKey != "test-public-key" {
			t.Errorf("Expected signing_public_key test-public-key, got %s", subscription.SigningPublicKey)
		}
		if subscription.Status != "live" {
			t.Errorf("Expected status live, got %s", subscription.Status)
		}
	})

	// Test empty subscriber ID
	t.Run("empty subscriber ID", func(t *testing.T) {
		config := &Config{
			BaseURL:      "https://test.com",
			ApiKey:       "test-key",
			NamespaceID:  "test-namespace",
			RegistryName: "test-registry",
		}

		client, closer, err := New(ctx, config)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		defer closer()

		req := &model.Subscription{
			Subscriber: model.Subscriber{
				SubscriberID: "",
			},
		}
		_, err = client.Lookup(ctx, req)
		if err == nil {
			t.Error("Expected error for empty subscriber ID, got nil")
		}
		if err.Error() != "subscriber_id is required for DeDi lookup" {
			t.Errorf("Expected specific error message, got %v", err)
		}
	})

	// Test HTTP error response
	t.Run("http error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Record not found"))
		}))
		defer server.Close()

		config := &Config{
			BaseURL:      server.URL,
			ApiKey:       "test-key",
			NamespaceID:  "test-namespace",
			RegistryName: "test-registry",
		}

		client, closer, err := New(ctx, config)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		defer closer()

		req := &model.Subscription{
			Subscriber: model.Subscriber{
				SubscriberID: "bap-network",
			},
		}
		_, err = client.Lookup(ctx, req)
		if err == nil {
			t.Error("Expected error for 404 response, got nil")
		}
	})

	// Test missing required fields
	t.Run("missing entity_name", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"details": map[string]interface{}{
						"entity_url": "https://test.example.com",
						"publicKey":  "test-public-key",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		config := &Config{
			BaseURL:      server.URL,
			ApiKey:       "test-key",
			NamespaceID:  "test-namespace",
			RegistryName: "test-registry",
		}

		client, closer, err := New(ctx, config)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		defer closer()

		req := &model.Subscription{
			Subscriber: model.Subscriber{
				SubscriberID: "bap-network",
			},
		}
		_, err = client.Lookup(ctx, req)
		if err == nil {
			t.Error("Expected error for missing details field, got nil")
		}
	})

	// Test invalid JSON response
	t.Run("invalid json response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		config := &Config{
			BaseURL:      server.URL,
			ApiKey:       "test-key",
			NamespaceID:  "test-namespace",
			RegistryName: "test-registry",
		}

		client, closer, err := New(ctx, config)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		defer closer()

		req := &model.Subscription{
			Subscriber: model.Subscriber{
				SubscriberID: "bap-network",
			},
		}
		_, err = client.Lookup(ctx, req)
		if err == nil {
			t.Error("Expected error for invalid JSON, got nil")
		}
	})

	// Test network error
	t.Run("network error", func(t *testing.T) {
		config := &Config{
			BaseURL:      "http://invalid-url-that-does-not-exist.local",
			ApiKey:       "test-key",
			NamespaceID:  "test-namespace",
			RegistryName: "test-registry",
			Timeout:      1,
		}

		client, closer, err := New(ctx, config)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		defer closer()

		req := &model.Subscription{
			Subscriber: model.Subscriber{
				SubscriberID: "bap-network",
			},
		}
		_, err = client.Lookup(ctx, req)
		if err == nil {
			t.Error("Expected network error, got nil")
		}
	})

	// Test missing data field
	t.Run("missing data field", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := map[string]interface{}{
				"message": "success",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		config := &Config{
			BaseURL:      server.URL,
			ApiKey:       "test-key",
			NamespaceID:  "test-namespace",
			RegistryName: "test-registry",
		}

		client, closer, err := New(ctx, config)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		defer closer()

		req := &model.Subscription{
			Subscriber: model.Subscriber{
				SubscriberID: "bap-network",
			},
		}
		_, err = client.Lookup(ctx, req)
		if err == nil {
			t.Error("Expected error for missing data field, got nil")
		}
	})
}