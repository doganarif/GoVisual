#!/bin/bash

# Colors for terminal output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

SERVER_URL="http://localhost:8080"
DASHBOARD_URL="$SERVER_URL/__viz"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   GoVisual Dashboard Test Script      ${NC}"
echo -e "${BLUE}========================================${NC}"

# Check if server is running
echo -e "\n${YELLOW}Checking if server is running...${NC}"
if curl -s "$SERVER_URL/health" > /dev/null; then
    echo -e "${GREEN}Server is running!${NC}"
else
    echo -e "${RED}Server is not running. Please start the server with 'go run cmd/example/main.go'${NC}"
    echo -e "${YELLOW}Opening a new terminal window to run the server...${NC}"
    open -a Terminal .
    echo -e "${YELLOW}Please run 'go run cmd/example/main.go' in the new terminal window${NC}"
    echo -e "${YELLOW}Then press Enter to continue...${NC}"
    read -p ""
fi

# Open dashboard in browser
echo -e "\n${YELLOW}Opening dashboard in browser...${NC}"
open "$DASHBOARD_URL"
echo -e "${GREEN}Dashboard opened!${NC}"

# Sleep to let dashboard load
echo -e "\n${YELLOW}Waiting for dashboard to load...${NC}"
sleep 2

# Function to send request with timing
send_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local message=$4

    echo -e "\n${YELLOW}$message${NC}"
    
    if [ -n "$data" ]; then
        curl -s -X "$method" "$SERVER_URL$endpoint" -H "Content-Type: application/json" -d "$data" > /dev/null
    else
        curl -s -X "$method" "$SERVER_URL$endpoint" > /dev/null
    fi
    
    echo -e "${GREEN}Done!${NC}"
    sleep 1
}

# Begin sending requests
echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}   Starting Basic API Tests             ${NC}"
echo -e "${BLUE}========================================${NC}"

# GET requests
send_request "GET" "/api/users" "" "GET /api/users - Fetching all users"
send_request "GET" "/api/users/1" "" "GET /api/users/1 - Fetching user with ID 1"
send_request "GET" "/api/users/2" "" "GET /api/users/2 - Fetching user with ID 2" 
send_request "GET" "/api/users/999" "" "GET /api/users/999 - Fetching non-existent user"
send_request "GET" "/api/users/invalid" "" "GET /api/users/invalid - Fetching user with invalid ID"

# POST request
send_request "POST" "/api/users" '{"name":"New User","email":"new@example.com"}' "POST /api/users - Creating a new user"
send_request "POST" "/api/users" '{"invalid json' "POST /api/users - Sending invalid JSON"

# PUT request
send_request "PUT" "/api/users/1" '{"name":"Updated User","email":"updated@example.com"}' "PUT /api/users/1 - Updating user 1"

# DELETE request
send_request "DELETE" "/api/users/3" "" "DELETE /api/users/3 - Deleting user 3"

# Invalid method
send_request "PATCH" "/api/users/1" '{"name":"Patch Test"}' "PATCH /api/users/1 - Using unsupported HTTP method"

# Performance testing
send_request "GET" "/api/slow" "" "GET /api/slow - Testing slow endpoint (500ms)"
send_request "GET" "/api/very-slow" "" "GET /api/very-slow - Testing very slow endpoint (2000ms)"

# Error endpoints
send_request "GET" "/api/error" "" "GET /api/error - Testing 500 server error"
send_request "GET" "/api/not-found" "" "GET /api/not-found - Testing 404 not found error"
send_request "GET" "/api/unauthorized" "" "GET /api/unauthorized - Testing 401 unauthorized error"

# Redirect endpoint
send_request "GET" "/api/redirect" "" "GET /api/redirect - Testing redirect"

# Large response
send_request "GET" "/api/large-response" "" "GET /api/large-response - Testing large JSON response"

# Non-existent endpoint
send_request "GET" "/api/does-not-exist" "" "GET /api/does-not-exist - Testing non-existent endpoint"

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}   Testing Middleware & Tracing         ${NC}"
echo -e "${BLUE}========================================${NC}"

# Middleware testing
send_request "GET" "/api/middleware/simple" "" "GET /api/middleware/simple - Testing simple middleware"
send_request "GET" "/api/middleware/chain" "" "GET /api/middleware/chain - Testing middleware chain"
send_request "GET" "/api/middleware/slow" "" "GET /api/middleware/slow - Testing slow middleware"
send_request "GET" "/api/middleware/error" "" "GET /api/middleware/error - Testing middleware error"

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}   Testing Detailed Timing              ${NC}"
echo -e "${BLUE}========================================${NC}"

# Timing testing
send_request "GET" "/api/timing/basic" "" "GET /api/timing/basic - Testing basic timing"
send_request "GET" "/api/timing/detailed" "" "GET /api/timing/detailed - Testing detailed timing"
send_request "GET" "/api/timing/network-simulation" "" "GET /api/timing/network-simulation - Testing network simulation"

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}   Testing Route Matching               ${NC}"
echo -e "${BLUE}========================================${NC}"

# Route matching testing
send_request "GET" "/api/routes/users/42" "" "GET /api/routes/users/42 - Testing dynamic route with ID"
send_request "GET" "/api/routes/users/abc" "" "GET /api/routes/users/abc - Testing dynamic route with invalid ID"
send_request "GET" "/api/routes/products/electronics/123" "" "GET /api/routes/products/electronics/123 - Testing nested route"
send_request "GET" "/api/routes/products/books/abc" "" "GET /api/routes/products/books/abc - Testing nested route with invalid ID"
send_request "GET" "/api/routes/products/single" "" "GET /api/routes/products/single - Testing invalid nested route format"
send_request "GET" "/api/routes/regex/items/456" "" "GET /api/routes/regex/items/456 - Testing regex route matching"
send_request "GET" "/api/routes/regex/items/abc" "" "GET /api/routes/regex/items/abc - Testing regex route with non-matching pattern"

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}   Testing Multiple Request Patterns    ${NC}"
echo -e "${BLUE}========================================${NC}"

# Send a burst of requests
echo -e "\n${YELLOW}Sending burst of requests to /api/users...${NC}"
for i in {1..10}; do
    curl -s "$SERVER_URL/api/users" > /dev/null &
done
wait
echo -e "${GREEN}Burst completed!${NC}"

# Mixed middleware requests in parallel
echo -e "\n${YELLOW}Sending middleware requests in parallel...${NC}"
curl -s "$SERVER_URL/api/middleware/simple" > /dev/null &
curl -s "$SERVER_URL/api/middleware/chain" > /dev/null &
curl -s "$SERVER_URL/api/middleware/slow" > /dev/null &
wait
echo -e "${GREEN}Middleware requests completed!${NC}"

# Mix of fast and slow requests
echo -e "\n${YELLOW}Sending mix of fast and slow requests...${NC}"
curl -s "$SERVER_URL/api/users" > /dev/null &
curl -s "$SERVER_URL/api/timing/basic" > /dev/null &
curl -s "$SERVER_URL/api/slow" > /dev/null &
curl -s "$SERVER_URL/api/timing/network-simulation" > /dev/null &
curl -s "$SERVER_URL/api/very-slow" > /dev/null &
wait
echo -e "${GREEN}Mixed requests completed!${NC}"

# Testing all timing endpoints at once
echo -e "\n${YELLOW}Testing all timing endpoints...${NC}"
curl -s "$SERVER_URL/api/timing/basic" > /dev/null &
curl -s "$SERVER_URL/api/timing/detailed" > /dev/null &
curl -s "$SERVER_URL/api/timing/network-simulation" > /dev/null &
wait
echo -e "${GREEN}Timing requests completed!${NC}"

echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}   All Test Requests Completed         ${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "\n${GREEN}The dashboard should now show all the test requests.${NC}"
echo -e "${GREEN}Check the detailed timing and middleware information in the dashboard.${NC}"
echo -e "${GREEN}Try filtering by different request types and response times.${NC}" 