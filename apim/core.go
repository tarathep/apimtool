package apim

import (
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
)

type Operation struct {
	Method      string
	Name        string
	URLTemplate string
}

type Api struct {
	Name        string
	DisplayName string
	Protocols   []string
	Path        string
	BackendURL  string
}

type Backend struct {
	Name     string
	URL      string
	Protocol string
}

func safePointerString(s *string) string {
	if s == nil {
		temp := "" // *string cannot be initialized
		s = &temp  // in one statement
	}
	value := *s // safe to dereference the *string
	return value
}

func (a APIM) getOperationPolicy(resourceGroup, serviceName, apiID, operationID string) ([]string, error) {

	var operationPolicies []string

	apiOperationPolicyClient, err := armapimanagement.NewAPIOperationPolicyClient(a.SubscriptionID, a.Credential, nil)
	if err != nil {
		log.Println("failed to create client: %v", err)
		return nil, err
	}

	listOperation, err := apiOperationPolicyClient.ListByOperation(a.Context, resourceGroup, serviceName, apiID, operationID, &armapimanagement.APIOperationPolicyClientListByOperationOptions{})
	if err != nil {
		log.Println("failed to create client: %v", err)
		return nil, err
	}
	for _, v := range listOperation.Value {
		operationPolicies = append(operationPolicies, string(*v.Properties.Value))
	}
	return operationPolicies, nil
}

func (a APIM) getOperations(resourceGroup, serviceName, apiID, filter string) ([]Operation, error) {
	apiOperationClient, err := armapimanagement.NewAPIOperationClient(a.SubscriptionID, a.Credential, nil)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	pager := apiOperationClient.NewListByAPIPager(resourceGroup, serviceName, apiID, &armapimanagement.APIOperationClientListByAPIOptions{
		Filter: nil,
		Top:    nil,
		Skip:   nil,
		Tags:   nil,
	})

	var operations []Operation

	for pager.More() {
		nextResult, err := pager.NextPage(a.Context)
		if err != nil {
			log.Fatalf("failed to advance page: %v", err)
		}
		for _, v := range nextResult.Value {
			operations = append(operations, Operation{
				Method:      string(*v.Properties.Method),
				Name:        string(*v.Name),
				URLTemplate: string(*v.Properties.URLTemplate),
			})
		}
	}
	return operations, nil
}

func (a APIM) getAPIPolicy(resourceGroup, serviceName, apiID string) ([]string, error) {

	var apiPolicies []string

	apiOperationPolicyClient, err := armapimanagement.NewAPIPolicyClient(a.SubscriptionID, a.Credential, nil)
	if err != nil {
		log.Println("failed to create client: %v", err)
		return nil, err
	}

	apiPolicyClientListByAPIResponse, err := apiOperationPolicyClient.ListByAPI(a.Context, resourceGroup, serviceName, apiID, &armapimanagement.APIPolicyClientListByAPIOptions{})
	if err != nil {
		log.Println("failed to create client: %v", err)
		return nil, err
	}

	for _, ps := range apiPolicyClientListByAPIResponse.Value {
		apiPolicies = append(apiPolicies, safePointerString(ps.Properties.Value))
	}

	return apiPolicies, err
}

func (a APIM) getAPIs(resourceGroup, serviceName, filter string) ([]Api, error) {
	client, err := armapimanagement.NewAPIClient(a.SubscriptionID, a.Credential, nil)
	if err != nil {
		log.Println("failed to create client: %v", err)
		return []Api{}, err
	}
	pager := client.NewListByServicePager(resourceGroup,
		serviceName,
		&armapimanagement.APIClientListByServiceOptions{
			Filter:              to.Ptr("contains(properties/displayName, '" + filter + "')"),
			Top:                 nil,
			Skip:                nil,
			Tags:                nil,
			ExpandAPIVersionSet: nil,
		})

	apis := []Api{}

	for pager.More() {
		nextResult, err := pager.NextPage(a.Context)
		if err != nil {
			return []Api{}, err
			log.Fatalf("failed to advance page: %v", err)
		}

		for _, v := range nextResult.Value {
			_ = v

			apis = append(apis, Api{
				Name:        safePointerString(v.Name),
				DisplayName: safePointerString(v.Properties.DisplayName),
				Protocols: func() []string {
					var ps []string
					for _, p := range v.Properties.Protocols {
						ps = append(ps, string(*p))
					}
					return ps
				}(),
				Path:       safePointerString(v.Properties.Path),
				BackendURL: safePointerString(v.Properties.ServiceURL),
			})
		}
	}
	return apis, nil
}

func (a APIM) createOrUpdateBackend(resourceGroup, serviceName, backendID, url, protocol string) (armapimanagement.BackendClientCreateOrUpdateResponse, error) {
	client, err := armapimanagement.NewBackendClient(a.SubscriptionID, a.Credential, nil)
	if err != nil {
		log.Println("failed to create client: %v", err)
		return armapimanagement.BackendClientCreateOrUpdateResponse{}, err
	}

	return client.CreateOrUpdate(
		a.Context,
		resourceGroup,
		serviceName,
		backendID,
		armapimanagement.BackendContract{
			Properties: &armapimanagement.BackendContractProperties{
				Protocol: func() *armapimanagement.BackendProtocol {
					switch protocol {
					case "http":
						return &armapimanagement.PossibleBackendProtocolValues()[0]
					case "soap":
						return &armapimanagement.PossibleBackendProtocolValues()[1]
					default:
						return nil
					}
				}(),
				URL:         to.Ptr(url),
				Credentials: &armapimanagement.BackendCredentialsContract{},
				Description: nil,
				Properties:  &armapimanagement.BackendProperties{},
				Proxy:       nil,
				ResourceID:  nil,
				TLS:         &armapimanagement.BackendTLSProperties{ValidateCertificateName: to.Ptr(false), ValidateCertificateChain: to.Ptr(false)},
				Title:       nil,
			},
		},
		&armapimanagement.BackendClientCreateOrUpdateOptions{})
}

// get backend from APIM Filter pettern {key}={val}
func (a APIM) getBackends(resourceGroup, serviceName, filter string) ([]Backend, error) {
	client, err := armapimanagement.NewBackendClient(a.SubscriptionID, a.Credential, nil)
	if err != nil {
		log.Println("failed to create client: %v", err)
		return []Backend{}, err
	}

	filters := strings.Split(filter, "=")
	key := filters[0]
	if !(key == "url" || key == "name") {
		key = "name"
	}
	val := ""
	for i, raw := range filters {
		if len(filters) == 1 {
			val = raw
			break
		}
		if i != 0 {
			val += raw
		}
	}

	pager := client.NewListByServicePager(resourceGroup,
		serviceName,
		&armapimanagement.BackendClientListByServiceOptions{
			Filter: to.Ptr("contains(properties/" + key + ", '" + val + "')"),
			Top:    nil,
			Skip:   nil,
		})

	backends := []Backend{}

	for pager.More() {
		nextResult, err := pager.NextPage(a.Context)
		if err != nil {
			log.Fatalf("failed to advance page: %v", err)
		}

		for _, v := range nextResult.Value {
			// TODO: use page item
			_ = v

			//fmt.Println("Backend ID : ", *v.ID)
			backends = append(backends, Backend{
				Name:     string(*v.Name),
				URL:      string(*v.Properties.URL),
				Protocol: string(*v.Properties.Protocol),
			})
		}
	}
	return backends, err
}

func (a APIM) getAPIsBindingBackend(resourceGroup, serviceName, filter string) (
	[]struct {
		BE   Backend
		APIs []Api
	}, error) {

	var backends []struct {
		BE   Backend
		APIs []Api
	}

	getBackends, err := a.getBackends(resourceGroup, serviceName, filter)

	if err != nil {
		return []struct {
			BE   Backend
			APIs []Api
		}{}, err
	}
	for _, backend := range getBackends {
		backends = append(backends, struct {
			BE   Backend
			APIs []Api
		}{
			BE: backend,
			APIs: func() []Api {
				//find api this use backend ID or URL
				//a.getAPIs()

				return []Api{}
			}(),
		})

	}
	return []struct {
		BE   Backend
		APIs []Api
	}{}, nil
}
