package mtvmcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateNetworkPairs(t *testing.T) {
	tests := []struct {
		name        string
		pairsStr    string
		shouldError bool
		errorTarget string
		description string
	}{
		{
			name:        "empty string is valid",
			pairsStr:    "",
			shouldError: false,
			description: "Empty pairs should be allowed",
		},
		{
			name:        "single pair is valid",
			pairsStr:    "source1:target1",
			shouldError: false,
			description: "Single pair should be valid",
		},
		{
			name:        "multiple pairs with different targets is valid",
			pairsStr:    "source1:target1,source2:target2,source3:target3",
			shouldError: false,
			description: "Multiple pairs with different targets should be valid",
		},
		{
			name:        "default with ignored is valid",
			pairsStr:    "source1:default,source2:ignored,source3:nad1",
			shouldError: false,
			description: "Default with ignored and NAD should be valid",
		},
		{
			name:        "multiple ignored is valid",
			pairsStr:    "source1:ignored,source2:ignored,source3:ignored,source4:nad1",
			shouldError: false,
			description: "Multiple ignored targets should be allowed",
		},
		{
			name:        "duplicate default is invalid",
			pairsStr:    "source1:default,source2:default",
			shouldError: true,
			errorTarget: "default",
			description: "Duplicate default (pod networking) should fail",
		},
		{
			name:        "duplicate NAD is invalid",
			pairsStr:    "source1:nad1,source2:nad1",
			shouldError: true,
			errorTarget: "nad1",
			description: "Duplicate NAD should fail",
		},
		{
			name:        "duplicate NAD with namespace is invalid",
			pairsStr:    "source1:ns1/nad1,source2:ns1/nad1",
			shouldError: true,
			errorTarget: "ns1/nad1",
			description: "Duplicate NAD with namespace should fail",
		},
		{
			name:        "three sources to same target is invalid",
			pairsStr:    "source1:target1,source2:target1,source3:target1",
			shouldError: true,
			errorTarget: "target1",
			description: "Three sources to same target should fail",
		},
		{
			name:        "mixed valid and invalid",
			pairsStr:    "source1:nad1,source2:nad2,source3:nad1",
			shouldError: true,
			errorTarget: "nad1",
			description: "Should catch duplicate even with other valid pairs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNetworkPairs(tt.pairsStr)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tt.description)
					return
				}

				// Check if error contains expected target
				if !strings.Contains(err.Error(), tt.errorTarget) {
					t.Errorf("Expected error to mention target '%s', but got: %s", tt.errorTarget, err.Error())
				}

				// Verify error is valid JSON
				var errData map[string]interface{}
				if jsonErr := json.Unmarshal([]byte(err.Error()), &errData); jsonErr != nil {
					t.Errorf("Error should be valid JSON, but got: %s", err.Error())
				}

				// Verify required fields in error
				if errData["error"] != "validation_error" {
					t.Errorf("Expected error field to be 'validation_error', got: %v", errData["error"])
				}
				if errData["type"] != "duplicate_network_target" {
					t.Errorf("Expected type field to be 'duplicate_network_target', got: %v", errData["type"])
				}
				if errData["target"] != tt.errorTarget {
					t.Errorf("Expected target field to be '%s', got: %v", tt.errorTarget, errData["target"])
				}

			} else {
				if err != nil {
					t.Errorf("Expected no error for %s, but got: %v", tt.description, err)
				}
			}
		})
	}
}

func TestValidateNetworkPairsWhitespace(t *testing.T) {
	// Test that whitespace is handled correctly
	tests := []struct {
		name        string
		pairsStr    string
		shouldError bool
	}{
		{
			name:        "whitespace around pairs",
			pairsStr:    " source1:target1 , source2:target2 ",
			shouldError: false,
		},
		{
			name:        "whitespace in pair",
			pairsStr:    "source1 : target1 , source2 : target2",
			shouldError: false,
		},
		{
			name:        "whitespace with duplicates",
			pairsStr:    "source1:default , source2:default",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNetworkPairs(tt.pairsStr)
			if tt.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
