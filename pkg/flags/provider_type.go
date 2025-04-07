package flags

import (
	"fmt"

	forkliftv1beta1 "github.com/konveyor/forklift-controller/pkg/apis/forklift/v1beta1"
)

// ProviderTypeFlag implements pflag.Value interface for provider type validation
type ProviderTypeFlag struct {
	value string
}

func (p *ProviderTypeFlag) String() string {
	return p.value
}

func (p *ProviderTypeFlag) Set(value string) error {
	validTypes := []forkliftv1beta1.ProviderType{
		forkliftv1beta1.OpenShift,
		forkliftv1beta1.VSphere,
		forkliftv1beta1.OVirt,
		forkliftv1beta1.OpenStack,
		forkliftv1beta1.Ova,
	}

	isValid := false
	for _, validType := range validTypes {
		if forkliftv1beta1.ProviderType(value) == validType {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid provider type: %s. Valid types are: openshift, vsphere, ovirt, openstack, ova", value)
	}

	p.value = value
	return nil
}

func (p *ProviderTypeFlag) Type() string {
	return "string"
}

// GetValue returns the provider type value
func (p *ProviderTypeFlag) GetValue() string {
	return p.value
}

// NewProviderTypeFlag creates a new provider type flag
func NewProviderTypeFlag() *ProviderTypeFlag {
	return &ProviderTypeFlag{}
}
