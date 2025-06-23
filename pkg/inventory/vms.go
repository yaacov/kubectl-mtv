package inventory

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	planv1beta1 "github.com/konveyor/forklift-controller/pkg/apis/forklift/v1beta1/plan"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/output"
	querypkg "github.com/yaacov/kubectl-mtv/pkg/query"
	"github.com/yaacov/kubectl-mtv/pkg/watch"
)

// countConcernsByCategory counts VM concerns by their category
func countConcernsByCategory(vm map[string]interface{}) map[string]int {
	counts := map[string]int{
		"Critical":    0,
		"Warning":     0,
		"Information": 0,
	}

	concerns, exists := vm["concerns"]
	if !exists {
		return counts
	}

	concernsArray, ok := concerns.([]interface{})
	if !ok {
		return counts
	}

	for _, c := range concernsArray {
		concern, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		category, ok := concern["category"].(string)
		if ok {
			counts[category]++
		}
	}

	return counts
}

// formatVMConcerns formats all concerns for a VM into a displayable string
func formatVMConcerns(vm map[string]interface{}) string {
	concerns, exists := vm["concerns"]
	if !exists {
		return "No concerns found"
	}

	concernsArray, ok := concerns.([]interface{})
	if !ok || len(concernsArray) == 0 {
		return "No concerns found"
	}

	var result strings.Builder

	for i, c := range concernsArray {
		concern, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		if i > 0 {
			result.WriteString("\n")
		}

		// Get category and use short form
		categoryShort := "[?]" // Default if category unknown
		if category, ok := concern["category"].(string); ok {
			switch category {
			case "Critical":
				categoryShort = "[C]"
			case "Warning":
				categoryShort = "[W]"
			case "Information":
				categoryShort = "[I]"
			default:
				categoryShort = "[" + string(category[0]) + "]"
			}
		}
		result.WriteString(categoryShort + " ")

		// Add assessment
		if assessment, ok := concern["assessment"].(string); ok {
			result.WriteString(assessment)
		} else {
			result.WriteString("No details available")
		}
	}

	return result.String()
}

// calculateTotalDiskCapacity returns the total disk capacity in GB
func calculateTotalDiskCapacity(vm map[string]interface{}) float64 {
	disks, exists := vm["disks"]
	if !exists {
		return 0
	}

	disksArray, ok := disks.([]interface{})
	if !ok {
		return 0
	}

	var totalCapacity float64
	for _, d := range disksArray {
		disk, ok := d.(map[string]interface{})
		if !ok {
			continue
		}

		if capacity, ok := disk["capacity"].(float64); ok {
			totalCapacity += capacity
		}
	}

	// Convert to GB (from bytes)
	return totalCapacity / (1024 * 1024 * 1024)
}

// ListVMs queries the provider's VM inventory and displays the results
func ListVMs(kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, extendedOutput bool, query string, watchMode bool) error {
	if watchMode {
		return watch.Watch(func() error {
			return listVMsOnce(kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, extendedOutput, query)
		}, 10*time.Second)
	}

	return listVMsOnce(kubeConfigFlags, providerName, namespace, inventoryURL, outputFormat, extendedOutput, query)
}

func listVMsOnce(kubeConfigFlags *genericclioptions.ConfigFlags, providerName, namespace string, inventoryURL string, outputFormat string, extendedOutput bool, query string) error {
	// Get the provider object
	provider, err := GetProviderByName(kubeConfigFlags, providerName, namespace)
	if err != nil {
		return err
	}

	// Fetch VM inventory from the provider
	data, err := client.FetchProviderInventory(kubeConfigFlags, inventoryURL, provider, "vms?detail=4")
	if err != nil {
		return fmt.Errorf("failed to fetch VM inventory: %v", err)
	}

	// Verify data is an array
	dataArray, ok := data.([]interface{})
	if !ok {
		return fmt.Errorf("unexpected data format: expected array for VM inventory")
	}

	// Convert to expected format
	vms := make([]map[string]interface{}, 0, len(dataArray))
	expandedData := make(map[string]string) // Map for expanded VM concerns

	for _, item := range dataArray {
		if vm, ok := item.(map[string]interface{}); ok {
			// Add provider name to each VM
			vm["provider"] = providerName

			// Format VM name for expanded data key
			vmName, _ := vm["name"].(string)

			// Add concern counts by category
			concernCounts := countConcernsByCategory(vm)
			vm["criticalConcerns"] = concernCounts["Critical"]
			vm["warningConcerns"] = concernCounts["Warning"]
			vm["infoConcerns"] = concernCounts["Information"]

			// Create a combined concerns string (Critical/Warning/Info)
			vm["concernsHuman"] = fmt.Sprintf("%d/%d/%d",
				concernCounts["Critical"],
				concernCounts["Warning"],
				concernCounts["Information"])

			// Add (*) indicator if critical concerns exist
			if concernCounts["Critical"] > 0 {
				vm["concernsHuman"] = vm["concernsHuman"].(string) + " (*)"
			}

			// If VM has concerns, create expanded data
			if extendedOutput && (concernCounts["Critical"] > 0 || concernCounts["Warning"] > 0 || concernCounts["Information"] > 0) {
				// Format concerns for expanded view
				expandedData[vmName] = formatVMConcerns(vm)
			}

			// Format memory in GB for display
			if memoryMB, exists := vm["memoryMB"]; exists {
				if memVal, ok := memoryMB.(float64); ok {
					vm["memoryGB"] = fmt.Sprintf("%.1f GB", memVal/1024)
				}
			}

			// Calculate and format disk capacity
			totalDiskCapacityGB := calculateTotalDiskCapacity(vm)
			vm["diskCapacity"] = fmt.Sprintf("%.1f GB", totalDiskCapacityGB)

			// Format storage used
			if storageUsed, exists := vm["storageUsed"]; exists {
				if storageVal, ok := storageUsed.(float64); ok {
					storageUsedGB := storageVal / (1024 * 1024 * 1024)
					vm["storageUsedGB"] = fmt.Sprintf("%.1f GB", storageUsedGB)
				}
			}

			// Humanize power state
			if powerState, exists := vm["powerState"]; exists {
				if ps, ok := powerState.(string); ok {
					if strings.Contains(strings.ToLower(ps), "on") {
						vm["powerStateHuman"] = "On"
					} else {
						vm["powerStateHuman"] = "Off"
					}
				}
			}

			vms = append(vms, vm)
		}
	}

	// Parse and apply query options
	queryOpts, err := querypkg.ParseQueryString(query)
	if err != nil {
		return fmt.Errorf("invalid query string: %v", err)
	}

	// Apply query options (sorting, filtering, limiting)
	vms, err = querypkg.ApplyQuery(vms, queryOpts)
	if err != nil {
		return fmt.Errorf("error applying query: %v", err)
	}

	// Format validation
	outputFormat = strings.ToLower(outputFormat)
	if outputFormat != "table" && outputFormat != "json" && outputFormat != "planvms" && outputFormat != "yaml" {
		return fmt.Errorf("unsupported output format: %s. Supported formats: table, json, yaml, planvms", outputFormat)
	}

	// Handle different output formats
	if outputFormat == "json" {
		// Use JSON printer
		jsonPrinter := output.NewJSONPrinter().
			WithPrettyPrint(true).
			AddItems(vms)

		if len(vms) == 0 {
			return jsonPrinter.PrintEmpty(fmt.Sprintf("No VMs found for provider %s", providerName))
		}
		return jsonPrinter.Print()
	} else if outputFormat == "yaml" {
		// Use YAML printer
		yamlPrinter := output.NewYAMLPrinter().
			AddItems(vms)

		if len(vms) == 0 {
			return yamlPrinter.PrintEmpty(fmt.Sprintf("No VMs found for provider %s", providerName))
		}
		return yamlPrinter.Print()
	} else if outputFormat == "planvms" {
		// Convert inventory VMs to plan VM structs
		planVMs := make([]planv1beta1.VM, 0, len(vms))
		for _, vm := range vms {
			vmName, ok := vm["name"].(string)
			if !ok {
				continue
			}

			planVM := planv1beta1.VM{}
			planVM.Name = vmName

			// Add ID if available
			if vmID, ok := vm["id"].(string); ok {
				planVM.ID = vmID
			}

			planVMs = append(planVMs, planVM)
		}

		// Marshal to YAML
		yamlData, err := yaml.Marshal(planVMs)
		if err != nil {
			return fmt.Errorf("failed to marshal plan VMs to YAML: %v", err)
		}

		// Print the YAML to stdout
		fmt.Println(string(yamlData))
		return nil
	} else {
		var tablePrinter *output.TablePrinter

		// Check if we should use custom headers from SELECT clause
		if queryOpts.HasSelect {
			headers := make([]output.Header, 0, len(queryOpts.Select))
			for _, sel := range queryOpts.Select {
				headers = append(headers, output.Header{
					DisplayName: sel.Alias,
					JSONPath:    sel.Alias,
				})
			}
			tablePrinter = output.NewTablePrinter().
				WithHeaders(headers...).
				WithSelectOptions(queryOpts.Select)
		} else {
			// Use default table headers
			tablePrinter = output.NewTablePrinter().WithHeaders(
				output.Header{DisplayName: "NAME", JSONPath: "name"},
				output.Header{DisplayName: "ID", JSONPath: "id"},
				output.Header{DisplayName: "POWER", JSONPath: "powerStateHuman"},
				output.Header{DisplayName: "CPU", JSONPath: "cpuCount"},
				output.Header{DisplayName: "MEMORY", JSONPath: "memoryGB"},
				output.Header{DisplayName: "DISK USAGE", JSONPath: "storageUsedGB"},
				output.Header{DisplayName: "GUEST OS", JSONPath: "guestId"},
				output.Header{DisplayName: "CONCERNS (C/W/I)", JSONPath: "concernsHuman"},
			)
		}

		// Add items with expanded concern data
		for _, vm := range vms {
			vmName, _ := vm["name"].(string)
			expandedText, hasExpanded := expandedData[vmName]

			if hasExpanded {
				tablePrinter.AddItemWithExpanded(vm, expandedText)
			} else {
				tablePrinter.AddItem(vm)
			}
		}

		if len(vms) == 0 {
			return tablePrinter.PrintEmpty(fmt.Sprintf("No VMs found for provider %s", providerName))
		}
		return tablePrinter.Print()
	}
}
