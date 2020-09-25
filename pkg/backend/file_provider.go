package backend

import (
  "errors"
  "fmt"
  "io/ioutil"
  "log"
  "os"
  "strings"
)

const backendsFileEnv = "FWD_BACKENDS_FILE"
const backendsFileEnvSeparator = ","

type FileProvider struct {
  filePath string
  backendsCached []string
}

func (f *FileProvider) GetBackendUrls(resetCache bool) []string {
  if !resetCache && len(f.backendsCached) > 0 {
    return f.backendsCached
  }

  // Read the contents of the backend file.
  contents, err := ioutil.ReadFile(f.filePath)
  if err != nil {
    log.Fatalf("%v\n", err)
  }

  // Split the string.
  f.backendsCached = strings.Split(string(contents), backendsFileEnvSeparator)

  // Trim all spaces.
  for i := range f.backendsCached {
    f.backendsCached[i] = strings.TrimSpace(f.backendsCached[i])
  }

  log.Println("Read from file")

  return f.backendsCached
}

func NewFileProvider(filePath string) Provider {
  return &FileProvider{
    filePath:       filePath,
  }
}

func NewFileProviderFromEnv() (Provider, error) {
  backendsFile := os.Getenv(backendsFileEnv)
  if backendsFile == "" {
    return nil, errors.New(fmt.Sprintf("The %s environment variable is empty or not set.", backendsFileEnv))
  }
  return NewFileProvider(backendsFile), nil
}
