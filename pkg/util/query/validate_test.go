package query

import (
	"strings"
	"testing"
)

func TestValidateQuerySyntax(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		expectError bool
		errorType   string
		errorMsg    string
	}{
		// Valid queries
		{
			name:        "empty query",
			query:       "",
			expectError: false,
		},
		{
			name:        "simple valid query",
			query:       "SELECT name WHERE id > 1 ORDER BY name LIMIT 10",
			expectError: false,
		},
		{
			name:        "valid query with sort by",
			query:       "WHERE status = 'active' SORT BY created_date",
			expectError: false,
		},

		// Typo detection tests
		{
			name:        "typo in SELECT",
			query:       "SELCT name FROM table",
			expectError: true,
			errorType:   "Keyword Typo",
			errorMsg:    "selct",
		},
		{
			name:        "typo in WHERE",
			query:       "SELECT name WHER id > 1",
			expectError: true,
			errorType:   "Keyword Typo",
			errorMsg:    "wher",
		},
		{
			name:        "typo in LIMIT",
			query:       "SELECT name LIMT 5",
			expectError: true,
			errorType:   "Keyword Typo",
			errorMsg:    "limt",
		},
		{
			name:        "typo in ORDER",
			query:       "ODER BY name",
			expectError: true,
			errorType:   "Keyword Typo",
			errorMsg:    "oder",
		},
		{
			name:        "typo in SORT",
			query:       "SROT BY name",
			expectError: true,
			errorType:   "Keyword Typo",
			errorMsg:    "srot",
		},

		// Duplicate keyword tests
		{
			name:        "duplicate SELECT",
			query:       "SELECT name SELECT age WHERE id > 1",
			expectError: true,
			errorType:   "Duplicate Keyword",
			errorMsg:    "SELECT",
		},
		{
			name:        "duplicate WHERE",
			query:       "SELECT name WHERE id > 1 WHERE status = 'active'",
			expectError: true,
			errorType:   "Duplicate Keyword",
			errorMsg:    "WHERE",
		},
		{
			name:        "duplicate LIMIT",
			query:       "SELECT name LIMIT 5 LIMIT 10",
			expectError: true,
			errorType:   "Duplicate Keyword",
			errorMsg:    "LIMIT",
		},
		{
			name:        "duplicate ORDER BY",
			query:       "SELECT name ORDER BY id ORDER BY name",
			expectError: true,
			errorType:   "Duplicate Keyword",
			errorMsg:    "ORDER BY",
		},

		// Clause ordering tests
		{
			name:        "WHERE before SELECT",
			query:       "WHERE id > 1 SELECT name",
			expectError: true,
			errorType:   "Invalid Clause Order",
			errorMsg:    "SELECT",
		},
		{
			name:        "LIMIT before WHERE",
			query:       "SELECT name LIMIT 5 WHERE id > 1",
			expectError: true,
			errorType:   "Invalid Clause Order",
			errorMsg:    "WHERE",
		},
		{
			name:        "ORDER BY before WHERE",
			query:       "SELECT name ORDER BY name WHERE id > 1",
			expectError: true,
			errorType:   "Invalid Clause Order",
			errorMsg:    "WHERE",
		},

		// Conflicting keywords tests
		{
			name:        "ORDER BY and SORT BY together",
			query:       "SELECT name ORDER BY id SORT BY name",
			expectError: true,
			errorType:   "Conflicting Keywords",
			errorMsg:    "ORDER BY",
		},

		// Empty clause tests
		{
			name:        "empty SELECT clause",
			query:       "SELECT WHERE id > 1",
			expectError: true,
			errorType:   "Empty Clause",
			errorMsg:    "SELECT",
		},
		{
			name:        "empty WHERE clause",
			query:       "SELECT name WHERE ORDER BY name",
			expectError: true,
			errorType:   "Empty Clause",
			errorMsg:    "WHERE",
		},
		{
			name:        "empty ORDER BY clause",
			query:       "SELECT name WHERE id > 1 ORDER BY LIMIT 10",
			expectError: true,
			errorType:   "Empty Clause",
			errorMsg:    "ORDER BY",
		},
		{
			name:        "empty LIMIT clause",
			query:       "SELECT name WHERE id > 1 LIMIT",
			expectError: true,
			errorType:   "Empty Clause",
			errorMsg:    "LIMIT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQuerySyntax(tt.query)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none for query: %s", tt.query)
					return
				}

				if !strings.Contains(err.Error(), tt.errorType) {
					t.Errorf("Expected error type '%s' but got: %s", tt.errorType, err.Error())
				}

				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s' but got: %s", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %s", err.Error())
				}
			}
		})
	}
}

func TestCheckTypos(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		expectError bool
		suggestion  string
	}{
		{
			name:        "no typos",
			query:       "SELECT name WHERE id > 1",
			expectError: false,
		},
		{
			name:        "selct typo",
			query:       "SELCT name",
			expectError: true,
			suggestion:  "SELECT",
		},
		{
			name:        "limt typo",
			query:       "LIMT 5",
			expectError: true,
			suggestion:  "LIMIT",
		},
		{
			name:        "wher typo",
			query:       "WHER id = 1",
			expectError: true,
			suggestion:  "WHERE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkTypos(tt.query)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if !strings.Contains(err.Error(), tt.suggestion) {
					t.Errorf("Expected suggestion '%s' in error: %s", tt.suggestion, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %s", err.Error())
				}
			}
		})
	}
}

func TestCountKeywordOccurrences(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected map[string]int
	}{
		{
			name:  "single keywords",
			query: "SELECT name WHERE id > 1 ORDER BY name LIMIT 10",
			expected: map[string]int{
				"SELECT":   1,
				"WHERE":    1,
				"ORDER BY": 1,
				"LIMIT":    1,
			},
		},
		{
			name:  "duplicate keywords",
			query: "SELECT name SELECT age WHERE id > 1",
			expected: map[string]int{
				"SELECT": 2,
				"WHERE":  1,
			},
		},
		{
			name:  "case insensitive",
			query: "select name where ID > 1 order by name",
			expected: map[string]int{
				"SELECT":   1,
				"WHERE":    1,
				"ORDER BY": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			occurrences := countKeywordOccurrences(tt.query)

			// Convert to map for easier comparison
			actual := make(map[string]int)
			for _, occ := range occurrences {
				actual[occ.Keyword.Name] = occ.Count
			}

			// Check each expected keyword
			for keyword, expectedCount := range tt.expected {
				if actualCount, found := actual[keyword]; !found {
					t.Errorf("Expected keyword '%s' not found", keyword)
				} else if actualCount != expectedCount {
					t.Errorf("Keyword '%s': expected count %d, got %d", keyword, expectedCount, actualCount)
				}
			}

			// Check no unexpected keywords
			for keyword := range actual {
				if _, expected := tt.expected[keyword]; !expected {
					t.Errorf("Unexpected keyword '%s' found with count %d", keyword, actual[keyword])
				}
			}
		})
	}
}

func TestParseQueryStringWithValidation(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		expectError bool
		errorType   string
	}{
		{
			name:        "valid query passes validation",
			query:       "SELECT name WHERE id > 1 ORDER BY name LIMIT 10",
			expectError: false,
		},
		{
			name:        "typo fails validation",
			query:       "SELCT name WHERE id > 1",
			expectError: true,
			errorType:   "Keyword Typo",
		},
		{
			name:        "duplicate keyword fails validation",
			query:       "SELECT name SELECT age WHERE id > 1",
			expectError: true,
			errorType:   "Duplicate Keyword",
		},
		{
			name:        "wrong order fails validation",
			query:       "WHERE id > 1 SELECT name",
			expectError: true,
			errorType:   "Invalid Clause Order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseQueryString(tt.query)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				if !strings.Contains(err.Error(), tt.errorType) {
					t.Errorf("Expected error type '%s' but got: %s", tt.errorType, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %s", err.Error())
				}
			}
		})
	}
}
