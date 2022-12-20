package models

type BackendTemplate struct {
	Schema         string `json:"$schema"`
	ContentVersion string `json:"contentVersion"`
	Parameters     struct {
		ApimServiceName struct {
			Type string `json:"type"`
		} `json:"ApimServiceName"`
	} `json:"parameters"`
	Resources []struct {
		Properties struct {
			Credentials struct {
				Query struct {
				} `json:"query"`
				Header struct {
				} `json:"header"`
			} `json:"credentials"`
			TLS struct {
				ValidateCertificateChain bool `json:"validateCertificateChain"`
				ValidateCertificateName  bool `json:"validateCertificateName"`
			} `json:"tls"`
			URL      string `json:"url"`
			Protocol string `json:"protocol"`
		} `json:"properties"`
		Name       string `json:"name"`
		Type       string `json:"type"`
		APIVersion string `json:"apiVersion"`
	} `json:"resources"`
}
