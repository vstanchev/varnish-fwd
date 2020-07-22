package main

import (
	"log"
	"net/http"
  "net/url"
  "os"
	"strconv"
  "strings"
  "varnish-fwd/pkg/backend"
)

const backendsProviderEnv = "FWD_PROVIDER"
const portEnv = "FWD_PORT"

func main() {
	// @todo connection keep-alive and timeouts
	// @todo logging of all requests and responses
	// @todo create a struct to hold all app config and parsing

	// Validate port
	listenPort := "6081"
	if port, ok := os.LookupEnv(portEnv); ok {
		if _, portErr := strconv.Atoi(port); portErr != nil {
			log.Fatalf("Specified port %s is not valid", port)
		} else {
			listenPort = port
		}
	}

	// Choose a provider implementation based on config.
	providerType := "ecs"
  if provider, ok := os.LookupEnv(backendsProviderEnv); ok {
    providerType = provider
  }
	var err error
	var provider backend.Provider

	switch providerType {
	case "static":
		provider, err = backend.NewStaticProviderFromEnv()
	case "ecs":
		provider, err = backend.NewAwsEcsProviderFromEnv()
	default:
		log.Fatalf("No provider defined by env variable %s", backendsProviderEnv)
	}

	// Error handling for provider initialization.
	if err != nil {
		log.Fatal(err)
	}

	// Validate and log all the backend URLs.
	validateUrls := provider.GetBackendUrls()
	if len(validateUrls) == 0 {
	  log.Fatal("No backends provided!")
  }

	for _, backendUrl := range validateUrls {
    if _, err := url.ParseRequestURI(backendUrl); err != nil {
      log.Fatalf("Backend URL %s is not valid!\n", backendUrl)
    }
  }
	log.Printf("Initialized a %s backend provider with backends [%s].\n", providerType, strings.Join(validateUrls, ", "))

	// Initialize a forwarder with the backends provider and start listening for requests.
	forwarder := backend.NewRequestForwarder(provider)
	http.HandleFunc("/", forwarder.Handle)
	log.Printf("Starting server on port %s", listenPort)
	log.Fatal(http.ListenAndServe(":" +listenPort, nil))
}
