package backend

import (
  "encoding/json"
  "log"
  "net/http"
  "os"
  "strings"
  "sync"
  "time"
)

const refreshIntervalEnv = "FWD_REFRESH_INTERVAL"

type RequestForwarder struct {
	Provider Provider
	httpClient *http.Client
	refreshMutex *sync.Mutex
}

func NewRequestForwarder(provider Provider) *RequestForwarder {
  httpClient := &http.Client{
    Timeout: time.Second * 10,
  }

	forwarder := &RequestForwarder{Provider: provider, httpClient: httpClient, refreshMutex: &sync.Mutex{}}

	// Parse duration from environment variable or use default.
	durationEnv := "5m"
	if durationValue, ok := os.LookupEnv(refreshIntervalEnv); ok {
	  durationEnv = durationValue
  }
  parsedDuration, err := time.ParseDuration(durationEnv)
  if err != nil {
    log.Fatalf("Cannot parse duration: %v", err)
  }

  // Start ticker if parsed duration is not zero.
  if parsedDuration != time.Duration(0) {
    forwarder.startRefreshTicker(parsedDuration)
  }
	return forwarder
}

// Refresh backends on a specified interval.
func (f *RequestForwarder) startRefreshTicker(duration time.Duration) {
  ticker := time.NewTicker(duration)

  go func() {
    for range ticker.C {
      f.processRefreshTicker()
    }
  }()
}

func (f *RequestForwarder) processRefreshTicker() {
  // Convert the old backends slice to a map for faster lookup.
  oldBackendsMap := make(map[string]struct{})
  for _, backend := range f.Provider.GetBackendUrls(false) {
    oldBackendsMap[backend] = struct{}{}
  }

  hasNew := false
  newBackends := f.refreshBackends()
  for _, backend := range newBackends {
    if _, exists := oldBackendsMap[backend]; !exists {
      hasNew = true
      continue
    }
  }

  if hasNew {
    log.Println("Detected new backends")
    f.purgeEverything()
  }
}

func (f *RequestForwarder) purgeEverything() {
  log.Println("Sending a BAN request to all backends")
  var wg sync.WaitGroup
  for _, backendUrl := range f.Provider.GetBackendUrls(false) {
    purgeRequest, err := http.NewRequest("BAN", backendUrl, strings.NewReader(""))
    if err != nil {
      log.Println(err)
    }
    purgeRequest.Header.Set("Cache-Tags", "http_response")

    wg.Add(1)
    go func(backendUrl string) {
      statusCode, err := f.makeBackendRequest(purgeRequest)
      if err != nil {
        log.Printf("Backend request error: %s\n", err.Error())
      }
      log.Printf("Backend %s responded with %d\n", backendUrl, statusCode)
      wg.Done()
    }(backendUrl)
  }
  wg.Wait()
}

// Calls the GetBackendUrls method with the reset cache option and prints the refreshed backends.
func (f *RequestForwarder) refreshBackends() []string {
  f.refreshMutex.Lock()
  defer f.refreshMutex.Unlock()
  backends := f.Provider.GetBackendUrls(true)
  log.Printf("Refreshed backends %s\n", strings.Join(backends, ", "))
  return backends
}

func (f *RequestForwarder) createNewRequest(backendUrl string, oldRequest *http.Request) (*http.Request, error) {
  // Append the current request path to the new backendUrl.
  newUrl := strings.TrimRight(backendUrl, "/") + oldRequest.URL.EscapedPath()
  newRequest, err := http.NewRequest(oldRequest.Method, newUrl, oldRequest.Body)
  if err != nil {
    return nil, err
  }
  // Set all headers from the original request.
  newRequest.Header = oldRequest.Header.Clone()
  newRequest.Header.Set("Connection", "keep-alive")

  return newRequest, nil
}

// Executes the given request and returns only the status code.
// In case of an error it returns -1 as the status code and the error itself.
func (f *RequestForwarder) makeBackendRequest(r *http.Request) (int, error) {
  // Make the request to the backend and collect only the response status code.
  resp, err := f.httpClient.Do(r)
  if err != nil {
    return -1, err
  }

  var body []byte
  resp.Body.Read(body)
  resp.Body.Close()
  return resp.StatusCode, nil
}

func (f *RequestForwarder) Handle(w http.ResponseWriter, r *http.Request)  {
  log.Printf("Got request: %s %s %v\n", r.Method, r.URL.String(), r.Header)

	globalStatus := http.StatusBadGateway
	backendUrls := f.Provider.GetBackendUrls(false)
  refreshBackends := false

	var statuses = make([]int, len(backendUrls))
  var wg sync.WaitGroup

  for i, backendUrl := range backendUrls {

    newRequest, err := f.createNewRequest(backendUrl, r)
    if err != nil {
      log.Println(err)
      continue
    }

		log.Printf("Forwarding request: %s %s %v\n", newRequest.Method, newRequest.URL.String(), newRequest.Header)
    wg.Add(1)
    go func(i int) {
      statuses[i], err = f.makeBackendRequest(newRequest)
      if err != nil {
        log.Printf("Backend request error: %s\n", err.Error())
        refreshBackends = true
      }

      if statuses[i] == http.StatusOK {
        globalStatus = http.StatusOK
      }
      wg.Done()
    }(i)
	}

	wg.Wait()

  // Refresh the backend URLs in case of a timeout and return an error so that the client is aware that it should retry.
  if refreshBackends {
    globalStatus = http.StatusBadGateway
    go f.refreshBackends()
  }

  w.WriteHeader(globalStatus)
	globalResponse, _ := json.Marshal(map[string][]int{"statuses": statuses})
	w.Write(globalResponse)
}
