#!/bin/bash
set -e

DB_NAME="meueu"
DB_USER="user"
DB_PASS="password"
DB_PORT="5432"
CONTAINER_NAME="meueu-db"

echo "Checking for existing container..."
if [ "$(docker ps -q -f name=$CONTAINER_NAME)" ]; then
    echo "Container $CONTAINER_NAME is already running."
    exit 0
fi

if [ "$(docker ps -aq -f name=$CONTAINER_NAME)" ]; then
    echo "Starting existing container..."
    docker start $CONTAINER_NAME
else
    echo "Creating and starting new container..."
    docker run -d \
        --name $CONTAINER_NAME \
        -e POSTGRES_USER=$DB_USER \
        -e POSTGRES_PASSWORD=$DB_PASS \
        -e POSTGRES_DB=$DB_NAME \
        -p $DB_PORT:5432 \
        postgres:15-alpine
fi

echo "Waiting for database to be ready..."
# Simple health check loop
until docker exec $CONTAINER_NAME pg_isready -U $DB_USER > /dev/null 2>&1; do
    echo -n "."
    sleep 1
done
echo "Database is ready!"
