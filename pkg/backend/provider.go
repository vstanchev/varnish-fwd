package backend

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const backendsEnv = "FWD_BACKENDS"
const backendsEnvSeparator = ","

// A common interface representing a backend provider.
//
// GetBackendUrls should return all backends that the provider provides
// with an option to reset the internal provider cache (if supported).
type Provider interface {
	GetBackendUrls(resetCache bool) []string
}

// Simple implementation of a Provider that has a static list of backends.
type StaticProvider struct {
	backendUrls []string
}

// Returns all backend URLs.
func (s *StaticProvider) GetBackendUrls(resetCache bool) []string {
	return s.backendUrls
}

// A constructor for the static backend provider.
func NewStaticProvider(backendUrls []string) Provider {
	return &StaticProvider{
		backendUrls: backendUrls,
	}
}

// Returns a new static backend provider initialized from environment variables.
func NewStaticProviderFromEnv() (Provider, error) {
	backendsList := os.Getenv(backendsEnv)
	if backendsList == "" {
		return nil, errors.New(fmt.Sprintf("The %s environment variable is empty or not set.", backendsEnv))
	}
	return NewStaticProvider(strings.Split(backendsList, backendsEnvSeparator)), nil
}
