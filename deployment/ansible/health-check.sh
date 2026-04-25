#!/bin/bash
# Script to verify MessageBroker cluster health

set -e

BROKERS=${1:-"localhost:9092"}

echo "=== MessageBroker Cluster Health Check ==="
echo ""

# Convert comma-separated brokers to array
IFS=',' read -ra BROKER_ARRAY <<< "$BROKERS"

for broker in "${BROKER_ARRAY[@]}"; do
    echo "Checking broker: $broker"
    
    # Check if broker is responding
    if curl -s "http://$broker/health" > /dev/null; then
        echo "  ✓ Broker is healthy"
    else
        echo "  ✗ Broker is not responding"
        continue
    fi
    
    # Get broker metadata
    echo "  Getting broker metadata..."
    curl -s "http://$broker/metadata" | jq . 2>/dev/null || echo "  (Could not retrieve metadata)"
    
    echo ""
done

echo "=== Cluster Replication Status ==="
curl -s "http://$(echo $BROKERS | cut -d',' -f1)/replicas" | jq . 2>/dev/null || echo "Could not retrieve replication status"

echo ""
echo "=== Topic Information ==="
curl -s "http://$(echo $BROKERS | cut -d',' -f1)/topics" | jq . 2>/dev/null || echo "Could not retrieve topics"

echo ""
echo "Health check complete!"
