#!/bin/bash

# Service Account Admin Access & Long-term Token Creation Helper Script
# This script creates a ServiceAccount with admin privileges and creates a long-term token
# for use in kubectl-mtv tests, particularly for creating OpenShift providers

set -e

# Default values
NAMESPACE="kubectl-mtv-test"
SERVICE_ACCOUNT="kubectl-mtv-admin"
PLATFORM=""
CLEANUP=false

# Function to display usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Creates a ServiceAccount with admin privileges and a long-term token for testing.

OPTIONS:
    -n, --namespace NAME        Namespace/project name (default: kubectl-mtv-test)
    -s, --service-account NAME  Service account name (default: kubectl-mtv-admin)
    -p, --platform PLATFORM     Platform: kubernetes or openshift (auto-detect if not specified)
    -c, --cleanup               Clean up existing resources instead of creating new ones
    -h, --help                  Show this help message

EXAMPLES:
    # Create with defaults (auto-detect platform)
    $0

    # Create with custom namespace
    $0 --namespace my-test-ns

    # Create for OpenShift specifically
    $0 --platform openshift --namespace my-project
    
    # Clean up existing resources
    $0 --cleanup
    
    # Clean up resources in specific namespace
    $0 --cleanup --namespace my-test-ns
EOF
}

# Function to detect platform
detect_platform() {
    if command -v oc >/dev/null 2>&1 && oc version --client >/dev/null 2>&1; then
        if oc api-resources | grep -q "routes.route.openshift.io" 2>/dev/null; then
            echo "openshift"
            return
        fi
    fi
    
    if command -v kubectl >/dev/null 2>&1; then
        echo "kubernetes"
        return
    fi
    
    echo "ERROR: Neither kubectl nor oc found in PATH" >&2
    exit 1
}

# Function to check if namespace exists
namespace_exists() {
    local ns="$1"
    if [[ "$PLATFORM" == "openshift" ]]; then
        oc get project "$ns" >/dev/null 2>&1
    else
        kubectl get namespace "$ns" >/dev/null 2>&1
    fi
}

# Function to create namespace/project
create_namespace() {
    local ns="$1"
    echo "Creating namespace/project: $ns"
    
    if [[ "$PLATFORM" == "openshift" ]]; then
        oc new-project "$ns" 2>/dev/null || oc project "$ns"
    else
        kubectl create namespace "$ns" 2>/dev/null || echo "Namespace $ns already exists"
    fi
}

# Function to create service account
create_service_account() {
    local ns="$1"
    local sa="$2"
    
    echo "Creating service account: $sa in namespace: $ns"
    
    if [[ "$PLATFORM" == "openshift" ]]; then
        oc create sa "$sa" -n "$ns" 2>/dev/null || echo "Service account $sa already exists"
    else
        kubectl create serviceaccount "$sa" -n "$ns" 2>/dev/null || echo "Service account $sa already exists"
    fi
}

# Function to grant admin privileges
grant_admin_privileges() {
    local ns="$1"
    local sa="$2"
    
    echo "Granting admin privileges to service account: $sa in namespace: $ns"
    
    if [[ "$PLATFORM" == "openshift" ]]; then
        oc adm policy add-role-to-user admin -z "$sa" -n "$ns"
    else
        kubectl create rolebinding "${sa}-admin-binding" \
            --clusterrole=admin \
            --serviceaccount="${ns}:${sa}" \
            --namespace="$ns" 2>/dev/null || echo "RoleBinding already exists"
    fi
}

# Function to create long-term token secret
create_long_term_token() {
    local ns="$1"
    local sa="$2"
    
    echo "Creating long-term token secret for service account: $sa"
    
    cat << EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: ${sa}-token
  annotations:
    kubernetes.io/service-account.name: $sa
  namespace: $ns
type: kubernetes.io/service-account-token
EOF

    # Wait a moment for the token to be populated
    echo "Waiting for token to be populated..."
    sleep 5
    
    echo "=== Generated Long-term Token ==="
    local token
    token=$(kubectl -n "$ns" get secret "${sa}-token" -o go-template='{{ .data.token | base64decode }}' 2>/dev/null || echo "")
    
    if [[ -n "$token" ]]; then
        echo "$token"
        echo
        echo "=== API Server URL ==="
        local api_url
        api_url=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
        echo "$api_url"
        return 0
    else
        echo "Token not yet available, trying again..."
        sleep 3
        token=$(kubectl -n "$ns" get secret "${sa}-token" -o go-template='{{ .data.token | base64decode }}' 2>/dev/null || echo "")
        if [[ -n "$token" ]]; then
            echo "$token"
            echo
            echo "=== API Server URL ==="
            local api_url
            api_url=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
            echo "$api_url"
            return 0
        else
            echo "ERROR: Failed to retrieve token. Try again in a few seconds with:"
            echo "kubectl -n $ns get secret ${sa}-token -o go-template='{{ .data.token | base64decode }}'"
            return 1
        fi
    fi
}

# Function to display verification commands
show_verification() {
    local ns="$1"
    local token="$2"
    
    local api_url
    api_url=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
    
    cat << EOF

=== Verification Commands ===

Test the token by setting it as an environment variable and running kubectl commands:

export OPENSHIFT_TARGET_TOKEN="$token"
export OPENSHIFT_TARGET_URL="$api_url"

# Test access within the namespace (should work)
kubectl --token="\$OPENSHIFT_TARGET_TOKEN" get all -n $ns

# Test access to other namespaces (should be forbidden)
kubectl --token="\$OPENSHIFT_TARGET_TOKEN" get pods -A

# Test with explicit server URL
kubectl --token="\$OPENSHIFT_TARGET_TOKEN" --server="\$OPENSHIFT_TARGET_URL" get all -n $ns
EOF
}

# Function to cleanup existing resources
cleanup_resources() {
    local ns="$1"
    local sa="$2"
    
    echo "=== kubectl-mtv Test Service Account Cleanup ==="
    echo "Platform: $PLATFORM"
    echo "Namespace: $ns"
    echo "Service Account: $sa"
    echo
    
    # Check if namespace exists
    if ! namespace_exists "$ns"; then
        echo "Namespace/project $ns does not exist - nothing to clean up"
        return 0
    fi
    
    echo "Cleaning up resources in namespace: $ns"
    
    # Delete secret if it exists
    echo "Deleting token secret if it exists..."
    if [[ "$PLATFORM" == "openshift" ]]; then
        oc delete secret "${sa}-token" -n "$ns" 2>/dev/null || echo "Secret ${sa}-token not found or already deleted"
    else
        kubectl delete secret "${sa}-token" -n "$ns" 2>/dev/null || echo "Secret ${sa}-token not found or already deleted"
    fi
    
    # Delete rolebinding (only for Kubernetes)
    if [[ "$PLATFORM" == "kubernetes" ]]; then
        echo "Deleting rolebinding if it exists..."
        kubectl delete rolebinding "${sa}-admin-binding" -n "$ns" 2>/dev/null || echo "RoleBinding ${sa}-admin-binding not found or already deleted"
    fi
    
    # Delete service account
    echo "Deleting service account if it exists..."
    if [[ "$PLATFORM" == "openshift" ]]; then
        oc delete sa "$sa" -n "$ns" 2>/dev/null || echo "Service account $sa not found or already deleted"
    else
        kubectl delete serviceaccount "$sa" -n "$ns" 2>/dev/null || echo "Service account $sa not found or already deleted"
    fi
    
    # Delete namespace/project
    echo "Deleting namespace/project..."
    if [[ "$PLATFORM" == "openshift" ]]; then
        oc delete project "$ns" 2>/dev/null || echo "Project $ns not found or already deleted"
    else
        kubectl delete namespace "$ns" 2>/dev/null || echo "Namespace $ns not found or already deleted"
    fi
    
    echo
    echo "=== Cleanup Complete! ==="
    echo "All resources for service account '$sa' in namespace '$ns' have been removed."
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -s|--service-account)
            SERVICE_ACCOUNT="$2"
            shift 2
            ;;
        -p|--platform)
            PLATFORM="$2"
            shift 2
            ;;
        -c|--cleanup)
            CLEANUP=true
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            usage
            exit 1
            ;;
    esac
done

# Auto-detect platform if not specified
if [[ -z "$PLATFORM" ]]; then
    PLATFORM=$(detect_platform)
    echo "Detected platform: $PLATFORM"
fi

# Validate platform
if [[ "$PLATFORM" != "kubernetes" && "$PLATFORM" != "openshift" ]]; then
    echo "ERROR: Platform must be either 'kubernetes' or 'openshift'" >&2
    exit 1
fi

# Handle cleanup mode
if [[ "$CLEANUP" == "true" ]]; then
    cleanup_resources "$NAMESPACE" "$SERVICE_ACCOUNT"
    exit 0
fi

echo "=== kubectl-mtv Test Service Account Setup ==="
echo "Platform: $PLATFORM"
echo "Namespace: $NAMESPACE"
echo "Service Account: $SERVICE_ACCOUNT"
echo

# Create namespace if it doesn't exist
if ! namespace_exists "$NAMESPACE"; then
    create_namespace "$NAMESPACE"
else
    echo "Namespace/project $NAMESPACE already exists"
fi

# Create service account
create_service_account "$NAMESPACE" "$SERVICE_ACCOUNT"

# Grant admin privileges
grant_admin_privileges "$NAMESPACE" "$SERVICE_ACCOUNT"

# Create long-term token
echo
if create_long_term_token "$NAMESPACE" "$SERVICE_ACCOUNT"; then
    # Get the token for verification commands
    TOKEN=$(kubectl -n "$NAMESPACE" get secret "${SERVICE_ACCOUNT}-token" -o go-template='{{ .data.token | base64decode }}' 2>/dev/null || echo "")
    
    # Show verification commands
    show_verification "$NAMESPACE" "$TOKEN"
    
    echo
    echo "=== Success! ==="
    echo "Service account '$SERVICE_ACCOUNT' created with admin privileges in namespace '$NAMESPACE'"
    echo "Long-term token created and stored in secret '${SERVICE_ACCOUNT}-token'"
    echo
    echo "For OpenShift provider tests, you now have:"
    echo "1. Token: (displayed above) - use as OPENSHIFT_TARGET_TOKEN"
    echo "2. API URL: (displayed above) - use as OPENSHIFT_TARGET_URL"
    echo
    echo "You can copy these values to your .env file or export them directly:"
    echo "export OPENSHIFT_TARGET_TOKEN=\"<token-from-above>\""
    echo "export OPENSHIFT_TARGET_URL=\"<api-url-from-above>\""
    echo
    echo "To clean up when done testing, run:"
    echo "$0 --cleanup --namespace $NAMESPACE --service-account $SERVICE_ACCOUNT"
else
    echo "WARNING: Token creation may have failed. Check the output above."
    exit 1
fi
