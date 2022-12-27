package apim

import (
	"context"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
	"github.com/fatih/color"
)

type APIM struct {
	SubscriptionID string
	Location       string
	Credential     *azidentity.DefaultAzureCredential
	Context        context.Context
}

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
	Name string
	URL  string
}

func Env() struct {
	SubscriptionID string
	Location       string
} {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	subscriptionID = "24750e68-d6c2-40b7-90f9-f55b5009e909"
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}
	location = "southeastasia"

	return struct {
		SubscriptionID string
		Location       string
	}{SubscriptionID: subscriptionID, Location: location}
}

func (a APIM) getOperationPolicy(resourceGroup, serviceName string, apiID string, operationID string) ([]string, error) {

	var operationPolicies []string

	apiOperationPolicyClient, err := armapimanagement.NewAPIOperationPolicyClient(a.SubscriptionID, a.Credential, nil)
	if err != nil {
		log.Println("failed to create client: %v", err)
		return nil, err
	}

	listOperation, err := apiOperationPolicyClient.ListByOperation(a.Context, resourceGroupName, serviceName, apiID, operationID, &armapimanagement.APIOperationPolicyClientListByOperationOptions{})
	if err != nil {
		log.Println("failed to create client: %v", err)
		return nil, err
	}
	for _, v := range listOperation.Value {
		operationPolicies = append(operationPolicies, string(*v.Properties.Value))
	}
	return operationPolicies, nil
}

func (a APIM) getOperations(resourceGroup, serviceName string, apiID string, filter string) ([]Operation, error) {
	apiOperationClient, err := armapimanagement.NewAPIOperationClient(a.SubscriptionID, a.Credential, nil)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	pager := apiOperationClient.NewListByAPIPager(resourceGroupName, serviceName, apiID, &armapimanagement.APIOperationClientListByAPIOptions{
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

func (a APIM) getAPIPolicy(resourceGroup, serviceName string, apiID string) ([]string, error) {

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

func (a APIM) getAPIs(resourceGroup, serviceName string, filter string) ([]Api, error) {
	client, err := armapimanagement.NewAPIClient(a.SubscriptionID, a.Credential, nil)
	if err != nil {
		log.Println("failed to create client: %v", err)
		return []Api{}, err
	}
	pager := client.NewListByServicePager(resourceGroupName,
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

func (a APIM) getBackends(resourceGroup, serviceName string, filter string) ([]Backend, error) {
	client, err := armapimanagement.NewBackendClient(a.SubscriptionID, a.Credential, nil)
	if err != nil {
		log.Println("failed to create client: %v", err)
		return []Backend{}, err
	}

	pager := client.NewListByServicePager(resourceGroupName,
		serviceName,
		&armapimanagement.BackendClientListByServiceOptions{
			Filter: to.Ptr("contains(properties/url, '" + filter + "')"),
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
				Name: string(*v.Name),
				URL:  string(*v.Properties.URL),
			})
		}
	}
	return backends, err
}

func (apim APIM) ListBackend(resourceGroup, serviceName, filterDisplayName string, option string) {
	color.New(color.Italic).Print("List Backend's\n")

	backends, err := apim.getBackends(resourceGroup, serviceName, filterDisplayName)
	if err != nil {
		color.New(color.FgRed).Println("Fail to get APIs", err)
		return
	}

	var (
		maxNameSize, maxBackendURLSize int
	)
	if option == "" {
		option = "table"
	}

	switch option {
	case "table":
		{
			if len(backends) == 0 {
				color.New(color.FgHiBlue).Println("Not Found")
				return
			}
			// find max len values for print
			for _, backend := range backends {
				if len(backend.Name) > maxNameSize {
					maxNameSize = len(backend.Name)
				}

				if len(backend.URL) > maxBackendURLSize {
					maxBackendURLSize = len(backend.URL)
				}

			}
			if maxNameSize < 4 {
				maxNameSize = 4
			}

			if maxBackendURLSize < 10 {
				maxBackendURLSize = 10
			}

			color.New(color.FgHiMagenta).Printf("%*s  %*s  %*s\n", 3, "No.", maxNameSize, "NAME", maxBackendURLSize, "BackendURL")
			for i, backend := range backends {
				color.New(color.FgHiWhite).Printf("%*d  %*s  %*s\n", 3, (i + 1), maxNameSize, backend.Name, maxBackendURLSize, backend.URL)
			}
		}
	case "list":
		{
			for i, backend := range backends {
				color.New(color.FgHiBlack).Print("No : ")
				color.New(color.FgWhite).Println(1 + i)
				color.New(color.FgHiBlack).Print("BACKEND NAME : ")
				color.New(color.FgWhite).Println(backend.Name)
				color.New(color.FgHiBlack).Print("BACKEND URL : ")
				color.New(color.FgWhite).Println(backend.URL)
				color.New(color.FgHiWhite).Println("------------------------------------------------------------")
			}

		}
	}

}

func (apim APIM) ListAPI(resourceGroup, serviceName, filterDisplayName string, option string) {

	color.New(color.Italic).Print("List API Management API's\n")

	apis, err := apim.getAPIs(resourceGroup, serviceName, filterDisplayName)
	if err != nil {
		color.New(color.FgRed).Println("Fail to get APIs", err)
		return
	}

	var (
		maxApiNameSize, maxDisplayNameSize, maxProtocalSize, maxApiPathSize, maxApiBackendURLSize int
	)
	if option == "" {
		option = "table"
	}

	switch option {
	case "table":
		{
			if len(apis) == 0 {
				color.New(color.FgHiBlue).Println("Not Found")
				return
			}
			// find max len values for print
			for _, api := range apis {
				if len(api.Name) > maxApiNameSize {
					maxApiNameSize = len(api.Name)
				}

				if len(api.DisplayName) > maxDisplayNameSize {
					maxDisplayNameSize = len(api.DisplayName)
				}

				maxProtocalSizeT := func() int {
					ps := ""
					for _, p := range api.Protocols {
						ps += " " + p
					}
					return len(ps)
				}()

				if maxProtocalSizeT > maxProtocalSize {
					maxProtocalSize = maxProtocalSizeT
				}

				if len(api.Path) > maxApiPathSize {
					maxApiPathSize = len(api.Path)
				}

				if len(api.BackendURL) > maxApiBackendURLSize {
					maxApiBackendURLSize = len(api.BackendURL)
				}
			}
			if maxApiNameSize < 4 {
				maxApiNameSize = 4
			}
			if maxDisplayNameSize < 11 {
				maxDisplayNameSize = 11
			}
			if maxProtocalSize < 11 {
				maxProtocalSize = 11
			}
			if maxApiPathSize < 4 {
				maxApiPathSize = 4
			}
			if maxApiBackendURLSize < 10 {
				maxApiBackendURLSize = 10
			}

			color.New(color.FgHiMagenta).Printf("%*s  %*s  %*s  %*s  %*s  %*s\n", 3, "No.", maxApiNameSize, "NAME", maxDisplayNameSize, "DisplayName", maxProtocalSize, "Protocol(s)", maxApiPathSize, "Path", maxApiBackendURLSize, "BackendURL")
			for i, api := range apis {
				color.New(color.FgHiWhite).Printf("%*d  %*s  %*s  %*s  %*s  %*s\n", 3, (i + 1), maxApiNameSize, api.Name, maxDisplayNameSize, api.DisplayName, maxProtocalSize, func() string {
					ps := ""
					for _, p := range api.Protocols {
						ps += " " + p
					}
					return ps
				}(), maxApiPathSize, api.Path, maxApiBackendURLSize, api.BackendURL)
			}
		}
	case "list":
		{
			// bs, e := apim.getBackends(resourceGroup, serviceName, "")
			// if e != nil {
			// 	return
			// }
			for i, api := range apis {
				color.New(color.FgHiBlack).Print("No : ")
				color.New(color.FgWhite).Println(1 + i)
				color.New(color.FgHiBlack).Print("API NAME : ")
				color.New(color.FgWhite).Println(api.Name)
				color.New(color.FgHiBlack).Print("API DISPLAY NAME : ")
				color.New(color.FgWhite).Println(api.DisplayName)
				color.New(color.FgHiBlack).Print("PROTOCOL(s) : ")
				color.New(color.FgWhite).Println(api.Protocols)
				color.New(color.FgHiBlack).Print("PATH : ")
				color.New(color.FgWhite).Println(api.Path)
				color.New(color.FgHiBlack).Print("Backend URL : ")
				color.New(color.FgWhite).Println(api.BackendURL)

				//color.New(color.FgHiBlack).Print("Backend ID : ")

				// bes := func(bUrl string) bool {
				// 	for _, b := range bs {
				// 		fmt.Println(b.URL)
				// 		if b.URL == bUrl {
				// 			return true
				// 		}
				// 	}
				// 	return false
				// }(api.BackendURL)

				// color.New(color.FgHiYellow).Println(bes)

				color.New(color.FgHiBlack).Print("Oprations : \n")
				//apim.getOperations(resourceGroup, serviceName, api.Name, "")
				//apim.getOperationPolicy(resourceGroup, serviceName, api.Name, "")

				apim.getAPIPolicy(resourceGroup, serviceName, api.Name)
				color.New(color.FgHiWhite).Println("------------------------------------------------------------")

			}

		}
	}

}
func safePointerString(s *string) string {
	if s == nil {
		temp := "" // *string cannot be initialized
		s = &temp  // in one statement
	}
	value := *s // safe to dereference the *string
	return value
}
