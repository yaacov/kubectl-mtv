---
name: forklift-types
description: Reference for Forklift CRD types, GVR constants, K8s client patterns, CRUD operations, and inventory API usage in kubectl-mtv. Use when working with Forklift resources, K8s API calls, or the inventory service.
---

# Forklift Types and K8s Patterns

kubectl-mtv interacts with Forklift CRDs via the K8s dynamic client and with the inventory service via HTTP.

## Import Path

```go
forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
```

Sub-packages:

```go
"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/plan"      // plan.Map
"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/provider"  // provider.Pair
"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/ref"       // ref.Ref
```

## CRD Types

| Type | Go Type | GVR Constant | Resource |
|------|---------|-------------|----------|
| Plan | `forkliftv1beta1.Plan` | `client.PlansGVR` | `plans` |
| Provider | `forkliftv1beta1.Provider` | `client.ProvidersGVR` | `providers` |
| NetworkMap | `forkliftv1beta1.NetworkMap` | `client.NetworkMapGVR` | `networkmaps` |
| StorageMap | `forkliftv1beta1.StorageMap` | `client.StorageMapGVR` | `storagemaps` |
| Host | `forkliftv1beta1.Host` | `client.HostsGVR` | `hosts` |
| Hook | (HookSpec) | `client.HooksGVR` | `hooks` |
| Migration | `forkliftv1beta1.Migration` | `client.MigrationsGVR` | `migrations` |

All GVRs use group `forklift.konveyor.io`, version `v1beta1`.

## GVR Constants

Defined in `pkg/util/client/client.go`:

```go
const Group = "forklift.konveyor.io"
const Version = "v1beta1"

var PlansGVR = schema.GroupVersionResource{Group: Group, Version: Version, Resource: "plans"}
var ProvidersGVR = schema.GroupVersionResource{Group: Group, Version: Version, Resource: "providers"}
var NetworkMapGVR = schema.GroupVersionResource{Group: Group, Version: Version, Resource: "networkmaps"}
var StorageMapGVR = schema.GroupVersionResource{Group: Group, Version: Version, Resource: "storagemaps"}
var HostsGVR = schema.GroupVersionResource{Group: Group, Version: Version, Resource: "hosts"}
var HooksGVR = schema.GroupVersionResource{Group: Group, Version: Version, Resource: "hooks"}
var MigrationsGVR = schema.GroupVersionResource{Group: Group, Version: Version, Resource: "migrations"}
```

## K8s Client

```go
import "github.com/yaacov/kubectl-mtv/pkg/util/client"

// Dynamic client for CRDs
dynamicClient, err := client.GetDynamicClient(configFlags)

// Typed clientset for core resources (Secrets, ConfigMaps)
k8sClient, err := client.GetKubernetesClientset(configFlags)
```

## CRUD Patterns

### List

```go
items, err := dynamicClient.Resource(client.PlansGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
// All namespaces:
items, err := dynamicClient.Resource(client.PlansGVR).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
```

### Get

```go
item, err := dynamicClient.Resource(client.PlansGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
```

### Create (typed -> unstructured)

```go
planObj := &forkliftv1beta1.Plan{
    TypeMeta: metav1.TypeMeta{APIVersion: "forklift.konveyor.io/v1beta1", Kind: "Plan"},
    ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
    Spec: forkliftv1beta1.PlanSpec{/* ... */},
}
unstructuredObj, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(planObj)
result, err := dynamicClient.Resource(client.PlansGVR).Namespace(namespace).Create(
    ctx, &unstructured.Unstructured{Object: unstructuredObj}, metav1.CreateOptions{})
```

### Delete

```go
err := dynamicClient.Resource(client.PlansGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
```

### Update

```go
existing, _ := dynamicClient.Resource(client.ProvidersGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
// Modify existing.Object fields...
_, err := dynamicClient.Resource(client.ProvidersGVR).Namespace(namespace).Update(ctx, existing, metav1.UpdateOptions{})
```

## Extracting Fields from Unstructured

```go
import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

name := item.GetName()
namespace := item.GetNamespace()
created := item.GetCreationTimestamp()

// Nested fields
value, found, err := unstructured.NestedString(item.Object, "spec", "provider", "name")
conditions, found, err := unstructured.NestedSlice(item.Object, "status", "conditions")
nested, found, err := unstructured.NestedMap(item.Object, "spec", "provider")
```

## Supporting Types

- `provider.Pair` -- source/destination provider references in a Plan
- `plan.Map` -- network and storage map references
- `ref.Ref` -- ID/Name reference for source networks/storage
- `forkliftv1beta1.NetworkPair`, `StoragePair` -- mapping pairs
- `forkliftv1beta1.DestinationNetwork`, `DestinationStorage` -- target side of mappings

## Inventory API

The inventory service provides VM, network, datastore, and host details from source providers.

### Client Setup

```go
httpClient, err := client.GetAuthenticatedHTTPClientWithInsecure(ctx, configFlags, inventoryURL, insecureSkipTLS)
```

### Discovery

```go
inventoryURL, err := client.DiscoverInventoryURL(ctx, configFlags)
```

This finds the Route with labels `app=forklift,service=forklift-inventory`.

### ProviderClient

`pkg/cmd/get/inventory/client.go` wraps inventory API calls:

```go
pc := inventory.NewProviderClientWithInsecure(configFlags, provider, inventoryURL, insecureSkipTLS)
vms, err := pc.GetVMs(ctx)
hosts, err := pc.GetHosts(ctx)
networks, err := pc.GetNetworks(ctx)
datastores, err := pc.GetDatastores(ctx)
```

### Namespace Resolution

```go
namespace := client.ResolveNamespaceWithAllFlag(configFlags, allNamespaces)
```

Returns `""` when `--all-namespaces` is set, otherwise resolves from kubeconfig context.

## Name Helpers

`pkg/util/client/client.go` provides name-listing helpers for shell completion:

- `client.GetAllPlanNames(ctx, configFlags, namespace)`
- `client.GetAllProviderNames(ctx, configFlags, namespace)`
- `client.GetAllHostNames(ctx, configFlags, namespace)`
- `client.GetAllHookNames(ctx, configFlags, namespace)`
- `client.GetAllNetworkMappingNames(ctx, configFlags, namespace)`
- `client.GetAllStorageMappingNames(ctx, configFlags, namespace)`
