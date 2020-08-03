package backend

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const backendsEnv = "FWD_BACKENDS"
const backendsEnvSeparator = ","

type Provider interface {
	GetBackendUrls(resetCache bool) []string
}

type StaticProvider struct {
	backendUrls []string
}

func (s *StaticProvider) GetBackendUrls(resetCache bool) []string {
	return s.backendUrls
}

func NewStaticProvider(backendUrls []string) Provider {
	return &StaticProvider{
		backendUrls: backendUrls,
	}
}

func NewStaticProviderFromEnv() (Provider, error) {
	backendsList := os.Getenv(backendsEnv)
	if backendsList == "" {
		return nil, errors.New(fmt.Sprintf("The %s environment variable is empty or not set.", backendsEnv))
	}
	return NewStaticProvider(strings.Split(backendsList, backendsEnvSeparator)), nil
}
