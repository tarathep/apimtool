package apim

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/fatih/color"
	"github.com/tarathep/apimtool/models"
)

type APIM struct {
	SubscriptionID string
	Location       string
	Credential     *azidentity.DefaultAzureCredential
	Context        context.Context
}

func Env() struct {
	SubscriptionID string
	Location       string
} {
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	subscriptionID = "24750e68-d6c2-40b7-90f9-f55b5009e909"
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}
	location := "southeastasia"

	return struct {
		SubscriptionID string
		Location       string
	}{SubscriptionID: subscriptionID, Location: location}
}

func (apim APIM) ListBackend(resourceGroup, serviceName, filterDisplayName, option string) {
	color.New(color.Italic).Print("List Backend's\n\n")

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

	color.New(color.Italic).Print("List API Management API's\n\n")

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

				color.New(color.FgHiBlack).Print("Backend Policy ID : ")
				backendPolicyID := apim.GetAPIPolicy(resourceGroup, serviceName, api.Name).Inbound.SetBackendService.BackendID
				color.New(color.FgWhite).Println(backendPolicyID)

				color.New(color.FgHiBlack).Print("Backend Policy URL : ")

				backendPolicyURL, err := apim.GetBackendURLfromID(resourceGroup, serviceName, backendPolicyID)
				if err != nil {
					color.New(color.FgHiRed).Println("Error", err)
					return
				}
				color.New(color.FgWhite).Println(backendPolicyURL)

				color.New(color.FgHiBlack).Print("Operations : \n")
				operations, err := apim.getOperations(resourceGroup, serviceName, api.Name, "")
				if err != nil {
					color.New(color.FgHiRed).Println("Error", err)
					return
				}
				for i, operation := range operations {
					color.New(color.FgHiWhite).Print("  ", (i + 1), " ")
					switch operation.Method {
					case "GET":
						color.New(color.FgHiBlue).Print(operation.Method, " ")
					case "POST":
						color.New(color.FgHiGreen).Print(operation.Method, " ")
					case "PUT":
						color.New(color.FgHiYellow).Print(operation.Method, " ")
					case "DELETE":
						color.New(color.FgHiRed).Print(operation.Method, " ")
					case "PATCH":
						color.New(color.FgHiCyan).Print(operation.Method, " ")
					default:
						color.New(color.FgHiBlack).Print(operation.Method, " ")
					}
					color.New(color.FgHiWhite).Println(operation.Name, operation.URLTemplate)

				}

				//apim.getOperationPolicy(resourceGroup, serviceName, api.Name, "")

				color.New(color.FgHiWhite).Println("------------------------------------------------------------")

			}

		}
	}
	urlx, err := apim.GetBackendIDfromURL(resourceGroup, serviceName, "https://app-tarathec-az-usw3-dev-001.azurewebsites.net")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(urlx)

}

func (a APIM) GetBackendURLfromID(resourceGroup, serviceName, backendID string) (string, error) {
	if backendID == "" {
		return "", nil
	}

	backends, err := a.getBackends(resourceGroup, serviceName, "name="+backendID)
	if err != nil {
		return "", err
	}
	URLs := "["
	for i, backend := range backends {
		if len(backends) == 1 {
			return backend.URL, nil
		}
		if (i + 1) == len(backends) {
			URLs += backend.URL + "]"
		} else {
			URLs += backend.URL + ","
		}
	}
	return URLs, nil

}

func (a APIM) GetBackendIDfromURL(resourceGroup, serviceName, url string) (string, error) {

	BackendIDs := "["
	backends, err := a.getBackends(resourceGroup, serviceName, url)

	if err != nil {
		return "err", err
	}
	for i, backend := range backends {
		if len(backends) == 1 {
			return backend.Name, nil
		}
		if (i + 1) == len(backends) {
			BackendIDs += backend.Name + "]"
		} else {
			BackendIDs += backend.Name + ","
		}
	}
	return "", nil

}

func (a APIM) GetAPIPolicy(resourceGroup, serviceName, apiID string) models.Policies {
	apiPoliciesHeader := models.Policies{}
	apiPolicy, err := a.getAPIPolicy(resourceGroup, serviceName, apiID)
	if err != nil {
		return models.Policies{}
	}
	xml.Unmarshal([]byte(apiPolicy[0]), &apiPoliciesHeader)
	return apiPoliciesHeader
}
