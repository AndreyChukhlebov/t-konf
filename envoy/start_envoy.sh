#!/bin/bash
set -e

TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S")_${SERVICE}
LOG_FILE="/tmp/envoy_access_${TIMESTAMP}.json.log"

echo "Starting Envoy with log file: $LOG_FILE"


cp /etc/envoy/envoy.yaml tmp/envoy.yaml
# Используем | как разделитель вместо /
sed -i "s|path: \"/tmp/envoy_access.json.log\"|path: \"${LOG_FILE}\"|" tmp/envoy.yaml

/usr/local/bin/envoy -c tmp/envoy.yaml