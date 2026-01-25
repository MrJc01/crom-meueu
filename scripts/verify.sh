#!/bin/bash
set -e

# Configuration
export DATABASE_URL="postgres://user:password@localhost:5432/meueu?sslmode=disable"
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo "=========================================="
echo "    Crom Social Protocol - Verification   "
echo "=========================================="

# 1. Start Database
echo -e "\n[1/5] 🐘 Starting Database..."
./scripts/start_db.sh

# 2. Build Tools
echo -e "\n[2/5] 🛠️  Building Migration Tool..."
go build -o bin/migrate cmd/migrate/main.go
echo -e "${GREEN}Migration tool built successfully.${NC}"

# 3. Run Migrations
echo -e "\n[3/5] 🔄 Running Database Migrations..."
./bin/migrate -direction=up
echo -e "${GREEN}Migrations applied.${NC}"

# 4. Run Tests
echo -e "\n[4/5] 🧪 Running Unit Tests..."
go test ./... -v
echo -e "${GREEN}Tests passed.${NC}"

# 5. Build API
echo -e "\n[5/5] 🏗️  Building API..."
go build -o bin/api cmd/api/main.go
echo -e "${GREEN}API binary built successfully.${NC}"

echo -e "\n=========================================="
echo -e "${GREEN}✅ ALL SYSTEMS GO! The project is healthy.${NC}"
echo "=========================================="
