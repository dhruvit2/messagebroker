# MessageBroker Deployment Guide

## Overview

MessageBroker is a Kafka-like distributed message queue system written in Go with support for:
- Multiple topics and partitions
- High Availability (HA) and replication
- Fault tolerance and automatic failover
- Producer and Consumer APIs
- Docker and Kubernetes deployment
- Ansible automation

## Prerequisites

- Go 1.21+
- Docker and Docker Compose
- Kubernetes cluster (for K8s deployment)
- Helm 3.x (for K8s deployment)
- Ansible 2.9+ (for Ansible deployment)
- etcd 3.5+ (for coordination)

## Quick Start - Local Development

### 1. Build Binaries

```bash
# Build all binaries
make build

# Or build individually
make build-broker
make build-producer
make build-consumer
```

### 2. Run with Docker Compose

```bash
# Start a 3-broker cluster with etcd
make run-docker

# Verify cluster is running
docker-compose -f deployment/docker/docker-compose.yml ps
```

### 3. Test Producer and Consumer

```bash
# In terminal 1: Start consumer
./bin/consumer --brokers localhost:9092 --topic test-topic --group consumer-group-1

# In terminal 2: Send messages with producer
./bin/producer --brokers localhost:9092 --topic test-topic --messages 10
```

### 4. Cleanup

```bash
make stop-docker
```

## Kubernetes Deployment (Helm)

### Prerequisites

- Running Kubernetes cluster
- kubectl configured
- Helm 3.x installed

### Deployment Steps

```bash
# 1. Build and push Docker image
make build-docker
make push-docker

# 2. Deploy to Kubernetes
make deploy-k8s

# 3. Check deployment status
make status-k8s

# 4. Port forward to test
kubectl port-forward -n messagebroker svc/messagebroker 9092:9092

# 5. Update deployment if needed
make update-k8s

# 6. Delete deployment
make delete-k8s
```

### Helm Configuration

Edit `deployment/helm/messagebroker/values.yaml` to customize:
- Number of replicas
- Resource limits
- Storage size
- Replication factor
- Retention policies

### Helm Deployment Details

```bash
# List deployed releases
helm list -n messagebroker

# Get deployment values
helm get values messagebroker -n messagebroker

# Upgrade deployment
helm upgrade messagebroker deployment/helm/messagebroker \
  -n messagebroker \
  --set replicaCount=5

# Delete deployment
helm uninstall messagebroker -n messagebroker
```

## Ansible Deployment

### Prerequisites

- Ansible 2.9+
- SSH access to target hosts
- Target hosts running Linux (Ubuntu/CentOS/RHEL)
- Docker installed on target hosts

### Preparation

1. Update inventory file:

```bash
# Edit deployment/ansible/inventory.ini
[messagebroker]
broker1 ansible_host=192.168.1.11 ansible_user=ubuntu
broker2 ansible_host=192.168.1.12 ansible_user=ubuntu
broker3 ansible_host=192.168.1.13 ansible_user=ubuntu
```

2. Configure SSH access:

```bash
# Test connectivity
ansible all -i deployment/ansible/inventory.ini -m ping
```

### Deployment

```bash
# Deploy entire cluster
make deploy-ansible

# Or run Ansible playbook directly
ansible-playbook -i deployment/ansible/inventory.ini deployment/ansible/deploy.yml

# Deploy specific role
ansible-playbook -i deployment/ansible/inventory.ini deployment/ansible/deploy.yml \
  --tags "install-dependencies"

# With extra variables
ansible-playbook -i deployment/ansible/inventory.ini deployment/ansible/deploy.yml \
  -e "replication_factor=2"
```

### Verify Deployment

```bash
# Check cluster health
make status-ansible

# SSH to a broker and check service
ssh ubuntu@192.168.1.11 "sudo systemctl status messagebroker"

# Check Docker container
ssh ubuntu@192.168.1.11 "docker ps | grep messagebroker"

# Check logs
ssh ubuntu@192.168.1.11 "docker logs messagebroker-broker-1"
```

## Architecture

### Broker Components

- **Broker**: Core message store with partition management
- **Replication Manager**: Handles leader election and ISR management
- **Storage Layer**: Persistent message storage with indexing
- **Coordinator**: Metadata management using etcd
- **Producer Client**: Sends messages with configurable partitioning
- **Consumer Client**: Receives messages with offset tracking

### Data Flow

1. **Producer** → Selects partition based on key → Sends to leader broker
2. **Leader Broker** → Stores message → Replicates to followers → Acknowledges
3. **Consumer** → Requests messages from offset → Receives from any replica → Commits offset

## Configuration

### Broker Configuration

```yaml
broker:
  id: 1
  host: "0.0.0.0"
  port: 9092
  replicationFactor: 3
  minISR: 2
  retentionMs: 604800000  # 7 days
  segmentBytes: 1073741824  # 1GB
  maxPartitions: 1000

coordinator:
  url: "localhost:2379"

storage:
  mountPath: "/app/data"
```

### Environment Variables

```bash
BROKER_ID=1
BROKER_HOST=localhost
BROKER_PORT=9092
COORDINATOR_URL=localhost:2379
DATA_DIR=/app/data
```

## Monitoring

### Health Checks

```bash
# Check broker health
curl http://localhost:9092/health

# Get broker metadata
curl http://localhost:9092/metadata

# Check replication status
curl http://localhost:9092/replicas
```

### Kubernetes Monitoring

```bash
# Get pod logs
kubectl logs -n messagebroker messagebroker-0

# Describe pod
kubectl describe pod -n messagebroker messagebroker-0

# Get events
kubectl get events -n messagebroker
```

### Docker Monitoring

```bash
# Check container status
docker ps -f name=messagebroker

# View logs
docker logs -f messagebroker-broker-1

# Inspect container
docker inspect messagebroker-broker-1
```

## Troubleshooting

### Broker not starting

```bash
# Check logs
docker logs messagebroker-broker-1

# Verify coordinator connectivity
curl http://localhost:2379/version

# Check port availability
lsof -i :9092
```

### Replication issues

```bash
# Check ISR status
curl http://localhost:9092/replicas | jq '.isr'

# Verify broker connectivity
curl http://broker2:9092/health

# Check etcd status
etcdctl endpoint health
```

### Consumer lag

```bash
# Monitor consumer offsets
curl http://localhost:9092/consumer/group/consumer-group-1

# Check broker offsets
curl http://localhost:9092/offsets/test-topic
```

## Performance Tuning

### Broker Tuning

- Increase `segmentBytes` for larger segments
- Adjust `retentionMs` based on storage capacity
- Configure `replicationFactor` based on availability requirements
- Set `minISR` to 2 for HA with replication factor 3

### Resource Allocation

```yaml
resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 1000m
    memory: 1Gi
```

### Storage Optimization

```bash
# Use fast storage for data directory
# Monitor disk usage
df -h /app/data

# Archive old segments
ls -la /app/data/test-topic/0/
```

## Advanced Topics

### Custom Partitioner

Implement `client.Partitioner` interface to use custom partition selection logic.

### Multi-Datacenter Setup

Deploy separate clusters in each datacenter and implement cross-datacenter replication.

### Security

- Enable TLS for gRPC communication
- Implement ACL for producer/consumer access
- Use network policies in Kubernetes

## API Reference

See [API Documentation](./pkg/client/README.md) for producer and consumer APIs.

## Contributing

Contributions are welcome! Please submit pull requests and issues to the repository.

## License

MIT License

## Support

For issues and questions, please open an issue on the GitHub repository.
