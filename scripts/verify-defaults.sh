#!/bin/bash
# verify-defaults.sh
# Compares kubectl-mtv settings defaults against forklift operator defaults
#
# Usage:
#   ./scripts/verify-defaults.sh [FORKLIFT_PATH]
#
# Environment:
#   FORKLIFT_PATH - Path to forklift repository (default: ../../kubev2v/forklift)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KUBECTL_MTV_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Default forklift path (relative to kubectl-mtv root)
FORKLIFT_PATH="${1:-${FORKLIFT_PATH:-$KUBECTL_MTV_ROOT/../../kubev2v/forklift}}"

# Resolve to absolute path
FORKLIFT_PATH="$(cd "$FORKLIFT_PATH" 2>/dev/null && pwd)" || {
    echo -e "${RED}Error: Forklift repository not found at: $FORKLIFT_PATH${NC}"
    echo "Please provide the path to forklift repository as argument or set FORKLIFT_PATH"
    exit 1
}

FORKLIFT_DEFAULTS="$FORKLIFT_PATH/operator/roles/forkliftcontroller/defaults/main.yml"
KUBECTL_MTV_TYPES="$KUBECTL_MTV_ROOT/pkg/cmd/settings/types.go"

if [[ ! -f "$FORKLIFT_DEFAULTS" ]]; then
    echo -e "${RED}Error: Forklift defaults file not found: $FORKLIFT_DEFAULTS${NC}"
    exit 1
fi

if [[ ! -f "$KUBECTL_MTV_TYPES" ]]; then
    echo -e "${RED}Error: kubectl-mtv types.go not found: $KUBECTL_MTV_TYPES${NC}"
    exit 1
fi

echo "Comparing settings defaults..."
echo "  Forklift:    $FORKLIFT_DEFAULTS"
echo "  kubectl-mtv: $KUBECTL_MTV_TYPES"
echo ""

# Temporary files for comparison
FORKLIFT_SETTINGS=$(mktemp)
KUBECTL_MTV_SETTINGS=$(mktemp)
trap "rm -f $FORKLIFT_SETTINGS $KUBECTL_MTV_SETTINGS" EXIT

# Extract settings from forklift defaults/main.yml
# Format: setting_name: value
# Skip template expressions and complex values
grep -E '^[a-z_]+:' "$FORKLIFT_DEFAULTS" | \
    grep -v '{{' | \
    grep -v '^app_' | \
    grep -v '_name:' | \
    grep -v '_service_' | \
    grep -v '_deployment_' | \
    grep -v '_volume_path' | \
    grep -v '_tls_secret' | \
    grep -v '_issuer_' | \
    grep -v '_certificate_' | \
    grep -v '_route_name' | \
    grep -v '_console_name' | \
    grep -v '_display_name' | \
    grep -v '_subapp_name' | \
    grep -v '_state:' | \
    grep -v 'forklift_' | \
    sed 's/: /=/; s/"//g; s/'\''//g' | \
    sort > "$FORKLIFT_SETTINGS"

# Extract settings from kubectl-mtv types.go
# Look for Default: values in the setting definitions
# Format: setting_name=value
# Use perl for better regex support (available on macOS)
perl -ne '
    if (/Name:\s*"([a-z_]+)"/) {
        $name = $1;
    }
    if (/Default:\s*(.+?),/ && $name) {
        $val = $1;
        $val =~ s/^\s+|\s+$//g;
        $val =~ s/"//g;
        $val =~ s/'\''//g;
        print "$name=$val\n";
        $name = "";
    }
' "$KUBECTL_MTV_TYPES" | sort > "$KUBECTL_MTV_SETTINGS"

# Compare the settings
echo "=== Settings Comparison ==="
echo ""

mismatches=0

# Check each forklift setting
while IFS='=' read -r name value; do
    [[ -z "$name" ]] && continue
    
    # Look for this setting in kubectl-mtv
    kubectl_value=$(grep "^${name}=" "$KUBECTL_MTV_SETTINGS" 2>/dev/null | cut -d= -f2-)
    
    if [[ -z "$kubectl_value" ]]; then
        # Setting exists in forklift but not in kubectl-mtv
        # This is expected for many settings
        continue
    fi
    
    # Normalize values for comparison
    forklift_normalized=$(echo "$value" | tr '[:upper:]' '[:lower:]' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    kubectl_normalized=$(echo "$kubectl_value" | tr '[:upper:]' '[:lower:]' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    
    if [[ "$forklift_normalized" != "$kubectl_normalized" ]]; then
        echo -e "${YELLOW}MISMATCH:${NC} $name"
        echo "  Forklift:    $value"
        echo "  kubectl-mtv: $kubectl_value"
        echo ""
        ((mismatches++))
    fi
done < "$FORKLIFT_SETTINGS"

# Summary
echo "=== Summary ==="
if [[ $mismatches -eq 0 ]]; then
    echo -e "${GREEN}No mismatches found!${NC}"
else
    echo -e "${YELLOW}Found $mismatches setting(s) with different defaults${NC}"
fi

echo ""
echo "Settings in kubectl-mtv: $(wc -l < "$KUBECTL_MTV_SETTINGS" | tr -d ' ')"
echo "Settings in forklift:    $(wc -l < "$FORKLIFT_SETTINGS" | tr -d ' ')"

exit $mismatches
