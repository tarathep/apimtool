package engine

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"

	"log"
	"os"

	"github.com/tarathep/apimtool/models"
	"gopkg.in/yaml.v3"
)

type Engine struct{}

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

func getQuotedString(s string) []string {
	ms := regexp.MustCompile(`'(.*?)'`).FindAllStringSubmatch(s, -1)
	ss := make([]string, len(ms))

	for i, m := range ms {
		ss[i] = m[1]
	}
	return ss
}

func getBackendIdfromURLsourceTemplate(backendTemplate models.BackendTemplate, backendURL string) string {
	for _, resource := range backendTemplate.Resources {
		id := strings.ReplaceAll(getQuotedString(resource.Name)[1], "/", "")

		if resource.Properties.URL == backendURL {
			fmt.Println(id, resource.Properties.URL)
			return id
		}

	}
	return ""
}

func generateXMLApiPolicyHeaders(outputPath string, api models.API) error {
	apiPolictXML := models.Policies{}

	// where ip to set backend id name
	apiPolictXML.Inbound.SetBackendService.BackendID = api.Policies.BackendURL

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

func checkPaths(paths []string) bool {
	for _, checkdir := range paths {
		if _, err := os.Stat(checkdir); os.IsNotExist(err) {
			log.Println("directory " + checkdir + " not found!")
			return false
		}
	}
	return true
}

// Convert Configuration API JSON file to csv, apiPolicyHeader.xml
func ConfigParser(env, apiId string) {

	//check path
	if !checkPaths([]string{"apis/dev", "apim-dev", "apim-prd", "apim-dev/templates"}) {
		return
	}

	//pathAPIs := "./apis/" + env + "/" + apiId + ".json"
	pathBackend := "./apim-" + env + "/templates/" + "backends.template" + ".json"

	// api, _ := loadApi(pathAPIs)
	// if len(api.Operations) == 0 {
	// 	log.Fatal("API config file not found")
	// 	return
	// }

	backend, _ := loadBackendTemplate(pathBackend)
	if backend.ContentVersion == "" {
		log.Fatal("backends.template.json not found")
		return
	}

	getBackendIdfromURLsourceTemplate(backend, "https://app-sunatdav-az-usw3-dev-001.azurewebsites.net")

	// outputPath := "apim-" + env + "/" + api.Apiname
	// os.Mkdir("apim-"+env+"/"+api.Apiname, 0755)

	// generateXMLApiPolicyHeaders(outputPath, api)
	// generateCSV(outputPath, api)
	// generateConfigYML(outputPath, api)
}
