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

func (f *RequestForwarder) Handle(w http.ResponseWriter, r *http.Request)  {
  log.Printf("Got request: %s %s %v\n", r.Method, r.URL.String(), r.Header)
	globalStatus := http.StatusBadGateway
	backendUrls := f.Provider.GetBackendUrls(false)
	var statuses = make([]int, len(backendUrls))
	var err error
  refreshBackends := false
  for i, backendUrl := range backendUrls {
		// Append the current request path to the new backendUrl.
		newUrl := strings.TrimRight(backendUrl, "/") + r.URL.EscapedPath()
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
		  var body []byte
		  resp.Body.Read(body)
		  resp.Body.Close()
			log.Printf("Response: %s from %s\n", resp.Status, backendUrl)
			statuses[i] = resp.StatusCode
			if resp.StatusCode == http.StatusOK {
				globalStatus = http.StatusOK
			}
		} else {
			log.Printf("Backend request error: %s\n", err.Error())
			statuses[i] = -1
      refreshBackends = true
		}
	}

  // Refresh the backend URLs in case of a timeout and return an error so that the next request can succeed.
  if refreshBackends {
    globalStatus = http.StatusBadGateway
    log.Printf("Refreshed backends %s\n", strings.Join(f.Provider.GetBackendUrls(true), ", "))
  }

  w.WriteHeader(globalStatus)
	globalResponse, _ := json.Marshal(map[string][]int{"statuses": statuses})
	w.Write(globalResponse)
}
