package apim

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"time"

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

type apiModel struct {
	No               int
	APIName          string
	APIDisplayName   string
	APIProtocols     []string
	APIPath          string
	APIBackendURL    string
	BackendPolicyID  string
	BackendPolicyURL string
	Operation        []Operation
}

func Env() struct {
	SubscriptionID string
	Location       string
} {
	subscriptionID := os.Getenv("APIMTOOL_AZURE_SUBSCRIPTION_ID")

	if len(subscriptionID) == 0 {
		log.Fatal("APIMTOOL_AZURE_SUBSCRIPTION_ID is not set.")
	}
	location := os.Getenv("APIMTOOL_AZURE_LOCATION")

	if len(location) == 0 {
		log.Fatal("APIMTOOL_AZURE_LOCATION is not set.")
	}

	return struct {
		SubscriptionID string
		Location       string
	}{SubscriptionID: subscriptionID, Location: location}
}

func (apim APIM) ListBackend(resourceGroup, serviceName, filterDisplayName, option string) {
	color.New(color.Italic, color.FgHiBlue, color.Bold).Print("List Backend's\n\n")

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
				fmt.Println(1 + i)
				color.New(color.FgHiBlack).Print("BACKEND NAME : ")
				fmt.Println(backend.Name)
				color.New(color.FgHiBlack).Print("BACKEND URL : ")
				fmt.Println(backend.URL)
				color.New(color.FgHiBlack).Print("BACKEND Protocol : ")
				fmt.Println(backend.Protocol)
				color.New(color.FgHiWhite).Println("------------------------------------------------------------")
			}

		}
	}
}

func (apim APIM) ListAPI(resourceGroup, serviceName, filterDisplayName string, option string) {
	start := time.Now()
	color.New(color.Italic, color.FgHiBlue, color.Bold).Print("List API Management API's\n\n")

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
			err, apiModels := apim.apis(resourceGroup, serviceName, filterDisplayName)
			if err != nil {
				return
			}
			// Sort by No, keeping original order or equal elements.
			sort.SliceStable(apiModels, func(i, j int) bool {
				return apiModels[i].No < apiModels[j].No
			})

			for i, model := range apiModels {
				color.New(color.FgHiBlack).Print("No : ")
				fmt.Println(1 + i)
				color.New(color.FgHiBlack).Print("API NAME : ")
				fmt.Println(model.APIName)
				color.New(color.FgHiBlack).Print("API DISPLAY NAME : ")
				fmt.Println(model.APIDisplayName)
				color.New(color.FgHiBlack).Print("PROTOCOL(s) : ")
				fmt.Println(model.APIProtocols)
				color.New(color.FgHiBlack).Print("PATH : ")
				fmt.Println(model.APIPath)
				color.New(color.FgHiBlack).Print("Backend URL : ")
				fmt.Println(model.APIBackendURL)

				color.New(color.FgHiBlack).Print("Backend Policy ID : ")

				fmt.Println(model.BackendPolicyID)

				color.New(color.FgHiBlack).Print("Backend Policy URL : ")

				fmt.Println(model.BackendPolicyURL)

				color.New(color.FgHiBlack).Print("Operations : \n")
				operations := model.Operation

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
				color.New(color.FgHiWhite).Println("------------------------------------------------------------")
			}

		}
	}
	fmt.Println("\nTime used is ", time.Since(start))
}

func (a APIM) apis(resourceGroup, serviceName, filterDisplayName string) (error, []apiModel) {
	apis, err := a.getAPIs(resourceGroup, serviceName, filterDisplayName)
	if err != nil {
		color.New(color.FgRed).Println("Fail to get APIs", err)
		return err, []apiModel{}
	}

	var apiModels []apiModel

	// Create a WaitGroup to synchronize the Go routines
	var wg sync.WaitGroup
	wg.Add(len(apis))

	for _, api := range apis {
		go func(api Api) {
			defer wg.Done()

			apiModel := func(api Api) apiModel {
				var model apiModel
				model.No = api.No
				model.APIName = api.Name
				model.APIDisplayName = api.DisplayName
				model.APIProtocols = api.Protocols
				model.APIPath = api.Path
				model.APIBackendURL = api.BackendURL

				//Create a Channel for GetAPIPolicy
				c1 := make(chan string)

				go func(a APIM, resourceGroup string, serviceName string, Name string) {
					c1 <- a.GetAPIPolicy(resourceGroup, serviceName, Name).Inbound.SetBackendService.BackendID
				}(a, resourceGroup, serviceName, api.Name)

				backendPolicyID := <-c1
				model.BackendPolicyID = backendPolicyID

				//Create a Channel for GetBackendURLfromID
				c2 := make(chan string)
				go func(a APIM, resourceGroup string, serviceName string, backendPolicyID string) {
					backendPolicyURL, err := a.GetBackendURLfromID(resourceGroup, serviceName, backendPolicyID)
					if err != nil {
						color.New(color.FgHiRed).Println("Error", err)
					}
					c2 <- backendPolicyURL
				}(a, resourceGroup, serviceName, backendPolicyID)

				//Create a Channel for Operation
				model.BackendPolicyURL = <-c2

				c3 := make(chan []Operation)
				go func(a APIM, resourceGroup string, serviceName string, Name string) {
					operation, err := a.getOperations(resourceGroup, serviceName, api.Name, "")
					if err != nil {
						color.New(color.FgHiRed).Println("Error", err)
					}
					c3 <- operation
				}(a, resourceGroup, serviceName, api.Name)

				model.Operation = <-c3

				return model
			}(api)

			apiModels = append(apiModels, apiModel)
		}(api)

	}
	wg.Wait()
	return nil, apiModels
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

// Get Backend ID by URL from APIM
func (a APIM) GetBackendIDfromURL(resourceGroup, serviceName, url string) (string, error) {
	BackendIDs := ""
	backends, err := a.getBackends(resourceGroup, serviceName, "url="+url)

	if err != nil {
		return "", err
	}
	for i, backend := range backends {
		if len(backends) == 1 {
			return backend.Name, nil
		}
		if (i + 1) == len(backends) {
			BackendIDs += backend.Name
		} else {
			BackendIDs += backend.Name + ","
		}
	}
	return BackendIDs, nil

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

// go run main.go apim backend create --resource-group rg-tarathec-poc-az-asse-sbx-001 --service-name apimpocazassesbx003 --backend-id hello --url https://tarathep.com --protocol http
func (a APIM) CreateOrUpdateBackend(resourceGroup, serviceName, backendID, url, protocol string) {

	color.New(color.Italic, color.FgHiBlue, color.Bold).Print("Create a new backend entity in Api Management.\n\n")

	fmt.Println("Backend ID \t:", backendID, "\nURL \t\t:", url, "\nProtocol \t:", protocol)

	color.New(color.FgHiBlack).Print("\nCreating : ")

	//Check existing backend using URL from APIM?
	beID, err := a.GetBackendIDfromURL(resourceGroup, serviceName, url)
	if err != nil {
		color.New(color.FgHiRed).Print("ERROR", err)
		os.Exit(-1)
		return
	}

	if beID != "" {
		//have exiting backend
		color.New(color.FgHiYellow).Print("This URL is using by backend-id (", beID, ") already exist on APIM\n")
		os.Exit(-1)
		return
	}

	result, err := a.createOrUpdateBackend(resourceGroup, serviceName, backendID, url, protocol)
	if err != nil {
		color.New(color.FgHiRed).Print("ERROR", err)
		os.Exit(-1)
		return
	}

	if safePointerString(result.Name) == backendID && safePointerString(result.Properties.URL) == url {
		color.New(color.FgHiGreen).Print("Done\n")
		return
	}
	color.New(color.FgHiRed).Print("ERROR", "Unknow?")
	os.Exit(-1)
}

func (apim APIM) ExportBackendsTemplate(resourceGroup, serviceName, pathBackend string) {
	if pathBackend == "" {
		pathBackend = "backends.template.json"
	} else {
		pathBackend = pathBackend + "/backends.template.json"
	}

	color.New(color.Italic, color.FgHiBlue, color.Bold).Print("Export Backends ARM template {backends.template.json} \n\n")

	backends, err := apim.getBackends(resourceGroup, serviceName, "")
	if err != nil {
		color.New(color.FgRed).Println("Fail to get APIs", err)
		return
	}

	var backendTemplate models.BackendTemplate

	for _, backend := range backends {
		//init arm header
		backendTemplate.Schema = "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#"
		backendTemplate.ContentVersion = "1.0.0.0"
		backendTemplate.Parameters.ApimServiceName.Type = "string"

		backendTemplate.Resources = append(backendTemplate.Resources,
			struct {
				Properties struct {
					Credentials struct {
						Query  struct{} "json:\"query\""
						Header struct{} "json:\"header\""
					} "json:\"credentials\""
					TLS struct {
						ValidateCertificateChain bool "json:\"validateCertificateChain\""
						ValidateCertificateName  bool "json:\"validateCertificateName\""
					} "json:\"tls\""
					URL      string "json:\"url\""
					Protocol string "json:\"protocol\""
				} "json:\"properties\""
				Name       string "json:\"name\""
				Type       string "json:\"type\""
				APIVersion string "json:\"apiVersion\""
			}{
				Properties: struct {
					Credentials struct {
						Query  struct{} "json:\"query\""
						Header struct{} "json:\"header\""
					} "json:\"credentials\""
					TLS struct {
						ValidateCertificateChain bool "json:\"validateCertificateChain\""
						ValidateCertificateName  bool "json:\"validateCertificateName\""
					} "json:\"tls\""
					URL      string "json:\"url\""
					Protocol string "json:\"protocol\""
				}{Credentials: struct {
					Query  struct{} "json:\"query\""
					Header struct{} "json:\"header\""
				}{Query: struct{}{}, Header: struct{}{}},
					TLS: struct {
						ValidateCertificateChain bool "json:\"validateCertificateChain\""
						ValidateCertificateName  bool "json:\"validateCertificateName\""
					}{ValidateCertificateChain: false, ValidateCertificateName: false},
					URL:      backend.URL,
					Protocol: backend.Protocol,
				},
				Name:       "[concat(parameters('ApimServiceName'), '/" + backend.Name + "')]",
				Type:       "Microsoft.ApiManagement/service/backends",
				APIVersion: "2021-01-01-preview",
			})
	}

	// Write to backends.template.json
	file, err := json.MarshalIndent(backendTemplate, " ", "\t")

	if err != nil {
		color.New(color.FgHiRed).Println("ERROR", err)
		os.Exit(-1)
		return
	}
	color.New(color.FgHiBlack).Print("\nExporting : ")
	if err := os.WriteFile(pathBackend, file, 0644); err != nil {
		color.New(color.FgHiRed).Println("ERROR", err)
		os.Exit(-1)
		return
	}
	color.New(color.FgHiGreen).Println("Done\n")
}

func (a APIM) ListAPIsDependingOnBackend(resourceGroup, serviceName, backendID, url string) {
	start := time.Now()
	color.New(color.Italic, color.FgHiBlue, color.Bold).Print("List API Management API's depending Backend \n\n")

	//LOAD LIST BACKEND ID AND BACKEND URL INTO CACHE MEMMORY

	//FIND BACKEND-URL LIKE COLLECT INTO BACKEND ID NAME INTO CAHCE MEM

	//
	//a.getAPIsBindingBackend(resourceGroup, serviceName, filter)
	err, apiModels := a.apis(resourceGroup, serviceName, "")
	if err != nil {
		return
	}

	var (
		maxApiNameSize, maxDisplayNameSize, maxProtocalSize, maxApiPathSize, maxApiBackendURLSize int
	)

	if len(apiModels) == 0 {
		color.New(color.FgHiBlue).Println("Not Found")
		return
	}
	// find max len values for print
	for _, api := range apiModels {
		if len(api.APIName) > maxApiNameSize {
			maxApiNameSize = len(api.APIName)
		}

		if len(api.APIDisplayName) > maxDisplayNameSize {
			maxDisplayNameSize = len(api.APIDisplayName)
		}

		maxProtocalSizeT := func() int {
			ps := ""
			for _, p := range api.APIProtocols {
				ps += " " + p
			}
			return len(ps)
		}()

		if maxProtocalSizeT > maxProtocalSize {
			maxProtocalSize = maxProtocalSizeT
		}

		if len(api.APIPath) > maxApiPathSize {
			maxApiPathSize = len(api.APIPath)
		}

		if len(api.APIBackendURL) > maxApiBackendURLSize {
			maxApiBackendURLSize = len(api.APIBackendURL)
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
	i := 0
	for _, api := range apiModels {

		var print = func() {
			color.New(color.FgHiWhite).Printf("%*d  %*s  %*s  %*s  %*s  %*s\n", 3, (i + 1), maxApiNameSize, api.APIName, maxDisplayNameSize, api.APIDisplayName, maxProtocalSize, func() string {
				ps := ""
				for _, p := range api.APIProtocols {
					ps += " " + p
				}
				return ps
			}(), maxApiPathSize, api.APIPath, maxApiBackendURLSize, api.APIBackendURL)
		}

		if backendID != "" && url != "" && api.BackendPolicyID == backendID && api.BackendPolicyURL == url {
			print()
		} else if backendID != "" && url == "" && api.BackendPolicyID == backendID {
			print()
		} else if backendID == "" && url != "" && api.BackendPolicyURL == url {
			print()
		}
	}

	fmt.Println("\nTime used is ", time.Since(start))

}
