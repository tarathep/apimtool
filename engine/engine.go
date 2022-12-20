package engine

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"

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

func GetBackendIPfromURL(source bool, backendURL string) {

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

// Convert Configuration API JSON file to csv, apiPolicyHeader.xml
func ConfigParser(env, apiId string) {

	for _, checkdir := range []string{"apis/dev", "apim-dev", "apim-prd"} {
		if _, err := os.Stat(checkdir); os.IsNotExist(err) {
			fmt.Println("directory " + checkdir + " not found!")
			return
		}
	}

	api, _ := loadApi("./apis/" + env + "/" + apiId + ".json")
	if len(api.Operations) == 0 {
		log.Fatal("API config file not found")
		return
	}
	outputPath := "apim-" + env + "/" + api.Apiname
	os.Mkdir("apim-"+env+"/"+api.Apiname, 0755)

	generateXMLApiPolicyHeaders(outputPath, api)
	generateCSV(outputPath, api)
	generateConfigYML(outputPath, api)
}
