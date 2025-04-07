package providerutil

// ProviderOptions contains the options for creating a provider
type ProviderOptions struct {
	Name            string
	Namespace       string
	Secret          string
	URL             string
	Username        string
	Password        string
	CACert          string
	InsecureSkipTLS bool
	VddkInitImage   string
	Token           string
}
