---
name: add-cobra-command
description: Step-by-step guide for adding new CLI commands to kubectl-mtv. Use when creating new subcommands, adding verbs or resources, or wiring Cobra commands into the command tree.
---

# Adding a Cobra Command

kubectl-mtv uses a two-layer pattern: thin Cobra wiring in `cmd/` delegates to business logic in `pkg/cmd/`.

## Architecture

```
cmd/<verb>/<resource>.go        -> Cobra command (flags, args, RunE)
pkg/cmd/<verb>/<resource>/      -> Business logic (K8s calls, output)
```

The parent verb (e.g. `cmd/get/get.go`) registers subcommands. The root (`cmd/kubectl-mtv.go`) registers top-level verbs via `rootCmd.AddCommand()`.

## Flags Over Positional Args

Always prefer named flags (`--name my-plan`) over positional arguments. The MCP layer maps flags directly to JSON `flags` objects. Positional args require special handling and are harder for LLMs to discover.

Use `cobra.NoArgs` and define all parameters as flags:

```go
cmd := &cobra.Command{
    Use:  "host",
    Args: cobra.NoArgs,
    // ...
}
cmd.Flags().StringVarP(&hostName, "name", "M", "", "Host name")
```

## Step-by-Step Checklist

### 1. Add business logic

Create `pkg/cmd/<verb>/<resource>/list.go` (or appropriate name):

```go
package resource

import (
    "context"
    "fmt"
    "strings"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/cli-runtime/pkg/genericclioptions"

    "github.com/yaacov/kubectl-mtv/pkg/util/client"
    "github.com/yaacov/kubectl-mtv/pkg/util/output"
    "github.com/yaacov/kubectl-mtv/pkg/util/watch"
)

func ListItems(ctx context.Context, configFlags *genericclioptions.ConfigFlags, namespace, outputFormat string, useUTC bool) error {
    dynamicClient, err := client.GetDynamicClient(configFlags)
    if err != nil {
        return fmt.Errorf("failed to get client: %v", err)
    }

    items, err := dynamicClient.Resource(client.YourGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
    if err != nil {
        return fmt.Errorf("failed to list: %v", err)
    }

    allItems := make([]map[string]interface{}, 0, len(items.Items))
    for _, item := range items.Items {
        allItems = append(allItems, createItem(item, useUTC))
    }

    switch strings.ToLower(outputFormat) {
    case "json":
        return output.PrintJSONWithEmpty(allItems, "No items found.")
    case "yaml":
        return output.PrintYAMLWithEmpty(allItems, "No items found.")
    default:
        return printTable(allItems)
    }
}

func List(ctx context.Context, configFlags *genericclioptions.ConfigFlags, namespace string, watchMode bool, outputFormat string, useUTC bool) error {
    return watch.WrapWithWatch(watchMode, outputFormat, func() error {
        return ListItems(ctx, configFlags, namespace, outputFormat, useUTC)
    }, watch.DefaultInterval)
}
```

### 2. Add Cobra command

Create `cmd/<verb>/<resource>.go`:

```go
package verb

import (
    "context"
    "time"

    "github.com/spf13/cobra"
    "k8s.io/cli-runtime/pkg/genericclioptions"

    "github.com/yaacov/kubectl-mtv/pkg/cmd/<verb>/<resource>"
    "github.com/yaacov/kubectl-mtv/pkg/cmd/help"
    "github.com/yaacov/kubectl-mtv/pkg/util/client"
    "github.com/yaacov/kubectl-mtv/pkg/util/completion"
    "github.com/yaacov/kubectl-mtv/pkg/util/flags"
)

func NewResourceCmd(kubeConfigFlags *genericclioptions.ConfigFlags, globalConfig GlobalConfigGetter) *cobra.Command {
    outputFormatFlag := flags.NewOutputFormatTypeFlag()
    var watch bool
    var name string

    cmd := &cobra.Command{
        Use:          "resource",
        Short:        "Get resources",
        Long:         `Detailed description of the resource.`,
        Example: `  # List all resources
  kubectl-mtv get resources

  # Get a specific resource
  kubectl-mtv get resource --name my-resource --output yaml`,
        Args:         cobra.NoArgs,
        SilenceUsage: true,
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()
            if !watch {
                var cancel context.CancelFunc
                ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
                defer cancel()
            }

            namespace := client.ResolveNamespaceWithAllFlag(
                globalConfig.GetKubeConfigFlags(), globalConfig.GetAllNamespaces())

            return resource.List(ctx, globalConfig.GetKubeConfigFlags(),
                namespace, watch, outputFormatFlag.GetValue(), globalConfig.GetUseUTC())
        },
    }

    cmd.Flags().StringVarP(&name, "name", "N", "", "Resource name")
    cmd.Flags().VarP(outputFormatFlag, "output", "o", "Output format (table, json, yaml)")
    cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
    help.MarkMCPHidden(cmd, "watch")

    // Register completions
    _ = cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
        return outputFormatFlag.GetValidValues(), cobra.ShellCompDirectiveNoFileComp
    })

    return cmd
}
```

### 3. Register subcommand

In the parent verb file (e.g. `cmd/get/get.go`):

```go
resourceCmd := NewResourceCmd(kubeConfigFlags, globalConfig)
resourceCmd.Aliases = []string{"resources"}
cmd.AddCommand(resourceCmd)
```

### 4. Add GVR (if new CRD)

In `pkg/util/client/client.go`:

```go
YourGVR = schema.GroupVersionResource{
    Group:    Group,    // "forklift.konveyor.io"
    Version:  Version,  // "v1beta1"
    Resource: "yourresources",
}
```

### 5. Add completion (if name-based)

In `pkg/util/completion/completion.go`, add a function that lists resource names for shell completion.

### 6. Update category (if new top-level verb)

In `pkg/cmd/help/generator.go`, update `getCategory()`:

```go
case "get", "describe", "health", "yournewverb":
    return "read"
```

This controls whether the MCP layer exposes the command via `mtv_read` or `mtv_write`.

### 7. Verify MCP discovery

Commands under known verbs are auto-discovered via `help --machine`. Run `kubectl-mtv help --machine` and verify your command appears with correct category and flags.

## Key Packages

| Package | Purpose |
|---------|---------|
| `pkg/util/flags/` | Shared flag types (`OutputFormatTypeFlag`, `ProviderTypeFlag`) |
| `pkg/util/client/` | K8s dynamic client, GVR constants, namespace helpers |
| `pkg/util/output/` | Table, JSON, YAML printers |
| `pkg/util/watch/` | `WrapWithWatch` for `--watch` mode |
| `pkg/util/completion/` | Shell completion for flag values |
| `pkg/util/config/` | `GlobalConfigGetter` interface |
| `pkg/cmd/help/` | `MarkMCPHidden`, category assignment |

## Conventions

- Use `cobra.NoArgs` -- pass everything via flags.
- Add plural alias (`Aliases = []string{"resources"}`).
- Set `SilenceUsage: true` on all commands.
- Add 30s context timeout for non-watch commands.
- Use `help.MarkMCPHidden(cmd, "watch")` for interactive/TUI flags.
- Follow the `GlobalConfigGetter` interface for accessing global config.
