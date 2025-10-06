#!/bin/bash

# MinLang Test Suite Runner
# Runs all tests with coverage and provides a nice summary

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}MinLang Comprehensive Test Suite${NC}"
echo -e "${BLUE}================================${NC}"
echo ""

# Function to run tests for a package
run_package_tests() {
    local package=$1
    local name=$2

    echo -e "${YELLOW}Testing: ${name}${NC}"
    if go test -v "./${package}" 2>&1 | grep -E "PASS|FAIL|RUN"; then
        echo -e "${GREEN}✓ ${name} tests passed${NC}"
        return 0
    else
        echo -e "${RED}✗ ${name} tests failed${NC}"
        return 1
    fi
    echo ""
}

# Track overall status
FAILED=0

# Run unit tests for each component
echo -e "${BLUE}Running Unit Tests...${NC}"
echo ""

run_package_tests "lexer" "Lexer" || FAILED=1
run_package_tests "parser" "Parser" || FAILED=1
run_package_tests "compiler" "Compiler" || FAILED=1
run_package_tests "vm" "VM" || FAILED=1

echo ""
echo -e "${BLUE}Running Integration Tests...${NC}"
echo ""

# Run integration tests
if go test -v -timeout 30s . 2>&1 | grep -E "PASS|FAIL|RUN|==="; then
    echo -e "${GREEN}✓ Integration tests passed${NC}"
else
    echo -e "${RED}✗ Integration tests failed${NC}"
    FAILED=1
fi

echo ""
echo -e "${BLUE}Running Coverage Analysis...${NC}"
echo ""

# Generate coverage report
go test -coverprofile=coverage.out ./... > /dev/null 2>&1
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')

echo -e "Total Coverage: ${GREEN}${COVERAGE}${NC}"
echo ""

# Optional: Show detailed coverage by package
echo -e "${BLUE}Coverage by Package:${NC}"
go tool cover -func=coverage.out | grep -v "total" | awk '{print $1, $3}' | column -t

echo ""
echo -e "${BLUE}================================${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed! ✓${NC}"
    echo -e "${BLUE}================================${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed! ✗${NC}"
    echo -e "${BLUE}================================${NC}"
    exit 1
fi
