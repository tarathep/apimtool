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
	subscriptionID := os.Getenv("APIMTOOL_AZURE_SUBSCRIPTION_ID")

	if len(subscriptionID) == 0 {
		log.Fatal("APIMTOOL_AZURE_SUBSCRIPTION_ID is not set.")
	}
	location := os.Getenv("APIMTOOL_AZURE_LOCATION") 
	if len(location) ==0{
		log.Fatal("APIMTOOL_AZURE_LOCATION is not set.")
	}
	
	return struct {
		SubscriptionID string
		Location       string
	}{SubscriptionID: subscriptionID, Location: location}
}

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