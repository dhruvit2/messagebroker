# Quick Reference Guide - MessageBroker

## Building and Running Locally

### Prerequisites
```bash
# Install Go 1.21+
# Docker and Docker Compose installed
```

### Option 1: Local Binary Execution
```bash
# Build binaries
make build

# Terminal 1: Start broker
./bin/broker --id 1 --host localhost --port 9092 --coordinator localhost:2379

# Terminal 2: Start producer
./bin/producer --brokers localhost:9092 --topic test --messages 100

# Terminal 3: Start consumer
./bin/consumer --brokers localhost:9092 --topics test --group group1
```

### Option 2: Docker Compose (Recommended for local testing)
```bash
# Start 3-broker cluster with etcd
make run-docker

# View logs
docker-compose -f deployment/docker/docker-compose.yml logs -f

# Stop cluster
make stop-docker
```

## Kubernetes Deployment (Production)

### Quick Deploy
```bash
# 1. Build and push image
make build-docker
make push-docker

# 2. Deploy to K8s
make deploy-k8s

# 3. Check status
make status-k8s

# 4. Port-forward for testing
kubectl port-forward -n messagebroker svc/messagebroker 9092:9092
```

### Configuration
Edit `deployment/helm/messagebroker/values.yaml`:
- `replicaCount`: Number of broker replicas (default: 3)
- `broker.replicationFactor`: Replication factor (default: 3)
- `broker.minISR`: Minimum in-sync replicas (default: 2)
- `persistence.size`: Storage size (default: 10Gi)
- `resources`: CPU/Memory limits

## Ansible Deployment

### Preparation
```bash
# Edit inventory
vi deployment/ansible/inventory.ini

# Update with your broker IPs:
# broker1 ansible_host=192.168.1.11
# broker2 ansible_host=192.168.1.12
# broker3 ansible_host=192.168.1.13

# Test connectivity
ansible all -i deployment/ansible/inventory.ini -m ping
```

### Deploy
```bash
# Deploy entire cluster
make deploy-ansible

# Check health
make status-ansible

# View logs
ssh ubuntu@<broker_ip> "docker logs messagebroker-broker-1"
```

## Producer API Usage

### Simple Producer
```go
config := &client.ProducerConfig{
    BrokerAddresses: []string{"localhost:9092"},
    Acks: "all",
}

producer, _ := client.NewProducer(config)
defer producer.Close()

// Send message synchronously
offset, err := producer.SendSync(&client.ProducerRecord{
    Topic: "my-topic",
    Key:   []byte("key1"),
    Value: []byte("message-value"),
})
```

## Consumer API Usage

### Simple Consumer
```go
config := &client.ConsumerConfig{
    BrokerAddresses: []string{"localhost:9092"},
    GroupID: "my-group",
    Topics: []string{"my-topic"},
}

consumer, _ := client.NewConsumer(config)
defer consumer.Close()

// Poll for messages
for {
    messages, _ := consumer.Poll(context.Background(), 1000)
    for _, msg := range messages {
        // Process message
        log.Printf("Message: %s", string(msg.Value))
    }
}
```

## Cluster Administration

### View Cluster Status
```bash
# Check Kubernetes pods
kubectl get pods -n messagebroker

# View broker logs
kubectl logs -n messagebroker messagebroker-0

# Get pod details
kubectl describe pod -n messagebroker messagebroker-0
```

### Health Checks
```bash
# Check broker health
curl http://localhost:9092/health

# Get broker metadata
curl http://localhost:9092/metadata

# Check replication status
curl http://localhost:9092/replicas
```

### Scaling
```bash
# Kubernetes: Update replicas
kubectl scale statefulset messagebroker -n messagebroker --replicas=5

# Or use Helm
helm upgrade messagebroker deployment/helm/messagebroker \
  -n messagebroker \
  --set replicaCount=5
```

## Performance Tuning

### For High Throughput
```yaml
# In values.yaml
resources:
  requests:
    cpu: 2000m
    memory: 2Gi
  limits:
    cpu: 4000m
    memory: 4Gi

broker:
  segmentBytes: 2147483648  # 2GB
  batchSize: 100
```

### For High Availability
```yaml
broker:
  replicationFactor: 3
  minISR: 2
  retentionMs: 1209600000  # 14 days
```

## Common Issues

### Broker Not Starting
```bash
# Check logs
docker logs messagebroker-broker-1

# Verify etcd is running
curl http://localhost:2379/version

# Check port availability
lsof -i :9092
```

### Consumer Lag
```bash
# Check consumer offsets
curl http://localhost:9092/consumer/group/my-group

# Monitor with kubectl
kubectl logs -f -n messagebroker messagebroker-0
```

### Replication Issues
```bash
# Check ISR status
curl http://localhost:9092/replicas | jq '.isr'

# Verify broker connectivity
curl http://broker2:9092/health

# Restart broker
kubectl delete pod -n messagebroker messagebroker-0
```

## Files and Structure

```
messagebroker/
├── cmd/                    # Entry points
│   ├── broker/            # Broker server
│   ├── producer/          # Producer CLI
│   └── consumer/          # Consumer CLI
├── pkg/                   # Core packages
│   ├── broker/           # Broker logic
│   ├── client/           # Producer/Consumer APIs
│   ├── partition/        # Partition management
│   ├── replication/      # Replication & HA
│   ├── storage/          # Persistence layer
│   ├── coordinator/      # etcd coordination
│   └── pb/               # Protocol buffers
├── deployment/           # Deployment configs
│   ├── docker/          # Dockerfile & Compose
│   ├── helm/            # Kubernetes Helm chart
│   └── ansible/         # Ansible playbooks
├── Makefile             # Build automation
├── README.md            # Project overview
├── ARCHITECTURE.md      # Design documentation
└── DEPLOYMENT.md        # Deployment guide
```

## Make Targets Summary

```bash
make build              # Build all binaries
make build-docker       # Build Docker image
make push-docker        # Push to registry
make run-docker         # Start Docker Compose cluster
make deploy-k8s         # Deploy to Kubernetes
make update-k8s         # Update K8s deployment
make delete-k8s         # Remove K8s deployment
make status-k8s         # Check K8s status
make deploy-ansible     # Deploy via Ansible
make status-ansible     # Check Ansible deployment
make test              # Run tests
make lint              # Run linters
make clean             # Clean build artifacts
```

## Default Ports

- Broker: **9092** (gRPC)
- etcd: **2379** (Client API)
- etcd peer: **2380** (Peer communication)

## Environment Variables

```bash
BROKER_ID=1
BROKER_HOST=0.0.0.0
BROKER_PORT=9092
COORDINATOR_URL=etcd:2379
DATA_DIR=/app/data
```

## Next Steps

1. Review [ARCHITECTURE.md](./ARCHITECTURE.md) for design details
2. Check [DEPLOYMENT.md](./DEPLOYMENT.md) for detailed deployment guide
3. Explore [cmd/broker](./cmd/broker) for server implementation
4. Review [pkg/client](./pkg/client) for API usage examples
5. Run local demo with `make run-docker`

---

For more information, see the complete documentation in the respective files.
