#!/bin/bash
# ============================================================
# 🧪 Crom Protocol — Comprehensive Test Script
# Tests ALL endpoints and frontend pages
# Requires: server running on localhost:8080
# ============================================================

BASE="http://localhost:8080"
ADMIN_TOKEN="123456"
PASS=0
FAIL=0
TOTAL=0

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

check() {
    local name="$1"
    local expected="$2"
    local actual="$3"
    TOTAL=$((TOTAL+1))
    if echo "$actual" | grep -qF "$expected"; then
        echo -e "  ${GREEN}✅ PASS${NC} — $name"
        PASS=$((PASS+1))
    else
        echo -e "  ${RED}❌ FAIL${NC} — $name (expected: '$expected')"
        echo -e "       got: $(echo "$actual" | head -c 200)"
        FAIL=$((FAIL+1))
    fi
}

check_http() {
    local name="$1"
    local expected_code="$2"
    local actual_code="$3"
    TOTAL=$((TOTAL+1))
    if [ "$actual_code" = "$expected_code" ]; then
        echo -e "  ${GREEN}✅ PASS${NC} — $name (HTTP $actual_code)"
        PASS=$((PASS+1))
    else
        echo -e "  ${RED}❌ FAIL${NC} — $name (expected HTTP $expected_code, got HTTP $actual_code)"
        FAIL=$((FAIL+1))
    fi
}

echo ""
echo -e "${CYAN}╔════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║   🧪 CROM PROTOCOL — COMPREHENSIVE TESTS      ║${NC}"
echo -e "${CYAN}╚════════════════════════════════════════════════╝${NC}"
echo ""

# ── 1. Server Health ──────────────────────────────────────
echo -e "${YELLOW}▸ 1. SERVER HEALTH${NC}"
meta=$(curl -sf "$BASE/meta" 2>&1)
check "GET /meta returns JSON" "Crom/Meueu Node" "$meta"
check "/meta has network_id" "meueu-mainnet-v1" "$meta"
check "/meta has version" "2.0.0" "$meta"

# ── 2. Frontend Pages ────────────────────────────────────
echo ""
echo -e "${YELLOW}▸ 2. FRONTEND PAGES${NC}"

for page in "" "publish.html" "profile.html" "thread.html" "explorer.html" "chat.html" "inbox.html" "admin.html"; do
    code=$(curl -sL -o /dev/null -w "%{http_code}" "$BASE/$page")
    check_http "GET /$page" "200" "$code"
done

for js in "js/sdk/auth.js" "js/sdk/client.js" "js/app.js" "js/ui_login.js"; do
    code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/$js")
    check_http "GET /$js" "200" "$code"
done

publish_content=$(curl -sL "$BASE/publish.html")
check "publish.html has Tailwind" "tailwindcss" "$publish_content"
check "publish.html has kind selector" "kind-grid" "$publish_content"

# ── 3. Query API ──────────────────────────────────────────
echo ""
echo -e "${YELLOW}▸ 3. QUERY API${NC}"
query_code=$(curl -s -o /tmp/crom_query.json -w "%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d '{"filters":{},"limit":10}' "$BASE/v1/query")
check_http "POST /v1/query → 200" "200" "$query_code"
query_body=$(cat /tmp/crom_query.json 2>/dev/null)
check "Query returns JSON array" "[]" "$query_body"

# ── 4. Publish Validation ────────────────────────────────
echo ""
echo -e "${YELLOW}▸ 4. PUBLISH VALIDATION${NC}"

# No author header
code=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d '{"kind":"text"}' "$BASE/v1/publish")
check_http "No X-MeuEu-Author → 400" "400" "$code"

# Invalid author format (not 64 hex)
code=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -H "X-MeuEu-Author: tooshort" \
    -d '{"kind":"text"}' "$BASE/v1/publish")
check_http "Short author header → 400" "400" "$code"

# Valid header but missing fields
FAKE_AUTHOR="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
code=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -H "X-MeuEu-Author: $FAKE_AUTHOR" \
    -d "{\"author_pubkey\":\"$FAKE_AUTHOR\",\"kind\":\"text\",\"payload\":{\"text\":\"hello\"},\"signature\":\"fake\"}" \
    "$BASE/v1/publish")
check_http "Missing network_id/nonce → 400" "400" "$code"

# All fields but invalid sig
code=$(curl -s -o /tmp/crom_pub.txt -w "%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -H "X-MeuEu-Author: $FAKE_AUTHOR" \
    -d "{\"author_pubkey\":\"$FAKE_AUTHOR\",\"kind\":\"text\",\"payload\":{\"text\":\"hello\"},\"signature\":\"deadbeef\",\"network_id\":\"meueu-mainnet-v1\",\"nonce\":\"test-nonce-123\",\"claimed_at\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}" \
    "$BASE/v1/publish")
check_http "Invalid signature → 400" "400" "$code"
pub_body=$(cat /tmp/crom_pub.txt 2>/dev/null)
check "Error mentions signature" "ignature" "$pub_body"

# ── 5. Peer Discovery ────────────────────────────────────
echo ""
echo -e "${YELLOW}▸ 5. PEERS & SYNC${NC}"
code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/v1/peers")
check_http "GET /v1/peers → 200" "200" "$code"
code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/v1/sync")
check_http "GET /v1/sync → 200" "200" "$code"

# ── 6. Admin API ─────────────────────────────────────────
echo ""
echo -e "${YELLOW}▸ 6. ADMIN API${NC}"

# No auth
code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/admin/stats")
check_http "Admin no token → 401" "401" "$code"

# Wrong auth
code=$(curl -s -o /dev/null -w "%{http_code}" -H "X-Admin-Token: wrongpassword" "$BASE/admin/stats")
check_http "Admin wrong token → 401" "401" "$code"

# Correct auth — stats
code=$(curl -s -o /tmp/crom_stats.json -w "%{http_code}" -H "X-Admin-Token: $ADMIN_TOKEN" "$BASE/admin/stats")
check_http "Admin stats → 200" "200" "$code"
stats=$(cat /tmp/crom_stats.json 2>/dev/null)
check "Stats has total_nodes" "total_nodes" "$stats"
check "Stats has total_users" "total_users" "$stats"

# Correct auth — other endpoints
for endpoint in "whitelist" "banned_words" "banned_users" "banned_hashes"; do
    code=$(curl -s -o /dev/null -w "%{http_code}" -H "X-Admin-Token: $ADMIN_TOKEN" "$BASE/admin/$endpoint")
    check_http "Admin GET /$endpoint → 200" "200" "$code"
done

# v1 path
code=$(curl -s -o /dev/null -w "%{http_code}" -H "X-Admin-Token: $ADMIN_TOKEN" "$BASE/v1/admin/stats")
check_http "V1 admin stats → 200" "200" "$code"

# ── 7. CORS ──────────────────────────────────────────────
echo ""
echo -e "${YELLOW}▸ 7. CORS HEADERS${NC}"
cors=$(curl -sI -X OPTIONS "$BASE/v1/query" 2>&1)
check "CORS Allow-Origin present" "Access-Control-Allow-Origin" "$cors"
check "CORS allows X-MeuEu-Author" "X-MeuEu-Author" "$cors"
check "CORS allows X-Admin-Token" "X-Admin-Token" "$cors"
code=$(curl -s -o /dev/null -w "%{http_code}" -X OPTIONS "$BASE/v1/query")
check_http "OPTIONS preflight → 200" "200" "$code"

# ── 8. Edge Cases ─────────────────────────────────────────
echo ""
echo -e "${YELLOW}▸ 8. EDGE CASES${NC}"
code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/v1/nonexistent")
check_http "Unknown route → 404" "404" "$code"
code=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE/v1/query")
check_http "DELETE /v1/query → 405" "405" "$code"

# ── SUMMARY ──────────────────────────────────────────────
echo ""
echo -e "${CYAN}╔════════════════════════════════════════════════════╗${NC}"
printf "${CYAN}║${NC}  Total: %-3d  ${GREEN}Passed: %-3d${NC}  ${RED}Failed: %-3d${NC}          ${CYAN}║${NC}\n" $TOTAL $PASS $FAIL
echo -e "${CYAN}╚════════════════════════════════════════════════════╝${NC}"

if [ $FAIL -gt 0 ]; then
    echo -e "\n${RED}⚠️  $FAIL test(s) failed${NC}"
    exit 1
else
    echo -e "\n${GREEN}🎉 All tests passed!${NC}"
    exit 0
fi
