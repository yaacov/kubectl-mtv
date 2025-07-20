#!/bin/bash
#
# Test runner script for kubectl-mtv e2e tests
#
# This script provides a convenient way to run e2e tests with proper
# environment setup and validation.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Default values
VERBOSE=false
PARALLEL=false
REPORT=false
COVERAGE=false
MARKERS=""
TEST_PATTERN=""

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✅${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}❌${NC} $1"
}

# Function to show usage
show_usage() {
    cat << EOF
kubectl-mtv E2E Test Runner

Usage: $0 [OPTIONS] [TEST_PATTERN]

OPTIONS:
    -h, --help              Show this help message
    -v, --verbose           Enable verbose output
    -p, --parallel          Run tests in parallel
    -r, --report            Generate HTML and JSON reports
    -c, --coverage          Run with coverage reporting
    -m, --markers MARKERS   Run only tests with specific markers
    -s, --setup             Set up test environment only
    --version-only          Run only version tests
    --providers-only        Run only provider tests
    --no-credentials        Skip tests requiring credentials
    --check-only            Only check prerequisites

TEST_PATTERN:
    Optional pytest pattern to filter tests (e.g., "test_version.py::TestVersion::test_basic")

EXAMPLES:
    $0                      # Run all tests
    $0 -v                   # Run with verbose output
    $0 -p -r                # Run in parallel with reports
    $0 --version-only       # Run only version tests
    $0 -m "not requires_credentials"  # Skip credential tests
    $0 test_version.py      # Run specific test file

ENVIRONMENT:
    Set up credentials in .env file (copy from .env.template)
    Ensure you're logged into an OpenShift/K8s cluster with admin privileges
    Build kubectl-mtv binary with 'make' in project root

EOF
}

# Function to check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check if we're in the right directory
    if [[ ! -f "$SCRIPT_DIR/conftest.py" ]]; then
        print_error "Not in e2e test directory"
        exit 1
    fi
    
    # Check Python
    if ! command -v python3 &> /dev/null; then
        print_error "Python 3 is required but not installed"
        exit 1
    fi
    
    # Check cluster connectivity
    if ! kubectl cluster-info --request-timeout=10s &> /dev/null; then
        print_error "Not connected to Kubernetes/OpenShift cluster"
        print_info "Please run: oc login <cluster-url> or configure kubectl"
        exit 1
    fi
    
    # Check admin permissions
    if ! kubectl auth can-i '*' '*' --all-namespaces &> /dev/null; then
        print_error "No cluster admin permissions"
        print_info "Please ensure you're logged in as cluster admin"
        exit 1
    fi
    
    # Check for kubectl-mtv binary
    local binary_found=false
    if command -v kubectl-mtv &> /dev/null; then
        binary_found=true
    elif [[ -f "$PROJECT_ROOT/kubectl-mtv" ]]; then
        binary_found=true
    fi
    
    if [[ "$binary_found" == "false" ]]; then
        print_error "kubectl-mtv binary not found"
        print_info "Please run 'make' in the project root: $PROJECT_ROOT"
        exit 1
    fi
    
    print_success "Prerequisites check passed"
    print_info "Cluster: $(kubectl config current-context)"
    print_info "User: $(kubectl config view --minify -o jsonpath='{.contexts[0].context.user}' 2>/dev/null || echo 'unknown')"
}

# Function to setup test environment
setup_environment() {
    print_info "Setting up test environment..."
    
    cd "$SCRIPT_DIR"
    
    # Create virtual environment if it doesn't exist
    if [[ ! -d "venv" ]]; then
        print_info "Creating Python virtual environment..."
        python3 -m venv venv
    fi
    
    # Activate virtual environment
    source venv/bin/activate
    
    # Install/upgrade dependencies
    print_info "Installing test dependencies..."
    pip install --upgrade pip > /dev/null
    pip install -r requirements.txt > /dev/null
    
    print_success "Test environment ready"
}

# Function to build pytest command
build_pytest_command() {
    local cmd="venv/bin/pytest"
    
    # Add verbosity
    if [[ "$VERBOSE" == "true" ]]; then
        cmd="$cmd -v"
    fi
    
    # Add parallel execution
    if [[ "$PARALLEL" == "true" ]]; then
        cmd="$cmd -n auto"
    fi
    
    # Add markers
    if [[ -n "$MARKERS" ]]; then
        cmd="$cmd -m \"$MARKERS\""
    fi
    
    # Add reporting
    if [[ "$REPORT" == "true" ]]; then
        mkdir -p reports
        cmd="$cmd --html=reports/report.html --self-contained-html"
        cmd="$cmd --json-report --json-report-file=reports/report.json"
    fi
    
    # Add coverage
    if [[ "$COVERAGE" == "true" ]]; then
        mkdir -p reports
        pip install pytest-cov > /dev/null
        cmd="$cmd --cov=. --cov-report=html:reports/coverage --cov-report=term-missing"
    fi
    
    # Add test pattern
    if [[ -n "$TEST_PATTERN" ]]; then
        cmd="$cmd $TEST_PATTERN"
    fi
    
    echo "$cmd"
}

# Function to run tests
run_tests() {
    print_info "Running kubectl-mtv e2e tests..."
    
    cd "$SCRIPT_DIR"
    source venv/bin/activate
    
    local cmd
    cmd=$(build_pytest_command)
    
    print_info "Command: $cmd"
    
    # Run the tests
    if eval "$cmd"; then
        print_success "Tests completed successfully"
        
        # Show report locations
        if [[ "$REPORT" == "true" ]]; then
            print_info "HTML report: $SCRIPT_DIR/reports/report.html"
            print_info "JSON report: $SCRIPT_DIR/reports/report.json"
        fi
        
        if [[ "$COVERAGE" == "true" ]]; then
            print_info "Coverage report: $SCRIPT_DIR/reports/coverage/index.html"
        fi
        
        return 0
    else
        print_error "Tests failed"
        return 1
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -p|--parallel)
            PARALLEL=true
            shift
            ;;
        -r|--report)
            REPORT=true
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -m|--markers)
            MARKERS="$2"
            shift 2
            ;;
        -s|--setup)
            check_prerequisites
            setup_environment
            print_success "Setup complete"
            exit 0
            ;;
        --version-only)
            MARKERS="version"
            shift
            ;;
        --providers-only)
            MARKERS="provider"
            shift
            ;;
        --no-credentials)
            MARKERS="not requires_credentials"
            shift
            ;;
        --check-only)
            check_prerequisites
            print_success "Prerequisites check complete"
            exit 0
            ;;
        -*|--*)
            print_error "Unknown option $1"
            show_usage
            exit 1
            ;;
        *)
            TEST_PATTERN="$1"
            shift
            ;;
    esac
done

# Main execution
check_prerequisites
setup_environment
run_tests
