package tools

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yaacov/kubectl-mtv/mcp-go/pkg/mtvmcp"
)

// VirtctlServiceManagementInput represents the input for VirtctlServiceManagement
type VirtctlServiceManagementInput struct {
	Operation    string                 `json:"operation" jsonschema:"Service operation (expose unexpose)"`
	ResourceName string                 `json:"resource_name" jsonschema:"Name of the VM or resource to manage"`
	ResourceType string                 `json:"resource_type,omitempty" jsonschema:"Resource type (vm vmi) (optional default: vm)"`
	Namespace    string                 `json:"namespace,omitempty" jsonschema:"Kubernetes namespace (optional)"`
	ExposeConfig map[string]interface{} `json:"expose_config,omitempty" jsonschema:"Configuration for service exposure (optional)"`
	ServiceName  string                 `json:"service_name,omitempty" jsonschema:"Name of the service for unexpose operation (optional)"`
}

// GetVirtctlServiceManagementTool returns the tool definition
func GetVirtctlServiceManagementTool() *mcp.Tool {
	return &mcp.Tool{
		Name: "VirtctlServiceManagement",
		Description: `Network services and connectivity management.

    PREREQUISITES:
    - VM must be running with network interfaces
    - For expose: VM must have services listening on target ports
    - Network connectivity: Ensure firewall rules allow traffic
    - Service mesh: Check if service mesh policies affect traffic

    OPERATION GUIDE:

    EXPOSE: Create Kubernetes Service for VM external access
    - Best for: Production workloads, web servers, databases, APIs
    - Creates stable endpoint for VM services
    - Supports LoadBalancer, NodePort, ClusterIP service types
    - Use for persistent access to VM services

    UNEXPOSE: Remove Kubernetes Service for VM
    - Best for: Removing temporary access, decommissioning services
    - Cleans up service resources and endpoints
    - Active connections to service will be terminated

    SERVICE TYPE SELECTION:

    LoadBalancer:
    - Best for: Production internet-facing services
    - Requires cloud provider load balancer support
    - Gets external IP address
    - Supports SSL termination, health checks

    NodePort:
    - Best for: Testing, development, internal services
    - Exposes service on all cluster nodes
    - Uses high-numbered port (30000-32767)
    - Accessible via any node IP

    ClusterIP (default):
    - Best for: Internal services, microservices communication
    - Only accessible from within cluster
    - Most secure option for internal communication

    PORT CONFIGURATION:

    Common Service Patterns:
    - Web servers: port=80, target_port=8080
    - HTTPS services: port=443, target_port=8443
    - SSH access: port=22, target_port=22
    - Databases: port=5432, target_port=5432 (PostgreSQL)
    - APIs: port=8080, target_port=8080

    TROUBLESHOOTING:

    Common Issues:
    - "Service not accessible": Check VM is running and listening on port
    - "Connection refused": Verify firewall rules inside VM
    - "LoadBalancer pending": Cloud provider may not support LoadBalancer
    - "Port already in use": Choose different external port

    Debugging Steps:
    1. Check VM status: kubectl get vmi {resource_name} -n {namespace}
    2. Test VM port: virtctl console {resource_name} then netstat -tlnp
    3. Check service: kubectl get svc {service_name} -n {namespace}
    4. Verify endpoints: kubectl get endpoints {service_name} -n {namespace}
    5. Test connectivity: kubectl port-forward svc/{service_name} local:remote

    Best Practices:
    - Use descriptive service names: web-server-svc, db-primary-svc
    - Add resource labels for monitoring and discovery
    - Use ClusterIP for internal service communication
    - Reserve LoadBalancer for internet-facing services
    - Test port accessibility before creating services

    Args:
        operation: Service operation (expose, unexpose)
        resource_name: Name of the VM or resource to manage
        resource_type: Resource type (vm, vmi) (optional, default: vm)
        namespace: Kubernetes namespace (optional)
        expose_config: Configuration for service exposure (optional)
        service_name: Name of the service for unexpose operation (optional)

    Expose Configuration Examples:
        # Basic web service
        expose_config = {
            "port": 80,
            "target_port": 8080,
            "service_type": "LoadBalancer",
            "service_name": "web-service"
        }

        # Internal database service
        expose_config = {
            "port": 5432,
            "target_port": 5432,
            "service_type": "ClusterIP",
            "service_name": "postgres-internal"
        }

        # Development service with NodePort
        expose_config = {
            "port": 3000,
            "target_port": 3000,
            "service_type": "NodePort",
            "service_name": "dev-app"
        }

    Workflow Examples:
        # Expose production web service
        virtctl_service_management(operation="expose", resource_name="web-server", resource_type="vm", namespace="production",
                                 expose_config={
                                     "port": 80,
                                     "target_port": 8080,
                                     "service_type": "LoadBalancer",
                                     "service_name": "web-service"
                                 })

        # Expose internal API service
        virtctl_service_management(operation="expose", resource_name="api-server", resource_type="vm",
                                 expose_config={
                                     "port": 8080,
                                     "target_port": 8080,
                                     "service_type": "ClusterIP"
                                 })

        # Remove service
        virtctl_service_management(operation="unexpose", resource_name="test-vm", resource_type="vm", namespace="development",
                                 service_name="test-service")`,
	}
}

func HandleVirtctlServiceManagement(ctx context.Context, req *mcp.CallToolRequest, input VirtctlServiceManagementInput) (*mcp.CallToolResult, any, error) {
	// Validate operation
	validOps := map[string]bool{"expose": true, "unexpose": true}
	if !validOps[input.Operation] {
		return nil, "", fmt.Errorf("invalid operation: %s. Valid operations: expose, unexpose", input.Operation)
	}

	// Build command
	var args []string
	var output string
	var err error

	namespace := mtvmcp.ResolveNamespace(input.Namespace)

	if input.Operation == "expose" {
		resType := input.ResourceType
		if resType == "" {
			resType = "vm"
		}
		args = []string{"expose", resType, input.ResourceName}
		if namespace != "" {
			args = append(args, "-n", namespace)
		}
		output, err = mtvmcp.RunVirtctlCommand(args)
		if err != nil {
			return nil, "", err
		}
		return nil, output, nil
	} else {
		// Use kubectl delete service for unexpose operation
		args = []string{"delete", "service", input.ServiceName}
		if namespace != "" {
			args = append(args, "-n", namespace)
		}
		result, err := mtvmcp.RunKubectlCommand(args)
		if err != nil {
			return nil, "", err
		}

		// Unmarshal the full CommandResponse to provide complete diagnostic information
		data, err := mtvmcp.UnmarshalJSONResponse(result)
		if err != nil {
			return nil, "", err
		}
		return nil, data, nil
	}
}
