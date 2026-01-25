#!/bin/bash

CONTAINER_NAME="meueu-db"

echo "Stopping database container..."
docker stop $CONTAINER_NAME
docker rm $CONTAINER_NAME
echo "Database container stopped and removed."
