package apim

import (
	"context"
	"fmt"
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

type Operation struct{}

type Api struct {
	Name        string
	DisplayName string
	Protocols   []string
	Path        string
	BackendURL  string
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
func (a APIM) getAPIs(resourceGroup, serviceName string, filter string) ([]Api, error) {
	client, err := armapimanagement.NewAPIClient(a.SubscriptionID, a.Credential, nil)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
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

func (apim APIM) ListBackend(resourceGroup, serviceName string) {
	client, err := armapimanagement.NewBackendClient(apim.SubscriptionID, apim.Credential, nil)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	pager := client.NewListByServicePager(resourceGroupName,
		serviceName,
		&armapimanagement.BackendClientListByServiceOptions{
			Filter: nil,
			Top:    nil,
			Skip:   nil,
		})

	for pager.More() {
		nextResult, err := pager.NextPage(apim.Context)
		if err != nil {
			log.Fatalf("failed to advance page: %v", err)
		}

		for _, v := range nextResult.Value {
			// TODO: use page item
			_ = v

			//fmt.Println("Backend ID : ", *v.ID)
			fmt.Println("Backend NAME : ", *v.Name)
			fmt.Println(*v.Properties.URL)

			fmt.Println("-------------------------------")
		}
	}
}

func (apim APIM) ListAPI(resourceGroup, serviceName, filterDisplayName string, mode byte) {

	color.New(color.Italic).Print("List API Management API's\n")

	apis, err := apim.getAPIs(resourceGroup, serviceName, filterDisplayName)
	if err != nil {
		color.New(color.FgRed).Println("Fail to get APIs", err)
		return
	}

	var (
		maxApiNameSize, maxDisplayNameSize, maxProtocalSize, maxApiPathSize, maxApiBackendURLSize int
	)

	switch mode {
	case 0:
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
	case 1:
		{
			for i, api := range apis {
				color.New(color.FgBlue).Print("No : ")
				color.New(color.FgWhite).Println(1 + i)
				color.New(color.FgCyan).Print("API NAME : ")
				color.New(color.FgWhite).Println(api.Name)
				color.New(color.FgGreen).Print("API DISPLAY NAME : ")
				color.New(color.FgWhite).Println(api.DisplayName)
				color.New(color.FgHiMagenta).Print("PROTOCOL(s) : ")
				color.New(color.FgWhite).Println(api.Protocols)
				color.New(color.FgHiRed).Print("PATH : ")
				color.New(color.FgWhite).Println(api.Path)
				color.New(color.FgHiBlack).Print("Backend URL : ")
				color.New(color.FgWhite).Println(api.BackendURL)
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
