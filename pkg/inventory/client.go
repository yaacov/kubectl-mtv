package inventory

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"

	"github.com/yaacov/kubectl-mtv/pkg/client"
)

// ProviderClient provides a unified client for all provider types
type ProviderClient struct {
	configFlags  *genericclioptions.ConfigFlags
	provider     *unstructured.Unstructured
	inventoryURL string
}

// NewProviderClient creates a new provider client
func NewProviderClient(configFlags *genericclioptions.ConfigFlags, provider *unstructured.Unstructured, inventoryURL string) *ProviderClient {
	return &ProviderClient{
		configFlags:  configFlags,
		provider:     provider,
		inventoryURL: inventoryURL,
	}
}

// GetResource fetches a resource from the provider using the specified path
func (pc *ProviderClient) GetResource(resourcePath string) (interface{}, error) {
	// Get provider info for logging
	providerName := pc.GetProviderName()
	providerNamespace := pc.GetProviderNamespace()
	providerType, _ := pc.GetProviderType()
	providerUID, _ := pc.GetProviderUID()

	// Log the inventory fetch request
	klog.V(2).Infof("Fetching inventory from provider %s/%s (type=%s, uid=%s) - path: %s, baseURL: %s",
		providerNamespace, providerName, providerType, providerUID, resourcePath, pc.inventoryURL)

	result, err := client.FetchProviderInventory(pc.configFlags, pc.inventoryURL, pc.provider, resourcePath)

	if err != nil {
		klog.V(1).Infof("Failed to fetch inventory from provider %s/%s - path: %s, error: %v",
			providerNamespace, providerName, resourcePath, err)
		return nil, err
	}

	// Log success with some response details
	resultType := "unknown"
	resultSize := 0

	switch v := result.(type) {
	case []interface{}:
		resultType = "array"
		resultSize = len(v)
	case map[string]interface{}:
		resultType = "object"
		resultSize = len(v)
	}

	klog.V(2).Infof("Successfully fetched inventory from provider %s/%s - path: %s, result_type: %s, result_size: %d",
		providerNamespace, providerName, resourcePath, resultType, resultSize)

	// Dump the full response at trace level (v=3)
	klog.V(3).Infof("Full inventory response from provider %s/%s - path: %s, response: %+v",
		providerNamespace, providerName, resourcePath, result)

	return result, nil
}

// GetResourceWithQuery fetches a resource with query parameters
func (pc *ProviderClient) GetResourceWithQuery(resourcePath, query string) (interface{}, error) {
	if query != "" {
		resourcePath = fmt.Sprintf("%s?%s", resourcePath, query)
	}
	return pc.GetResource(resourcePath)
}

// GetResourceCollection fetches a collection of resources
func (pc *ProviderClient) GetResourceCollection(collection string, detail int) (interface{}, error) {
	return pc.GetResourceWithQuery(collection, fmt.Sprintf("detail=%d", detail))
}

// GetResourceByID fetches a specific resource by ID
func (pc *ProviderClient) GetResourceByID(collection, id string) (interface{}, error) {
	return pc.GetResource(fmt.Sprintf("%s/%s", collection, id))
}

// oVirt Provider Resources
func (pc *ProviderClient) GetDataCenters(detail int) (interface{}, error) {
	return pc.GetResourceCollection("datacenters", detail)
}

func (pc *ProviderClient) GetDataCenter(id string) (interface{}, error) {
	return pc.GetResourceByID("datacenters", id)
}

func (pc *ProviderClient) GetClusters(detail int) (interface{}, error) {
	return pc.GetResourceCollection("clusters", detail)
}

func (pc *ProviderClient) GetCluster(id string) (interface{}, error) {
	return pc.GetResourceByID("clusters", id)
}

func (pc *ProviderClient) GetHosts(detail int) (interface{}, error) {
	return pc.GetResourceCollection("hosts", detail)
}

func (pc *ProviderClient) GetHost(id string) (interface{}, error) {
	return pc.GetResourceByID("hosts", id)
}

func (pc *ProviderClient) GetVMs(detail int) (interface{}, error) {
	return pc.GetResourceCollection("vms", detail)
}

func (pc *ProviderClient) GetVM(id string) (interface{}, error) {
	return pc.GetResourceByID("vms", id)
}

func (pc *ProviderClient) GetStorageDomains(detail int) (interface{}, error) {
	return pc.GetResourceCollection("storagedomains", detail)
}

func (pc *ProviderClient) GetStorageDomain(id string) (interface{}, error) {
	return pc.GetResourceByID("storagedomains", id)
}

func (pc *ProviderClient) GetNetworks(detail int) (interface{}, error) {
	return pc.GetResourceCollection("networks", detail)
}

func (pc *ProviderClient) GetNetwork(id string) (interface{}, error) {
	return pc.GetResourceByID("networks", id)
}

func (pc *ProviderClient) GetDisks(detail int) (interface{}, error) {
	return pc.GetResourceCollection("disks", detail)
}

func (pc *ProviderClient) GetDisk(id string) (interface{}, error) {
	return pc.GetResourceByID("disks", id)
}

func (pc *ProviderClient) GetDiskProfiles(detail int) (interface{}, error) {
	return pc.GetResourceCollection("diskprofiles", detail)
}

func (pc *ProviderClient) GetDiskProfile(id string) (interface{}, error) {
	return pc.GetResourceByID("diskprofiles", id)
}

func (pc *ProviderClient) GetNICProfiles(detail int) (interface{}, error) {
	return pc.GetResourceCollection("nicprofiles", detail)
}

func (pc *ProviderClient) GetNICProfile(id string) (interface{}, error) {
	return pc.GetResourceByID("nicprofiles", id)
}

func (pc *ProviderClient) GetWorkloads(detail int) (interface{}, error) {
	return pc.GetResourceCollection("workloads", detail)
}

func (pc *ProviderClient) GetWorkload(id string) (interface{}, error) {
	return pc.GetResourceByID("workloads", id)
}

func (pc *ProviderClient) GetTree() (interface{}, error) {
	return pc.GetResource("tree")
}

func (pc *ProviderClient) GetClusterTree() (interface{}, error) {
	return pc.GetResource("tree/cluster")
}

// vSphere Provider Resources (aliases to generic resources with vSphere context)
func (pc *ProviderClient) GetDatastores(detail int) (interface{}, error) {
	// vSphere datastores map to generic storage resources
	return pc.GetResourceCollection("datastores", detail)
}

func (pc *ProviderClient) GetDatastore(id string) (interface{}, error) {
	return pc.GetResourceByID("datastores", id)
}

func (pc *ProviderClient) GetResourcePools(detail int) (interface{}, error) {
	return pc.GetResourceCollection("resourcepools", detail)
}

func (pc *ProviderClient) GetResourcePool(id string) (interface{}, error) {
	return pc.GetResourceByID("resourcepools", id)
}

func (pc *ProviderClient) GetFolders(detail int) (interface{}, error) {
	return pc.GetResourceCollection("folders", detail)
}

func (pc *ProviderClient) GetFolder(id string) (interface{}, error) {
	return pc.GetResourceByID("folders", id)
}

// OpenStack Provider Resources
func (pc *ProviderClient) GetInstances(detail int) (interface{}, error) {
	// OpenStack instances are equivalent to VMs
	return pc.GetResourceCollection("instances", detail)
}

func (pc *ProviderClient) GetInstance(id string) (interface{}, error) {
	return pc.GetResourceByID("instances", id)
}

func (pc *ProviderClient) GetImages(detail int) (interface{}, error) {
	return pc.GetResourceCollection("images", detail)
}

func (pc *ProviderClient) GetImage(id string) (interface{}, error) {
	return pc.GetResourceByID("images", id)
}

func (pc *ProviderClient) GetFlavors(detail int) (interface{}, error) {
	return pc.GetResourceCollection("flavors", detail)
}

func (pc *ProviderClient) GetFlavor(id string) (interface{}, error) {
	return pc.GetResourceByID("flavors", id)
}

func (pc *ProviderClient) GetSubnets(detail int) (interface{}, error) {
	return pc.GetResourceCollection("subnets", detail)
}

func (pc *ProviderClient) GetSubnet(id string) (interface{}, error) {
	return pc.GetResourceByID("subnets", id)
}

func (pc *ProviderClient) GetPorts(detail int) (interface{}, error) {
	return pc.GetResourceCollection("ports", detail)
}

func (pc *ProviderClient) GetPort(id string) (interface{}, error) {
	return pc.GetResourceByID("ports", id)
}

func (pc *ProviderClient) GetVolumeTypes(detail int) (interface{}, error) {
	return pc.GetResourceCollection("volumetypes", detail)
}

func (pc *ProviderClient) GetVolumeType(id string) (interface{}, error) {
	return pc.GetResourceByID("volumetypes", id)
}

func (pc *ProviderClient) GetVolumes(detail int) (interface{}, error) {
	return pc.GetResourceCollection("volumes", detail)
}

func (pc *ProviderClient) GetVolume(id string) (interface{}, error) {
	return pc.GetResourceByID("volumes", id)
}

func (pc *ProviderClient) GetSecurityGroups(detail int) (interface{}, error) {
	return pc.GetResourceCollection("securitygroups", detail)
}

func (pc *ProviderClient) GetSecurityGroup(id string) (interface{}, error) {
	return pc.GetResourceByID("securitygroups", id)
}

func (pc *ProviderClient) GetFloatingIPs(detail int) (interface{}, error) {
	return pc.GetResourceCollection("floatingips", detail)
}

func (pc *ProviderClient) GetFloatingIP(id string) (interface{}, error) {
	return pc.GetResourceByID("floatingips", id)
}

func (pc *ProviderClient) GetProjects(detail int) (interface{}, error) {
	return pc.GetResourceCollection("projects", detail)
}

func (pc *ProviderClient) GetProject(id string) (interface{}, error) {
	return pc.GetResourceByID("projects", id)
}

// OpenShift Provider Resources
func (pc *ProviderClient) GetStorageClasses(detail int) (interface{}, error) {
	return pc.GetResourceCollection("storageclasses", detail)
}

func (pc *ProviderClient) GetStorageClass(id string) (interface{}, error) {
	return pc.GetResourceByID("storageclasses", id)
}

func (pc *ProviderClient) GetPersistentVolumeClaims(detail int) (interface{}, error) {
	return pc.GetResourceCollection("persistentvolumeclaims", detail)
}

func (pc *ProviderClient) GetPersistentVolumeClaim(id string) (interface{}, error) {
	return pc.GetResourceByID("persistentvolumeclaims", id)
}

func (pc *ProviderClient) GetNamespaces(detail int) (interface{}, error) {
	return pc.GetResourceCollection("namespaces", detail)
}

func (pc *ProviderClient) GetNamespace(id string) (interface{}, error) {
	return pc.GetResourceByID("namespaces", id)
}

func (pc *ProviderClient) GetDataVolumes(detail int) (interface{}, error) {
	return pc.GetResourceCollection("datavolumes", detail)
}

func (pc *ProviderClient) GetDataVolume(id string) (interface{}, error) {
	return pc.GetResourceByID("datavolumes", id)
}

// OVA Provider Resources
func (pc *ProviderClient) GetOVAFiles(detail int) (interface{}, error) {
	return pc.GetResourceCollection("ovafiles", detail)
}

func (pc *ProviderClient) GetOVAFile(id string) (interface{}, error) {
	return pc.GetResourceByID("ovafiles", id)
}

// Generic helper functions for provider-agnostic operations
func (pc *ProviderClient) GetProviderType() (string, error) {
	providerType, found, err := unstructured.NestedString(pc.provider.Object, "spec", "type")
	if err != nil || !found {
		return "", fmt.Errorf("provider type not found or error retrieving it: %v", err)
	}
	return providerType, nil
}

func (pc *ProviderClient) GetProviderUID() (string, error) {
	providerUID, found, err := unstructured.NestedString(pc.provider.Object, "metadata", "uid")
	if err != nil || !found {
		return "", fmt.Errorf("provider UID not found or error retrieving it: %v", err)
	}
	return providerUID, nil
}

func (pc *ProviderClient) GetProviderName() string {
	return pc.provider.GetName()
}

func (pc *ProviderClient) GetProviderNamespace() string {
	return pc.provider.GetNamespace()
}
