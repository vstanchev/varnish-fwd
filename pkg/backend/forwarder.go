package backend

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

type RequestForwarder struct {
	Provider Provider
	httpClient *http.Client
}

func NewRequestForwarder(provider Provider) *RequestForwarder {
  httpClient := &http.Client{
    Timeout: time.Second * 10,
  }
	return &RequestForwarder{Provider: provider, httpClient: httpClient}
}

func (f RequestForwarder) Handle(w http.ResponseWriter, r *http.Request)  {
  log.Printf("Got request: %s %s %v\n", r.Method, r.URL.String(), r.Header)
	globalStatus := http.StatusBadGateway
	backendUrls := f.Provider.GetBackendUrls()
	var statuses = make([]int, len(backendUrls))
	var err error
	for i, url := range backendUrls {
		// Append the current request path to the new url.
		newUrl := strings.TrimRight(url, "/") + r.URL.EscapedPath()
		var newRequest *http.Request
		if newRequest, err = http.NewRequest(r.Method, newUrl, r.Body); err != nil {
			log.Println(err)
			continue
		}
		// Set all headers from the original request.
		newRequest.Header = r.Header.Clone()
		newRequest.Header.Set("Connection", "keep-alive")

		log.Printf("Forwarding request: %s %s %v\n", newRequest.Method, newRequest.URL.String(), newRequest.Header)

		// Make the request to the backend and collect the response status code.
		if resp, err := f.httpClient.Do(newRequest); err == nil {
			log.Printf("Response: %s from %s\n", resp.Status, url)
			statuses[i] = resp.StatusCode
			if resp.StatusCode == http.StatusOK {
				globalStatus = http.StatusOK
			}
		} else {
			log.Println(err)
			statuses[i] = -1
		}
	}

	w.WriteHeader(globalStatus)
	globalResponse, _ := json.Marshal(map[string][]int{"statuses": statuses})
	w.Write(globalResponse)
}
