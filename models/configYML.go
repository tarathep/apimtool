package models

type ConfigYML struct {
	Version         string      `yaml:"version"`
	ApimServiceName string      `yaml:"apimServiceName"`
	Apis            []APIConfig `yaml:"apis"`
	OutputLocation  string      `yaml:"outputLocation"`
}

type APIConfig struct {
	Name                   string `yaml:"name"`
	OpenAPISpec            string `yaml:"openApiSpec"`
	Policy                 string `yaml:"policy"`
	Suffix                 string `yaml:"suffix"`
	Protocols              string `yaml:"protocols"`
	Revision               int    `yaml:"revision"`
	AuthenticationSettings struct {
		SubscriptionKeyRequired bool `yaml:"subscriptionKeyRequired"`
	} `yaml:"authenticationSettings"`
	SubscriptionKeyParameterNames struct {
		Header string `yaml:"header"`
		Query  string `yaml:"query"`
	} `yaml:"subscriptionKeyParameterNames"`
	Tags []string `yaml:"tags"`
}
