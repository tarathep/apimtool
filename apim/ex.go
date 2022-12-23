// https://github.com/Azure-Samples/azure-sdk-for-go-samples/tree/main/sdk/resourcemanager/apimanagement
package apim

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/apimanagement/armapimanagement"
)

var (
	subscriptionID    = "24750e68-d6c2-40b7-90f9-f55b5009e909"
	location          = "southeastasia"
	resourceGroupName = "rg-tarathec-poc-az-asse-sbx-001"
	serviceName       = "apimpocazassesbx003"
)

func Execute() {
	// subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	// if len(subscriptionID) == 0 {
	// 	log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	// }

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// apiManagementService, err := getApiManagementService(ctx, cred)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Println("get api management service:", *apiManagementService.ID)

	ExampleAPIClient_NewListByServicePager(ctx, cred)

	//getApiOperationPolicy(ctx, cred, "echo-api", "create-resource")
}

func ExampleAPIClient_NewListByServicePager(ctx context.Context, cred azcore.TokenCredential) {
	client, err := armapimanagement.NewAPIClient(subscriptionID, cred, nil)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	pager := client.NewListByServicePager(resourceGroupName,
		serviceName,
		&armapimanagement.APIClientListByServiceOptions{Filter: nil,
			Top:                 nil,
			Skip:                nil,
			Tags:                nil,
			ExpandAPIVersionSet: nil,
		})

	fmt.Println("-------------------------------")
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to advance page: %v", err)
		}

		for _, v := range nextResult.Value {
			// TODO: use page item
			_ = v

			fmt.Println("API Name : ", *v.Properties.DisplayName)
			// fmt.Println(*v.Properties.Protocols[0])
			// fmt.Println(*v.Properties.Path)

			getApiOperation(ctx, cred, *v.Name)
			fmt.Println("-------------------------------")

		}
	}
}

func getApiOperation(ctx context.Context, cred azcore.TokenCredential, apiID string) {
	apiOperationClient, err := armapimanagement.NewAPIOperationClient(subscriptionID, cred, nil)

	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	pager := apiOperationClient.NewListByAPIPager(resourceGroupName, serviceName, apiID, &armapimanagement.APIOperationClientListByAPIOptions{
		Filter: nil,
		Top:    nil,
		Skip:   nil,
		Tags:   nil,
	})

	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to advance page: %v", err)
		}
		for _, v := range nextResult.Value {
			// TODO: use page item
			_ = v

			fmt.Println("\t", *v.Properties.Method, *v.Name)

		}
	}

}
func getApiOperationPolicy(ctx context.Context, cred azcore.TokenCredential, apiID string, operationID string) {
	apiOperationPolicyClient, err := armapimanagement.NewAPIOperationPolicyClient(subscriptionID, cred, nil)

	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	listOperation, err := apiOperationPolicyClient.ListByOperation(ctx, resourceGroupName, serviceName, apiID, operationID, &armapimanagement.APIOperationPolicyClientListByOperationOptions{})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	for _, v := range listOperation.Value {

		fmt.Println(*v.Properties.Value)
	}
}

// The resource type 'getDomainOwnershipIdentifier' could not be found in the namespace 'Microsoft.ApiManagement' for api version '2021-04-01-preview'. The supported api-versions are '2020-12-01,2021-01-01-preview'."}
func getDomainOwnershipIdentifier(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.ServiceGetDomainOwnershipIdentifierResult, error) {
	apiManagementServiceClient, err := armapimanagement.NewServiceClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := apiManagementServiceClient.GetDomainOwnershipIdentifier(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ServiceGetDomainOwnershipIdentifierResult, nil
}

func getSsoToken(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.ServiceGetSsoTokenResult, error) {
	apiManagementServiceClient, err := armapimanagement.NewServiceClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := apiManagementServiceClient.GetSsoToken(ctx, resourceGroupName, serviceName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ServiceGetSsoTokenResult, nil
}

func getApiManagementService(ctx context.Context, cred azcore.TokenCredential) (*armapimanagement.ServiceResource, error) {
	apiManagementServiceClient, err := armapimanagement.NewServiceClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := apiManagementServiceClient.Get(ctx, resourceGroupName, serviceName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ServiceResource, nil
}
