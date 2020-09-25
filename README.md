# Environment variables

| Variable | Default value | Description |
|---|---|---|
| `FWD_PORT`                  | `6081` | Listen on this port |
| `FWD_PROVIDER`              | `ecs`  | The provider of backend URLs. Can be `ecs`, `static` or `file` |
| `FWD_REFRESH_INTERVAL`      | `5m`   | The interval on which to refresh the backends from the provider. Set to "0" to disable refreshing. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". [See more here](https://golang.org/pkg/time/#ParseDuration) |
| `VARNISH_PORT`              | `6081` | *(for `ecs`)* Port to use for backends |
| `VARNISH_SCHEME`            | `http` | *(for `ecs`)* Protocol to use for backends |
| `CLUSTER_NAME`              |        | *(for `ecs`)* Name of the ECS cluster |
| `VARNISH_SERVICE_NAME`      |        | *(for `ecs`)* Name of the ECS service which has the Varnish tasks |
| `FWD_BACKENDS`              |        | *(for `static`)* List of backend URLs separated by comma. E.g. `http://localhost:3000,http://localhost:3001` |
| `FWD_BACKENDS_FILE`         |        | *(for `file`)* Path to the file that contains list of backend URLs separated by comma. E.g. `http://localhost:3000,http://localhost:3001` |
| `AWS_PROFILE`               |        | *(optional)* the profile to use when making requests to AWS |
| `AWS_IAM_AUTHENTICATOR_CACHE_FILE` | | *(optional)* full path to the file where AWS session credentials will be cached |


Sample Environment config string for GoLand build configuration
```shell script
AWS_PROFILE=myprofile;AWS_IAM_AUTHENTICATOR_CACHE_FILE=.aws-credentials-cache;CLUSTER_NAME=staging;VARNISH_SERVICE_NAME=varnish;FWD_PROVIDER=static;FWD_BACKENDS=http://localhost:3000,http://localhost:3001;VARNISH_SCHEME=https;VARNISH_PORT=6081
```

You can launch a test HTTP echo server based on NodeJS as follows
```shell script
$ npx http-echo-server 3000
```
