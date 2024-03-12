# Service Account Issuer Discovery
[![REUSE status](https://api.reuse.software/badge/github.com/gardener/service-account-issuer-discovery)](https://api.reuse.software/info/github.com/gardener/service-account-issuer-discovery)

A simple server that allows exposing the OpenID discovery documents of a Kubernetes cluster.

Work in progress... Partial documentation ahead.

### Quick start

To run the server with minimal configuration export the `KUBECONFIG` environment variable and run:
``` 
go run ./cmd/service-account-issuer-discovery/main.go --hostname=<issuer-of-cluster>
```
Or pass the `kubeconfig` as a flag:
``` 
go run ./cmd/service-account-issuer-discovery/main.go --kubeconfig=<path-to-my-kubeconfig> --hostname=<issuer-of-cluster>
```

Retrieve the `well-known` document by querying `/.well-known/openid-configuration`.
