package flags

import (
	"fmt"
)

// MigrationTypeFlag implements pflag.Value interface for migration type validation
type MigrationTypeFlag struct {
	value string
}

func (m *MigrationTypeFlag) String() string {
	return m.value
}

func (m *MigrationTypeFlag) Set(value string) error {
	validTypes := []string{"cold", "warm", "live"}

	isValid := false
	for _, validType := range validTypes {
		if value == validType {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid migration type: %s. Valid types are: cold, warm, live", value)
	}

	m.value = value
	return nil
}

func (m *MigrationTypeFlag) Type() string {
	return "string"
}

// GetValue returns the migration type value
func (m *MigrationTypeFlag) GetValue() string {
	return m.value
}

// GetValidValues returns all valid migration type values for auto-completion
func (m *MigrationTypeFlag) GetValidValues() []string {
	return []string{"cold", "warm", "live"}
}

// NewMigrationTypeFlag creates a new migration type flag
func NewMigrationTypeFlag() *MigrationTypeFlag {
	return &MigrationTypeFlag{}
}
