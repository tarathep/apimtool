package arm

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type ARM struct {
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

// resourceGroupName     = "deployARM-how-to"
// 	resourceGroupLocation = "eastus"
// 	deploymentName        = "deployARM-how-to"
// 	templateFile          = "template.json"

func readJSON(path string) (map[string]interface{}, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}
	contents := make(map[string]interface{})
	_ = json.Unmarshal(data, &contents)
	return contents, nil
}

func (arm ARM) GetTemplate(resourceGroupName, resourceGroupLocation, deploymentName, templateFile string) {
	_, err := armresources.NewResourceGroupsClient(arm.SubscriptionID, arm.Credential, nil)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

}

//GET https://management.azure.com/subscriptions/24750e68-d6c2-40b7-90f9-f55b5009e909/resourcegroups/rg-tarathec-poc-az-asse-sbx-001/providers/Microsoft.Resources/deployments/Microsoft.ApiManagement/service/backends?api-version=2021-04-01
