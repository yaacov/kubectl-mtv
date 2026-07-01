package azure

import (
	"context"
	"fmt"
	"sort"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/ref"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/klog/v2"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/create/plan/storage/fetchers"
	"github.com/yaacov/kubectl-mtv/pkg/cmd/get/inventory"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
)

// AzureStorageFetcher implements storage fetching for Azure providers
type AzureStorageFetcher struct{}

// NewAzureStorageFetcher creates a new Azure storage fetcher
func NewAzureStorageFetcher() fetchers.StorageFetcher {
	return &AzureStorageFetcher{}
}

// FetchSourceStorages fetches disk type SKUs from Azure provider
func (f *AzureStorageFetcher) FetchSourceStorages(ctx context.Context, configFlags *genericclioptions.ConfigFlags, providerName, namespace, inventoryURL string, _ []string, insecureSkipTLS bool) ([]ref.Ref, error) {
	klog.V(4).Infof("DEBUG: Azure - Fetching source storage types from provider: %s", providerName)

	provider, err := inventory.GetProviderByName(ctx, configFlags, providerName, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get Azure provider: %v", err)
	}

	storageInventory, err := client.FetchProviderInventoryWithInsecure(ctx, configFlags, inventoryURL, provider, "storages?detail=4", insecureSkipTLS)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Azure storage inventory: %v", err)
	}

	storageArray, ok := storageInventory.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected data format: expected array for storage inventory")
	}

	// Collect unique disk type SKUs
	var sourceStorages []ref.Ref
	diskTypeSet := make(map[string]struct{})

	for _, item := range storageArray {
		storage, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Azure disk types have a "name" field with the SKU (e.g., "Premium_LRS")
		name, _ := storage["name"].(string)
		id, _ := storage["id"].(string)

		if name != "" {
			if _, seen := diskTypeSet[name]; !seen {
				diskTypeSet[name] = struct{}{}
				sourceStorages = append(sourceStorages, ref.Ref{
					ID:   id,
					Name: name,
				})
			}
		}
	}

	sort.Slice(sourceStorages, func(i, j int) bool {
		return sourceStorages[i].Name < sourceStorages[j].Name
	})

	klog.V(4).Infof("DEBUG: Azure - Found %d source storage types", len(sourceStorages))
	return sourceStorages, nil
}

// FetchTargetStorages fetches target storage from Azure provider (not typically used as Azure is usually source)
func (f *AzureStorageFetcher) FetchTargetStorages(ctx context.Context, configFlags *genericclioptions.ConfigFlags, providerName, namespace, inventoryURL string, insecureSkipTLS bool) ([]forkliftv1beta1.DestinationStorage, error) {
	klog.V(4).Infof("DEBUG: Azure - Fetching target storage (Azure is typically not a migration target)")
	return []forkliftv1beta1.DestinationStorage{}, nil
}
