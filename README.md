# tool-gateway-kgateway-operator

Operator that deploys ToolGateway instances based on agentgateway/kgateway

## Description

This is a Kubernetes operator built with [operator-sdk](https://sdk.operatorframework.io/) that manages ToolGateway instances.

## Getting Started

### Prerequisites

- Go 1.24.0 or later
- Kubernetes cluster (or kind/minikube for local development)
- kubectl configured to interact with your cluster

### Development

Build the operator:
```bash
make build
```

Run tests:
```bash
make test
```

Run linter:
```bash
make lint
```

### Installation

Install CRDs into the cluster:
```bash
make install
```

Deploy the operator:
```bash
make deploy
```

### Running Locally

Run the operator locally against your configured Kubernetes cluster:
```bash
make run
```

## Project Structure

This project is scaffolded using operator-sdk and follows the standard Kubebuilder project layout:

- `cmd/` - Main entry point for the operator
- `config/` - Kubernetes manifests for deployment
- `test/` - Test files and utilities
- `Makefile` - Build, test, and deployment automation

## License

Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
