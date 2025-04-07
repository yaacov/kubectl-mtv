package query

import (
	"fmt"

	"github.com/yaacov/tree-search-language/v6/pkg/tsl"
	"github.com/yaacov/tree-search-language/v6/pkg/walkers/semantics"
)

// FilterItems filters the items based on a WHERE clause using the tree-search-language
func FilterItems(items []map[string]interface{}, whereClause string) ([]map[string]interface{}, error) {
	// Parse the WHERE clause into a TSL tree
	tree, err := tsl.ParseTSL(whereClause)
	if err != nil {
		return nil, fmt.Errorf("failed to parse where clause: %v", err)
	}
	defer tree.Free()

	var results []map[string]interface{}

	// Filter the items collection using the TSL tree
	for _, item := range items {
		eval := evalFactory(item)

		matchingFilter, err := semantics.Walk(tree, eval)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate where clause: %v", err)
		}

		// Convert interface{} to bool
		if match, ok := matchingFilter.(bool); ok && match {
			results = append(results, item)
		}
	}

	return results, nil
}

// evalFactory gets an item and returns a method that will get the field and return its value
func evalFactory(item map[string]interface{}) semantics.EvalFunc {
	return func(k string) (interface{}, bool) {
		// First try direct access to the field
		if v, ok := item[k]; ok {
			return v, true
		}

		// If direct access fails, try to use JSONPath
		v, err := GetValueByPathString(item, "."+k)
		if err == nil && v != nil {
			return v, true
		}

		// If not found, don't return an error, it's ok to not find a field
		return nil, true
	}
}
