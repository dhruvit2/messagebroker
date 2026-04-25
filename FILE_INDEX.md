# MessageBroker Project - File Index

## Documentation

### [README.md](./README.md)
- Project overview
- Quick feature summary
- Basic build and run instructions
- Deployment overview

### [QUICKSTART.md](./QUICKSTART.md) ⭐ START HERE
- Quick reference for all operations
- Copy-paste ready commands
- Common troubleshooting
- API usage examples

### [DEPLOYMENT.md](./DEPLOYMENT.md)
- Comprehensive deployment guide
- Step-by-step instructions for each deployment method
- Configuration details
- Monitoring and troubleshooting

### [ARCHITECTURE.md](./ARCHITECTURE.md)
- System architecture diagram
- Design decisions and rationale
- Data flow and persistence model
- Replication strategy
- Performance characteristics

## Core Source Code

### Broker (`pkg/broker/`)
- **types.go**: Data structures (Message, Topic, Partition, BrokerConfig)
- **broker.go**: Main broker implementation with topic/partition management
- **errors.go**: Error definitions

### Replication (`pkg/replication/`)
- **replication.go**: Replication manager for HA and fault tolerance
  - Leader election
  - ISR tracking
  - Failure detection
  - Replication coordination

### Storage (`pkg/storage/`)
- **storage.go**: Persistent storage layer
  - Log-structured storage
  - Index management
  - Offset tracking
  - Retention policies

### Client APIs (`pkg/client/`)
- **producer.go**: Producer client for publishing messages
  - Partitioning strategies
  - Batching and compression
  - Retry logic
  - Metrics

- **consumer.go**: Consumer client for subscribing to topics
  - Consumer groups
  - Offset management
  - Auto-commit support
  - Rebalancing

### Protocol Buffers (`pkg/pb/`)
- **messagebroker.proto**: gRPC service definitions
  - CreateTopic, ProduceMessage, ConsumeMessages
  - GetTopicMetadata, BrokerMetadata
  - Message and metadata structures

## Entry Points

### Broker Server (`cmd/broker/main.go`)
- Starts the message broker server
- Initializes broker, replication manager, storage
- Sets up gRPC server
- Implements broker API

### Producer Example (`cmd/producer/main.go`)
- Example producer client
- Sends configurable number of messages
- Demonstrates synchronous sending
- Shows metrics collection

### Consumer Example (`cmd/consumer/main.go`)
- Example consumer client
- Polls messages from topics
- Demonstrates consumer groups
- Shows offset management

## Deployment Configuration

### Docker (`deployment/docker/`)
- **Dockerfile**: Multi-stage build for minimal image
  - Builder stage: Compiles Go binaries
  - Runtime stage: Alpine base with binaries
  - Health checks configured

- **docker-compose.yml**: 3-broker cluster with etcd
  - Pre-configured broker network
  - Persistent volumes
  - Environment variables
  - Easy local testing

### Kubernetes/Helm (`deployment/helm/messagebroker/`)
- **Chart.yaml**: Helm chart metadata
- **values.yaml**: Configuration defaults
  - Resource limits
  - Replication settings
  - Storage configuration
  - Pod anti-affinity rules

- **templates/:**
  - **namespace.yaml**: Kubernetes namespace
  - **serviceaccount.yaml**: Service account for pods
  - **clusterrole.yaml**: RBAC permissions
  - **clusterrolebinding.yaml**: Role binding
  - **service.yaml**: ClusterIP and headless services
  - **configmap.yaml**: Broker configuration
  - **statefulset.yaml**: Main broker deployment
  - **_helpers.tpl**: Helm templating helpers

### Ansible (`deployment/ansible/`)
- **deploy.yml**: Main playbook orchestrating deployment
  - Calls roles in sequence
  - Pre and post tasks
  - Health checks

- **Roles:**
  - **install-dependencies/**: System packages, Docker, Go
  - **configure-broker/**: User creation, directories, config files
  - **deploy-broker/**: Docker container deployment
  - **configure-etcd/**: etcd installation and setup
  - **health-check/**: Cluster verification

- **Templates (in roles/*/templates/):**
  - **broker-config.yaml.j2**: Broker configuration template
  - **messagebroker.service.j2**: Systemd service file
  - **start-broker.sh.j2**: Broker startup script

- **Handlers (in roles/*/handlers/):**
  - **main.yml**: Service restart, systemd reload handlers

- **Supporting Files:**
  - **inventory.ini**: Ansible inventory with broker hosts
  - **health-check.sh**: Cluster health verification script
  - **run.sh**: Easy playbook execution wrapper

## Build Automation

### [Makefile](./Makefile)
Provides convenient targets for all operations:
- `make build*`: Compile binaries
- `make build-docker`: Build Docker image
- `make deploy-k8s*`: Kubernetes operations
- `make deploy-ansible`: Ansible deployment
- `make run-docker`: Start local cluster
- `make test`, `make lint`: Quality checks

## Project Root Files

### [go.mod](./go.mod)
Go module definition with dependencies:
- google.golang.org/grpc
- google.golang.org/protobuf
- github.com/sirupsen/logrus
- etcd.io/etcd/client/v3

### [.gitignore](./.gitignore)
Excludes build artifacts, environment files, IDE configs

## Navigation Tips

### I want to...

**Understand the architecture**
→ Read [ARCHITECTURE.md](./ARCHITECTURE.md)

**Deploy locally for testing**
→ See [QUICKSTART.md](./QUICKSTART.md) > Docker Compose section

**Deploy to Kubernetes**
→ See [QUICKSTART.md](./QUICKSTART.md) > Kubernetes section
→ Customize [deployment/helm/messagebroker/values.yaml](./deployment/helm/messagebroker/values.yaml)

**Deploy with Ansible**
→ Edit [deployment/ansible/inventory.ini](./deployment/ansible/inventory.ini)
→ Run `make deploy-ansible`

**Build a custom application**
→ See [pkg/client/producer.go](./pkg/client/producer.go) and consumer.go for API examples
→ Reference [DEPLOYMENT.md](./DEPLOYMENT.md) API section

**Monitor/troubleshoot**
→ See [DEPLOYMENT.md](./DEPLOYMENT.md) > Monitoring section
→ Check [QUICKSTART.md](./QUICKSTART.md) > Common Issues

**Scale the cluster**
→ [DEPLOYMENT.md](./DEPLOYMENT.md) > Performance Tuning

## Development Workflow

1. **Make changes** → Edit source files in `pkg/` and `cmd/`
2. **Build** → `make build` or `make build-docker`
3. **Test locally** → `make run-docker` or `make test`
4. **Deploy to dev** → `make deploy-k8s` or `make deploy-ansible`
5. **Monitor** → Check logs and health endpoints
6. **Scale** → Adjust replicas or configuration

## Total Project Size

- **Go source code**: ~2000+ lines
- **Configuration files**: 20+ YAML/template files
- **Documentation**: 4 comprehensive markdown files
- **Build automation**: Makefile with 20+ targets
- **Deployment options**: 3 (Docker, Kubernetes, Ansible)

## Key Numbers

- **File count**: 30+
- **Packages**: 7 core packages + tests
- **gRPC methods**: 5 main operations
- **Deployment templates**: 8 Kubernetes + 6 Ansible roles
- **Configuration options**: 40+

---

**Ready to start?** → Run `make help` or see [QUICKSTART.md](./QUICKSTART.md)
