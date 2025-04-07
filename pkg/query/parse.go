package query

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseQueryString parses a query string into its component parts
func ParseQueryString(query string) (*QueryOptions, error) {
	options := &QueryOptions{
		Limit: -1, // Default to no limit
	}

	if query == "" {
		return options, nil
	}

	// Convert query to lowercase for case-insensitive matching but preserve original for extraction
	queryLower := strings.ToLower(query)

	// Check for SELECT clause
	selectIndex := strings.Index(queryLower, "select ")
	whereIndex := strings.Index(queryLower, "where ")
	orderByIndex := strings.Index(queryLower, "order by ")
	limitIndex := strings.Index(queryLower, "limit ")

	// Extract SELECT clause if it exists
	if selectIndex >= 0 {
		selectEnd := len(query)
		if whereIndex > selectIndex {
			selectEnd = whereIndex
		} else if orderByIndex > selectIndex {
			selectEnd = orderByIndex
		} else if limitIndex > selectIndex {
			selectEnd = limitIndex
		}

		// Extract select clause (skip "select " prefix which is 7 chars)
		selectClause := strings.TrimSpace(query[selectIndex+7 : selectEnd])
		selectFields := strings.Split(selectClause, ",")

		for _, field := range selectFields {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}

			// Check if the field has an alias using "as" keyword
			parts := regexp.MustCompile(`(?i)\s+as\s+`).Split(field, 2)
			fieldPath := strings.TrimSpace(parts[0])
			alias := fieldPath

			if len(parts) > 1 {
				alias = strings.TrimSpace(parts[1])
			}

			// Make sure field is properly formatted for JSONPath
			if !strings.HasPrefix(fieldPath, ".") && !strings.HasPrefix(fieldPath, "{") {
				fieldPath = "." + fieldPath
			}

			options.Select = append(options.Select, SelectOption{
				Field: fieldPath,
				Alias: alias,
			})
		}

		options.HasSelect = len(options.Select) > 0
	}

	// Extract WHERE clause if it exists
	if whereIndex >= 0 {
		whereEnd := len(query)
		if orderByIndex > whereIndex {
			whereEnd = orderByIndex
		} else if limitIndex > whereIndex {
			whereEnd = limitIndex
		}

		// Extract where clause (skip "where " prefix which is 6 chars)
		options.Where = strings.TrimSpace(query[whereIndex+6 : whereEnd])
	}

	// Extract ORDER BY clause if it exists
	if orderByIndex >= 0 {
		orderByEnd := len(query)
		if limitIndex > orderByIndex {
			orderByEnd = limitIndex
		}

		// Extract order by clause (skip "order by " prefix which is 9 chars)
		orderByClause := strings.TrimSpace(query[orderByIndex+9 : orderByEnd])
		orderFields := strings.Split(orderByClause, ",")

		for _, field := range orderFields {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}

			descending := false
			// Check if field ends with DESC
			if strings.HasSuffix(strings.ToUpper(field), " DESC") {
				descending = true
				field = strings.TrimSuffix(field, " DESC")
				field = strings.TrimSuffix(field, " desc")
				field = strings.TrimSpace(field)
			} else {
				// Remove optional ASC suffix
				field = strings.TrimSuffix(field, " ASC")
				field = strings.TrimSuffix(field, " asc")
				field = strings.TrimSpace(field)
			}

			// Make sure field is properly formatted for JSONPath
			if !strings.HasPrefix(field, ".") && !strings.HasPrefix(field, "{") {
				field = "." + field
			}

			options.OrderBy = append(options.OrderBy, OrderOption{
				Field:      field,
				Descending: descending,
			})
		}

		options.HasOrderBy = len(options.OrderBy) > 0
	}

	// Extract LIMIT clause using regex (for simplicity with number extraction)
	limitRegex := regexp.MustCompile(`(?i)limit\s+(\d+)`)
	limitMatches := limitRegex.FindStringSubmatch(query)
	if len(limitMatches) > 1 {
		limit, err := strconv.Atoi(limitMatches[1])
		if err != nil {
			return nil, fmt.Errorf("invalid limit value: %v", err)
		}
		options.Limit = limit
		options.HasLimit = true
	}

	return options, nil
}
