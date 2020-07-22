package backend

import (
  "errors"
  "fmt"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/credentials"
  "github.com/aws/aws-sdk-go/aws/credentials/stscreds"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/ecs"
  "log"
  "os"
  "varnish-fwd/pkg/token"
)

const serviceNameEnv = "VARNISH_SERVICE_NAME"
const clusterEnv = "CLUSTER_NAME"
const backendSchemeEnv = "VARNISH_SCHEME"
const backendPortEnv = "VARNISH_PORT"

type AwsEcsProvider struct {
	ServiceName    string
	Cluster        string
	ecsClient      *ecs.ECS
	backendsCached []string
	backendScheme string
	backendPort string
}

// The following environment variables are available:
//  - AWS_PROFILE (optional): the profile to use when making requests to AWS
//  - AWS_IAM_AUTHENTICATOR_CACHE_FILE (optional): full path to the file where session credentials will be cached
func NewAwsEcsProvider(serviceName, cluster, backendScheme, backendPort string) Provider {
	provider := &AwsEcsProvider{
		ServiceName: serviceName,
		Cluster:     cluster,
		backendPort: backendPort,
		backendScheme: backendScheme,
	}
	provider.InitializeEcsClient()
	return provider
}

func NewAwsEcsProviderFromEnv() (Provider, error) {
	serviceName := os.Getenv(serviceNameEnv)
	cluster := os.Getenv(clusterEnv)
	if serviceName == "" || cluster == "" {
		return nil, errors.New(fmt.Sprintf("The %s and %s environment variables are empty or not set.", serviceNameEnv, clusterEnv))
	}

	backendScheme := "http"
	backendPort := "6081"
	if len(os.Getenv(backendSchemeEnv)) > 0 {
		backendScheme = os.Getenv(backendSchemeEnv)
	}
	if len(os.Getenv(backendPortEnv)) > 0 {
		backendPort = os.Getenv(backendPortEnv)
	}

	return NewAwsEcsProvider(serviceName, cluster, backendScheme, backendPort), nil
}

func (p *AwsEcsProvider) InitializeEcsClient() {
	awsConfig := aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
	}
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
		SharedConfigState:       session.SharedConfigEnable,
		Config:                  awsConfig,
	}))

	// The first three arguments are used for cache key. Since we support only one AWS account and one cache they can
	// be hardcoded strings.
	if cacheProvider, err := token.NewFileCacheProvider("default", "default", "default", sess.Config.Credentials); err == nil {
		sess.Config.Credentials = credentials.NewCredentials(&cacheProvider)
	} else {
		log.Fatalf("Unable to use cache: %v\n", err)
	}

	p.ecsClient = ecs.New(sess)
}

func (p *AwsEcsProvider) GetBackendUrls() []string {
  // Caching the backends indefinitely upon first call.
	if len(p.backendsCached) > 0 {
		return p.backendsCached
	}

	tasksList, err := p.ecsClient.ListTasks(&ecs.ListTasksInput{
		ServiceName: aws.String(p.ServiceName),
		Cluster:     aws.String(p.Cluster),
	})

	if err != nil {
	  log.Fatalf("Could not ListTasks: %v\n", err)
  }

	tasks, err := p.ecsClient.DescribeTasks(&ecs.DescribeTasksInput{
		Tasks:   tasksList.TaskArns,
		Cluster: aws.String(p.Cluster),
	})

  if err != nil {
    log.Fatalf("Could not DescribeTasks: %v\n", err)
  }

	for _, task := range tasks.Tasks {
		for _, container := range task.Containers {
			for _, networkInterface := range container.NetworkInterfaces {

			  // Build URI from the IP address of the ECS container.
				p.backendsCached = append(
				  p.backendsCached,
				  fmt.Sprintf(
				    "%s://%s:%s",
				    p.backendScheme,
				    *networkInterface.PrivateIpv4Address,
				    p.backendPort,
          ),
        )

			}
		}
	}
	return p.backendsCached
}
