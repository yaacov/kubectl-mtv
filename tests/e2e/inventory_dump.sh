#!/bin/bash

# inventory_dump.sh - Dump inventory data to JSON files for testing
# This script finds all test namespaces, discovers providers, and dumps
# VMs, networks, and storage inventory to JSON files for each provider in each namespace.

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}/inventory_json_dump"
KUBECTL_MTV="${KUBECTL_MTV:-../../kubectl-mtv}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Function to print log messages
log_info() {
    echo "[INFO] $1"
}

log_success() {
    echo "[SUCCESS] $1"
}

log_warning() {
    echo "[WARNING] $1"
}

log_error() {
    echo "[ERROR] $1"
}

# Function to find all test namespaces
find_all_test_namespaces() {
    # Look for namespaces starting with kubectl-mtv-shared-
    local test_namespaces
    test_namespaces=$(kubectl get namespaces -o name | grep "kubectl-mtv-shared-" | sed 's/namespace\///' | sort -r)
    
    if [[ -z "$test_namespaces" ]]; then
        # Default to a common test namespace
        echo "openshift-mtv"
        return
    fi
    
    # Return all test namespaces
    echo "$test_namespaces"
}

# Function to find providers in the namespace
find_providers() {
    local namespace=$1
    
    # Get providers using kubectl-mtv
    local providers
    providers=$($KUBECTL_MTV get provider -n "$namespace" -o json 2>/dev/null | jq -r '.items[]?.metadata.name // empty' | sort)
    
    if [[ -z "$providers" ]]; then
        # Try different provider CRD names
        for crd in providers.forklift.konveyor.io providers.migration.openshift.io; do
            providers=$(kubectl get "$crd" -n "$namespace" -o json 2>/dev/null | jq -r '.items[]?.metadata.name // empty' | sort || true)
            if [[ -n "$providers" ]]; then
                break
            fi
        done
    fi
    
    echo "$providers"
}

# Function to get provider type
get_provider_type() {
    local provider=$1
    local namespace=$2
    
    # Try to get provider details to determine type
    local provider_type
    provider_type=$(kubectl get providers.forklift.konveyor.io "$provider" -n "$namespace" -o json 2>/dev/null | jq -r '.spec.type // empty' || true)
    
    if [[ -z "$provider_type" ]]; then
        provider_type=$(kubectl get providers.migration.openshift.io "$provider" -n "$namespace" -o json 2>/dev/null | jq -r '.spec.type // empty' || true)
    fi
    
    # If we can't determine the type, try to infer from name
    if [[ -z "$provider_type" ]]; then
        case "$provider" in
            *vsphere*|*vcenter*) provider_type="vsphere" ;;
            *ovirt*|*rhv*) provider_type="ovirt" ;;
            *openstack*) provider_type="openstack" ;;
            *openshift*|*ocp*) provider_type="openshift" ;;
            *ova*) provider_type="ova" ;;
            *) provider_type="unknown" ;;
        esac
    fi
    
    echo "$provider_type"
}

# Function to dump inventory for a provider
dump_provider_inventory() {
    local provider=$1
    local provider_type=$2
    local namespace=$3
    local provider_dir="$OUTPUT_DIR/$namespace/$provider"
    
    # Create provider directory under namespace
    mkdir -p "$provider_dir"
    
    # Dump VMs
    if ! $KUBECTL_MTV get inventory vm "$provider" -n "$namespace" -o json > "$provider_dir/vms.json" 2>/dev/null; then
        echo '{"items": [], "error": "failed to fetch VMs"}' > "$provider_dir/vms.json"
    fi
    
    # Dump Networks
    if ! $KUBECTL_MTV get inventory network "$provider" -n "$namespace" -o json > "$provider_dir/networks.json" 2>/dev/null; then
        echo '{"items": [], "error": "failed to fetch networks"}' > "$provider_dir/networks.json"
    fi
    
    # Dump Storage
    if ! $KUBECTL_MTV get inventory storage "$provider" -n "$namespace" -o json > "$provider_dir/storage.json" 2>/dev/null; then
        echo '{"items": [], "error": "failed to fetch storage"}' > "$provider_dir/storage.json"
    fi
    
    # Dump provider-specific additional resources
    case "$provider_type" in
        "vsphere")
            $KUBECTL_MTV get inventory datastore "$provider" -n "$namespace" -o json > "$provider_dir/datastores.json" 2>/dev/null || echo '{"items": []}' > "$provider_dir/datastores.json"
            $KUBECTL_MTV get inventory resource-pool "$provider" -n "$namespace" -o json > "$provider_dir/resource_pools.json" 2>/dev/null || echo '{"items": []}' > "$provider_dir/resource_pools.json"
            $KUBECTL_MTV get inventory folder "$provider" -n "$namespace" -o json > "$provider_dir/folders.json" 2>/dev/null || echo '{"items": []}' > "$provider_dir/folders.json"
            ;;
        "ovirt")
            $KUBECTL_MTV get inventory disk-profile "$provider" -n "$namespace" -o json > "$provider_dir/disk_profiles.json" 2>/dev/null || echo '{"items": []}' > "$provider_dir/disk_profiles.json"
            $KUBECTL_MTV get inventory nic-profile "$provider" -n "$namespace" -o json > "$provider_dir/nic_profiles.json" 2>/dev/null || echo '{"items": []}' > "$provider_dir/nic_profiles.json"
            ;;
        "openstack")
            $KUBECTL_MTV get inventory instance "$provider" -n "$namespace" -o json > "$provider_dir/instances.json" 2>/dev/null || echo '{"items": []}' > "$provider_dir/instances.json"
            $KUBECTL_MTV get inventory image "$provider" -n "$namespace" -o json > "$provider_dir/images.json" 2>/dev/null || echo '{"items": []}' > "$provider_dir/images.json"
            $KUBECTL_MTV get inventory flavor "$provider" -n "$namespace" -o json > "$provider_dir/flavors.json" 2>/dev/null || echo '{"items": []}' > "$provider_dir/flavors.json"
            $KUBECTL_MTV get inventory project "$provider" -n "$namespace" -o json > "$provider_dir/projects.json" 2>/dev/null || echo '{"items": []}' > "$provider_dir/projects.json"
            ;;
        "openshift")
            $KUBECTL_MTV get inventory pvc "$provider" -n "$namespace" -o json > "$provider_dir/pvcs.json" 2>/dev/null || echo '{"items": []}' > "$provider_dir/pvcs.json"
            $KUBECTL_MTV get inventory datavolume "$provider" -n "$namespace" -o json > "$provider_dir/datavolumes.json" 2>/dev/null || echo '{"items": []}' > "$provider_dir/datavolumes.json"
            ;;
    esac
    
    # Add provider metadata
    cat > "$provider_dir/metadata.json" <<EOF
{
    "provider_name": "$provider",
    "provider_type": "$provider_type",
    "namespace": "$namespace",
    "dump_timestamp": "$TIMESTAMP",
    "kubectl_mtv_version": "$($KUBECTL_MTV version -o json 2>/dev/null | jq -r '.clientVersion // "unknown"' || echo "unknown")"
}
EOF
}

# Function to create summary report
create_summary() {
    local namespaces="$1"
    local summary_file="$OUTPUT_DIR/summary.json"
    
    # Count total files and providers across all namespaces
    local total_provider_count=0
    local total_namespace_count
    total_namespace_count=$(echo "$namespaces" | wc -l)
    
    local total_files
    total_files=$(find "$OUTPUT_DIR" -name "*.json" | wc -l)
    
    cat > "$summary_file" <<EOF
{
    "dump_info": {
        "timestamp": "$TIMESTAMP",
        "output_directory": "$OUTPUT_DIR",
        "total_namespaces": $total_namespace_count,
        "total_json_files": $total_files
    },
    "namespaces": [
EOF
    
    # Add namespace summaries
    local namespace_first=true
    while IFS= read -r namespace; do
        if [[ -n "$namespace" && -d "$OUTPUT_DIR/$namespace" ]]; then
            if [[ "$namespace_first" == "true" ]]; then
                namespace_first=false
            else
                echo "," >> "$summary_file"
            fi
            
            local namespace_provider_count
            namespace_provider_count=$(find "$OUTPUT_DIR/$namespace" -maxdepth 1 -type d ! -path "$OUTPUT_DIR/$namespace" | wc -l)
            total_provider_count=$((total_provider_count + namespace_provider_count))
            
            cat >> "$summary_file" <<EOF
        {
            "namespace": "$namespace",
            "provider_count": $namespace_provider_count,
            "providers": [
EOF
            
            # Add provider summaries for this namespace
            local provider_first=true
            for provider_dir in "$OUTPUT_DIR/$namespace"/*/; do
                if [[ -d "$provider_dir" && -f "$provider_dir/metadata.json" ]]; then
                    if [[ "$provider_first" == "true" ]]; then
                        provider_first=false
                    else
                        echo "," >> "$summary_file"
                    fi
                    
                    local provider_name
                    provider_name=$(jq -r '.provider_name' "$provider_dir/metadata.json")
                    
                    local provider_type
                    provider_type=$(jq -r '.provider_type' "$provider_dir/metadata.json")
                    
                    local vm_count
                    vm_count=$(jq '.items | length' "$provider_dir/vms.json" 2>/dev/null || echo "0")
                    
                    local network_count
                    network_count=$(jq '.items | length' "$provider_dir/networks.json" 2>/dev/null || echo "0")
                    
                    local storage_count
                    storage_count=$(jq '.items | length' "$provider_dir/storage.json" 2>/dev/null || echo "0")
                    
                    cat >> "$summary_file" <<EOF
                {
                    "name": "$provider_name",
                    "type": "$provider_type",
                    "resources": {
                        "vms": $vm_count,
                        "networks": $network_count,
                        "storage": $storage_count
                    }
                }
EOF
                fi
            done
            
            cat >> "$summary_file" <<EOF
            ]
        }
EOF
        fi
    done <<< "$namespaces"
    
    # Update the dump_info with total provider count
    cat >> "$summary_file" <<EOF
    ],
    "totals": {
        "total_providers": $total_provider_count
    }
}
EOF
}

# Main execution
main() {
    log_info "MTV Inventory JSON Dump Tool"
    log_info "============================"
    
    # Check prerequisites
    if ! command -v jq &> /dev/null; then
        log_error "jq is required but not installed. Please install jq."
        exit 1
    fi
    
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is required but not installed."
        exit 1
    fi
    
    if ! command -v "$KUBECTL_MTV" &> /dev/null; then
        log_error "$KUBECTL_MTV is not available. Please ensure kubectl-mtv is installed and in PATH."
        exit 1
    fi
    
    # Create output directory
    rm -rf "$OUTPUT_DIR"
    mkdir -p "$OUTPUT_DIR"
    log_info "Output directory: $OUTPUT_DIR"
    
    # Find all test namespaces
    local namespaces
    namespaces=$(find_all_test_namespaces)
    log_success "Found namespaces: $(echo "$namespaces" | tr '\n' ' ')"
    
    # Process each namespace
    local total_providers=0
    while IFS= read -r namespace; do
        if [[ -n "$namespace" ]]; then
            log_info "Processing namespace: $namespace"
            
            # Find providers in this namespace
            local providers
            providers=$(find_providers "$namespace")
            
            if [[ -z "$providers" ]]; then
                log_warning "No providers found in namespace $namespace"
                continue
            fi
            
            local provider_count
            provider_count=$(echo "$providers" | wc -l)
            total_providers=$((total_providers + provider_count))
            
            log_success "Found $provider_count providers in $namespace: $(echo "$providers" | tr '\n' ' ')"
            
            # Dump inventory for each provider in this namespace
            while IFS= read -r provider; do
                if [[ -n "$provider" ]]; then
                    local provider_type
                    provider_type=$(get_provider_type "$provider" "$namespace")
                    log_info "Dumping inventory for provider: $provider (type: $provider_type) in namespace: $namespace"
                    dump_provider_inventory "$provider" "$provider_type" "$namespace"
                fi
            done <<< "$providers"
        fi
    done <<< "$namespaces"
    
    if [[ $total_providers -eq 0 ]]; then
        log_error "No providers found in any namespace"
        exit 1
    fi
    
    # Create summary
    create_summary "$namespaces"
    
    log_success "Inventory dump completed successfully!"
    log_info "Processed $(echo "$namespaces" | wc -l) namespaces with $total_providers total providers"
    log_info "Results saved to: $OUTPUT_DIR"
    log_info "View summary: cat $OUTPUT_DIR/summary.json | jq"
}

# Run main function
main "$@" 