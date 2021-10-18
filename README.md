# Varnish forwarder
Transparently forwards all requests to multiple Varnish cache servers. Has an option to get the servers' IP
addresses from [AWS ECS](https://aws.amazon.com/ecs/) given a cluster and service names.

# Build and run
```bash
$ go mod download
$ go build -o main .
$ FWD_PROVIDER=static FWD_BACKENDS=http://127.0.0.1:3000,http://127.0.0.1:3001 ./main
```

In two other terminals launch the two example backends by running:
```bash
$ npx http-echo-server 3000
```
and
```bash
$ npx http-echo-server 3001
```

Sending requests to `http://localhost:6081/test` will result in the following output.
```bash
$ curl -v -X BAN http://localhost:6081/test
*   Trying 127.0.0.1:6081...
* TCP_NODELAY set
* Connected to localhost (127.0.0.1) port 6081 (#0)
> BAN /test HTTP/1.1
> Host: localhost:6081
> User-Agent: curl/7.68.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Date: Mon, 18 Oct 2021 16:08:44 GMT
< Content-Length: 22
< Content-Type: text/plain; charset=utf-8
<
{"statuses":[200,200]}
```

# Docker
There's a `Dockerfile` which can be built and run as follows:
```bash
$ docker build -t varnish-fwd -f Dockerfile .
$ docker run --rm -it -p 6081:6081 -e FWD_PROVIDER=static -e FWD_BACKENDS=http://172.17.0.1:3000,http://172.17.0.1:3001 varnish-fwd
```

# Environment variables

| Variable                           | Default | Description |
| ---                                | ---     | ---         |
| `FWD_PORT`                         | `6081`  | Listen on this port |
| `FWD_PROVIDER`                     | `ecs`   | The provider of backend URLs. Can be `ecs`, `static` or `file` |
| `FWD_REFRESH_INTERVAL`             | `5m`    | The interval on which to refresh the backends from the provider. Set to "0" to disable refreshing. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". [See more here](https://golang.org/pkg/time/#ParseDuration) |
| `VARNISH_PORT`                     | `6081`  | *(for `ecs`)* Port to use for backends |
| `VARNISH_SCHEME`                   | `http`  | *(for `ecs`)* Protocol to use for backends |
| `CLUSTER_NAME`                     |         | *(for `ecs`)* Name of the ECS cluster |
| `VARNISH_SERVICE_NAME`             |         | *(for `ecs`)* Name of the ECS service which has the Varnish tasks |
| `FWD_BACKENDS`                     |         | *(for `static`)* List of backend URLs separated by comma. E.g. `http://localhost:3000,http://localhost:3001` |
| `FWD_BACKENDS_FILE`                |         | *(for `file`)* Path to the file that contains list of backend URLs separated by comma. E.g. `http://localhost:3000,http://localhost:3001` |
| `AWS_PROFILE`                      |         | *(optional)* the profile to use when making requests to AWS |
| `AWS_IAM_AUTHENTICATOR_CACHE_FILE` |         | *(optional)* full path to the file where AWS session credentials will be cached |


Sample Environment config string for GoLand build configuration
```shell script
AWS_PROFILE=myprofile;AWS_IAM_AUTHENTICATOR_CACHE_FILE=.aws-credentials-cache;CLUSTER_NAME=staging;VARNISH_SERVICE_NAME=varnish;FWD_PROVIDER=static;FWD_BACKENDS=http://localhost:3000,http://localhost:3001;VARNISH_SCHEME=https;VARNISH_PORT=6081
```

You can launch a test HTTP echo server based on NodeJS as follows
```shell script
$ npx http-echo-server 3000
```
