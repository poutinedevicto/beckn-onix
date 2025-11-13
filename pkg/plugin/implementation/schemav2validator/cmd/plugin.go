package main

import (
	"context"
	"errors"
	"strconv"

	"github.com/beckn-one/beckn-onix/pkg/plugin/definition"
	"github.com/beckn-one/beckn-onix/pkg/plugin/implementation/schemav2validator"
)

// schemav2ValidatorProvider provides instances of schemav2Validator.
type schemav2ValidatorProvider struct{}

// New initialises a new Schemav2Validator instance.
func (vp schemav2ValidatorProvider) New(ctx context.Context, config map[string]string) (definition.SchemaValidator, func() error, error) {
	if ctx == nil {
		return nil, nil, errors.New("context cannot be nil")
	}

	url, ok := config["url"]
	if !ok || url == "" {
		return nil, nil, errors.New("url not configured")
	}

	cacheTTL := 3600
	if ttlStr, ok := config["cacheTTL"]; ok {
		if ttl, err := strconv.Atoi(ttlStr); err == nil && ttl > 0 {
			cacheTTL = ttl
		}
	}

	cfg := &schemav2validator.Config{
		URL:	url,
		CacheTTL:	cacheTTL,
	}

	return schemav2validator.New(ctx, cfg)
}

// Provider is the exported plugin provider.
var Provider schemav2ValidatorProvider
