---
# MessageBroker - Architecture and Design Document

## High-Level Architecture

### Components

```
┌─────────────────────────────────────────────────────────────┐
│                      Clients Layer                          │
├─────────────────────┬───────────────────┬──────────────────┤
│  Producer Client    │  Consumer Client  │  Admin Client    │
└──────────┬──────────┴──────────┬────────┴────────┬─────────┘
           │                     │                 │
           │      gRPC API       │                 │
           ▼                     ▼                 ▼
┌─────────────────────────────────────────────────────────────┐
│                    Broker Cluster                           │
├────────────┬────────────┬─────────────┬──────────────────────┤
│  Broker 1  │  Broker 2  │  Broker 3   │  ...                │
│ (Leader)   │ (Replica)  │ (Replica)   │                     │
└────────────┴────────────┴─────────────┴──────────────────────┘
           │                     │                 │
           │     Replication     │                 │
           │     & Sync          │                 │
           └─────────┬───────────┘                 │
                     │                             │
                     ▼                             │
        ┌──────────────────────────┐               │
        │  In-Sync Replicas (ISR)  │◄──────────────┘
        │  Tracking & Management   │
        └──────────────────────────┘
           │
           │      Coordinator API
           ▼
┌─────────────────────────────────────────────────────────────┐
│                     etcd (Coordinator)                      │
│  (Leader Election, Metadata, Config Management)            │
└─────────────────────────────────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────────────────┐
│                  Storage Layer                              │
│  (Persistent Logs, Indices, Segments)                      │
└─────────────────────────────────────────────────────────────┘
```

## Key Design Decisions

### 1. **Topic and Partition Model**

- **Topics**: Logical message categories
- **Partitions**: Physical message segments within a topic
- Each partition can have millions of messages
- Partition-based parallelism for throughput

### 2. **Replication Strategy**

- **Leader-Follower Model**: 
  - One leader per partition handles writes
  - Followers replicate data from leader
  - Automatic failover on leader failure

- **In-Sync Replicas (ISR)**:
  - Tracks replicas caught up with leader
  - Configurable minimum ISR for durability
  - Dynamic ISR updates on replica failure

### 3. **High Availability**

- **Multi-broker deployment**: Distribute partitions across brokers
- **Replication factor**: Each partition replicated on multiple brokers
- **Automatic failover**: etcd-based leader election
- **Health monitoring**: Regular heartbeat and status checks

### 4. **Fault Tolerance**

- **Persistent storage**: Messages stored on disk
- **Write-ahead logs**: Ensure durability
- **Idempotent producers**: Prevent duplicate messages
- **Consumer offset management**: Track consumption progress

### 5. **Producer Model**

- **Partitioning strategies**:
  - Round-robin (for null keys)
  - Hash-based (for keyed messages)
  - Custom partitioner support

- **Acknowledgment levels**:
  - `none`: Fire and forget
  - `leader`: Leader ACK
  - `all`: All ISR ACK

### 6. **Consumer Model**

- **Consumer groups**: Multiple consumers sharing message load
- **Offset management**: Track reading position per partition
- **Auto-commit**: Periodic offset commits
- **Rebalancing**: Dynamic partition assignment

## Data Persistence

### Storage Layout

```
/app/data/
├── topic-1/
│   ├── 0/              # Partition 0
│   │   ├── log.data    # Message log
│   │   ├── index.idx   # Offset index
│   │   └── metadata    # Partition metadata
│   ├── 1/              # Partition 1
│   └── ...
├── topic-2/
│   └── ...
└── __internal__/       # System topics
    └── offsets/        # Consumer offsets
```

### Segment Management

- Segments rotate based on:
  - Time threshold (segmentMs)
  - Size threshold (segmentBytes)
- Old segments archived/deleted based on retention policy

## Replication Flow

### Write Path

```
1. Producer sends message to leader
2. Leader:
   - Appends to local log
   - Creates index entry
   - Returns offset
3. Leader replicates to followers
4. Followers:
   - Append to local log
   - Send ACK to leader
5. Leader updates ISR
6. Producer receives confirmation
```

### Read Path

```
1. Consumer requests messages from offset
2. Broker:
   - Looks up offset in index
   - Retrieves messages from log
   - Returns messages to consumer
3. Consumer processes messages
4. Consumer commits offset
5. Offset stored in cluster
```

## Coordination with etcd

### Metadata Stored in etcd

- Broker registration and health
- Leader for each partition
- ISR list per partition
- Topic configuration
- Consumer group state

### Leader Election Process

```
1. Broker publishes health to etcd
2. On leader failure, etcd detects timeout
3. etcd triggers election among ISR
4. New leader announced to all brokers
5. Brokers update routing information
```

## Scalability Considerations

### Horizontal Scaling

- Add brokers to cluster
- Rebalance partitions across new brokers
- Distribute load proportionally

### Vertical Scaling

- Increase broker resources (CPU, Memory, Storage)
- Monitor and tune JVM settings (if using JVM)
- Optimize storage with SSD

### Network Optimization

- Use batch processing for replication
- Compression support (snappy, gzip)
- Connection pooling between brokers

## Performance Characteristics

### Throughput

- Single broker: 100K+ messages/sec
- Multi-broker cluster: Linear with broker count
- Network bandwidth: Primary limiting factor

### Latency

- Producer latency: ~10-50ms (with replication to all ISR)
- Consumer latency: Milliseconds (disk I/O dependent)
- Replication lag: <100ms typical

### Durability

- Replication factor 3: Survives 1 broker failure
- Min ISR 2: Ensures durability with 1 broker down
- Persistence: Data survives process restarts

## Security Considerations

### Current Implementation

- gRPC for inter-broker communication
- Service account for pod authentication (K8s)
- Network policies (can be enabled)

### Future Enhancements

- TLS/mTLS for encrypted communication
- RBAC for producer/consumer access control
- Audit logging for compliance
- Data encryption at rest

## Monitoring and Observability

### Metrics to Monitor

- Message throughput (msgs/sec)
- Producer/Consumer latency (p50, p99)
- Replication lag (offset lag)
- Broker health (heartbeat/response time)
- Storage usage (disk space, segment count)

### Health Indicators

- Leader election time (<5 seconds)
- ISR stability (minimal changes)
- Partition distribution (balanced load)
- Consumer lag (within SLA)

## Future Enhancements

1. **Stream Processing**: Local windowing and aggregation
2. **Schema Registry**: Schema validation and evolution
3. **Transactions**: Cross-partition atomic writes
4. **KStream API**: Higher-level stream processing
5. **Connect Framework**: Data integration connectors
6. **Multi-tenancy**: Namespace isolation
7. **Tiered Storage**: Hot/cold data separation
