package engine

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"os"

	"github.com/fatih/color"
	"github.com/tarathep/apimtool/apim"
	"github.com/tarathep/apimtool/models"
	"gopkg.in/yaml.v3"

	"github.com/rs/zerolog/log"
)

type Engine struct {
	apim.APIM
}

func loadApi(filename string) (models.API, error) {
	file, _ := os.ReadFile(filename)

	data := models.API{}

	err := json.Unmarshal([]byte(file), &data)
	if err != nil {
		return models.API{}, err
	}

	return data, nil
}

func generateCSV(outputPath string, api models.API) error {

	records := [][]string{}

	for _, oper := range api.Operations {
		records = append(records, []string{oper.Name, oper.Method, oper.URL})
	}

	file, err := os.Create(outputPath + "/" + api.Apiname + ".csv")

	if err != nil {
		return err
	}

	w := csv.NewWriter(file)
	if err := w.WriteAll(records); err != nil {
		return err
	}

	defer file.Close()

	return nil
}

// Load backends in backends.template.json
func loadBackendTemplate(filename string) (models.BackendTemplate, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return models.BackendTemplate{}, err
	}

	data := models.BackendTemplate{}
	if err := json.Unmarshal([]byte(file), &data); err != nil {
		return models.BackendTemplate{}, err
	}

	return data, nil
}

func (e Engine) validateBackendID(backendTemplate models.BackendTemplate, resourceGroup, serviceName, url string) (bool, string) {
	backendIdSource := getBackendIDfromURLsourceTemplate(backendTemplate, url)

	backendIdAPIM, err := e.GetBackendIDfromURL(resourceGroup, serviceName, "url="+url)
	if err != nil {
		return false, ""
	}
	if backendIdSource != "" && backendIdAPIM != "" {
		log.Debug().Msgf(backendIdAPIM, backendIdSource)
		return true, backendIdAPIM
	}
	return false, ""
}

// Get exsiting Backend ID from URL source template
func getBackendIDfromURLsourceTemplate(backendTemplate models.BackendTemplate, backendURL string) string {

	u, err := url.Parse(backendURL)
	if err != nil {
		log.Error().Str("func", "getBackendIdfromURLsourceTemplate").Err(err).Msg("error on parsing URL backend")
		os.Exit(0)
	}

	for _, resource := range backendTemplate.Resources {
		id := strings.ReplaceAll(getQuotedString(resource.Name)[1], "/", "")

		port := ""
		if u.Port() != "" {
			port = ":" + u.Port()
		}
		backendURL = u.Scheme + "://" + u.Hostname() + port

		if resource.Properties.URL == backendURL {
			log.Debug().Str("func", "getBackendIdfromURLsourceTemplate").Msgf("Found ID=" + id)
			return id
		}
	}

	log.Debug().Str("func", "getBackendIdfromURLsourceTemplate").Msg("Not found")

	return ""
}

func generateXMLApiPolicyHeaders(outputPath string, api models.API, backendID string) error {
	apiPolictXML := models.Policies{}
	apiPolictXML.Inbound.SetBackendService.BackendID = backendID

	for _, policyHeaders := range api.Policies.SetHeaders {
		apiPolictXML.Inbound.SetHeader = append(apiPolictXML.Inbound.SetHeader, struct {
			Text         string "xml:\",chardata\""
			Name         string "xml:\"name,attr\""
			ExistsAction string "xml:\"exists-action,attr\""
			Value        string "xml:\"value\""
		}{
			Text:         "",
			Name:         policyHeaders.Name,
			ExistsAction: "override",
			Value:        policyHeaders.Value,
		})
	}

	file, err := xml.MarshalIndent(apiPolictXML, " ", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath+"/apiPolicyHeader.xml", file, 0644)
}

func generateConfigYML(outputPath string, api models.API) error {
	configYML := models.ConfigYML{}

	// enter value
	configYML.Version = "0.0.1"
	configYML.ApimServiceName = "apimpocazassesbx003"

	// apis
	apiConfig := models.APIConfig{}

	apiConfig.Name = api.Apiname
	apiConfig.OpenAPISpec = "./swagger.json"
	apiConfig.Policy = "./apiPolicyHeaders.xml"
	apiConfig.Suffix = api.Apiname
	apiConfig.Protocols = "https"
	apiConfig.Revision = 1
	apiConfig.AuthenticationSettings = struct {
		SubscriptionKeyRequired bool "yaml:\"subscriptionKeyRequired\""
	}{false}
	apiConfig.SubscriptionKeyParameterNames = struct {
		Header string "yaml:\"header\""
		Query  string "yaml:\"query\""
	}{"Ocp-Apim-Subscription-Key", "subscription-key"}
	apiConfig.Tags = api.Tags

	configYML.Apis = append(configYML.Apis, apiConfig)
	configYML.OutputLocation = "../../templates/apis/" + api.Apiname

	data, err := yaml.Marshal(&configYML)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath+"/config.yaml", data, 0644)
}

// Convert Configuration API JSON file to csv, apiPolicyHeader.xml
func (e Engine) ConfigParser(env, apiId, resourceGroup, serviceName string) {

	// CHECK PATH ALL OPERATIONS
	if !checkPaths([]string{"apis/dev", "apim-dev/sources", "apim-dev/templates"}) {
		return
	}

	pathAPIs := "./apis/" + env + "/" + apiId + ".json"
	pathBackend := "./apim-" + env + "/templates/" + "backends.template" + ".json"

	// LOAD CONFIGURATION FILE {apis/env/apiId.json}
	api, _ := loadApi(pathAPIs)
	if len(api.Operations) == 0 {
		color.New(color.FgYellow).Println("API config file not found")
		log.Logger.Warn().Msg("API config file not found")
		return
	}

	// LOAD LIST OF BACKEND IN backends.template.json
	backendTemplate, _ := loadBackendTemplate(pathBackend)
	if backendTemplate.ContentVersion == "" {
		color.New(color.FgYellow).Println("backends.template.json not found")
		log.Logger.Fatal().Msg("backends.template.json not found")
		return
	}

	// PREPARE OUTPUT DIRECTORY SOURCE WHEN PARSER FILE
	outputPath := "apim-" + env + "/sources/" + api.Apiname
	os.Mkdir(outputPath, 0755)

	// VALIDATE BACKEND ID IF ALREADY EXIST RETURN BACKEND ID ? CREATE NEW
	exist, backendId := e.validateBackendID(backendTemplate, resourceGroup, serviceName, api.Policies.BackendURL)
	if !exist {
		color.New(color.FgYellow).Println("new backend")
		return
	}

	generateXMLApiPolicyHeaders(outputPath, api, backendId)
	generateCSV(outputPath, api)
	generateConfigYML(outputPath, api)
}

func (Engine) removeBackendTemplateJsonByID(pathBackend string, backendTemplate models.BackendTemplate, backendID string) error {
	var beTempl models.BackendTemplate

	beTempl.Schema = backendTemplate.Schema
	beTempl.ContentVersion = backendTemplate.ContentVersion
	beTempl.Parameters = backendTemplate.Parameters

	for _, res := range backendTemplate.Resources {
		if !(res.Name == "[concat(parameters('ApimServiceName'), '/"+backendID+"')]") {
			beTempl.Resources = append(beTempl.Resources, res)
		}
	}

	file, err := json.MarshalIndent(beTempl, " ", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(pathBackend, file, 0644)
}

func (Engine) addBackendTemplateJSON(pathBackend string, backendTemplate models.BackendTemplate, backendID string, url string, protocol string) error {

	//CHECK DUPLICATE?
	for _, res := range backendTemplate.Resources {
		if res.Properties.URL == url && res.Properties.Protocol == protocol {
			return errors.New("duplicate backend endpoint at Backend ID " + res.Name)
		}
		if res.Name == "[concat(parameters('ApimServiceName'), '/"+backendID+"')]" {
			return errors.New("duplicate backend id")
		}
	}

	//APPEND
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
				URL:      url,
				Protocol: protocol,
			},
			Name:       "[concat(parameters('ApimServiceName'), '/" + backendID + "')]",
			Type:       "Microsoft.ApiManagement/service/backends",
			APIVersion: "2021-01-01-preview",
		})

	// Write to backends.template.json
	file, err := json.MarshalIndent(backendTemplate, " ", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(pathBackend, file, 0644)
}

func (e Engine) AddBackendTemplateJSON(env string, backendID string, url string, protocol string) {

	color.New(color.Italic, color.FgHiBlue, color.Bold).Print("Create a new backend entity in backends.template.json\n\n")

	fmt.Println("Backend ID \t:", backendID, "\nURL \t\t:", url, "\nProtocol \t:", protocol)

	pathBackend := "./apim-" + env + "/templates/" + "backends.template" + ".json"
	backendTemplate, _ := loadBackendTemplate(pathBackend)
	if backendTemplate.ContentVersion == "" {
		color.New(color.FgYellow).Println("backends.template.json not found")
		log.Logger.Fatal().Msg("backends.template.json not found")
		return
	}

	color.New(color.FgHiBlack).Print("\nCreating : ")
	if err := e.addBackendTemplateJSON(pathBackend, backendTemplate, backendID, url, protocol); err != nil {
		color.New(color.FgHiRed).Println("ERROR", err)
		return
	}
	color.New(color.FgHiGreen).Println("Done\n")
}

func (e Engine) DeleteBackendTemplateJSONByID(env string, backendID string) {

	color.New(color.Italic, color.FgHiYellow, color.Bold).Print("Delete a backend entity in backends.template.json\n\n")

	fmt.Println("Backend ID \t:", backendID)

	pathBackend := "./apim-" + env + "/templates/" + "backends.template" + ".json"
	backendTemplate, _ := loadBackendTemplate(pathBackend)
	if backendTemplate.ContentVersion == "" {
		color.New(color.FgYellow).Println("backends.template.json not found")
		log.Logger.Fatal().Msg("backends.template.json not found")
		return
	}

	color.New(color.FgHiBlack).Print("\nDeleing : ")
	if err := e.removeBackendTemplateJsonByID(pathBackend, backendTemplate, backendID); err != nil {
		color.New(color.FgHiRed).Println("ERROR", err)
		return
	}
	color.New(color.FgHiGreen).Println("Done\n")
}
