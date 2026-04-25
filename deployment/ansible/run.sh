#!/bin/bash
# Ansible playbook runner for easy deployment

set -e

PLAYBOOK=${1:-"deployment/ansible/deploy.yml"}
INVENTORY=${2:-"deployment/ansible/inventory.ini"}
TAGS=${3:-""}
EXTRA_VARS=${4:-""}

echo "=== MessageBroker Ansible Deployment ==="
echo "Playbook: $PLAYBOOK"
echo "Inventory: $INVENTORY"
echo ""

# Check if inventory file exists
if [ ! -f "$INVENTORY" ]; then
    echo "Error: Inventory file not found: $INVENTORY"
    echo "Please update the inventory file with your broker hosts"
    exit 1
fi

# Run playbook
if [ -z "$TAGS" ]; then
    ansible-playbook -i "$INVENTORY" "$PLAYBOOK" $EXTRA_VARS
else
    ansible-playbook -i "$INVENTORY" "$PLAYBOOK" --tags "$TAGS" $EXTRA_VARS
fi

echo ""
echo "Deployment completed!"
