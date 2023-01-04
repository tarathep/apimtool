package engine

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
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

type Identification struct {
	ID    string
	Phone int64
	Email string
}

func (e Engine) addBackendTemplateJSON() {

	jsonText := ""
	// define slice of Identification
	var idents []Identification

	// Unmarshall it
	if err := json.Unmarshal([]byte(jsonText), &idents); err != nil {
		fmt.Println(err)
		return
	}

	// add further value into it
	idents = append(idents, Identification{ID: "ID", Phone: 15555555555, Email: "Email"})

	// now Marshal it
	result, err := json.Marshal(idents)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(result))
}
