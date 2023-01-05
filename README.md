# APIM Tool
Azure API Management Tool

## Using Azure CLI to Sign In

You could easily use az login in command line to sign in to Azure via your default browser. Detail instructions can be found in Sign in with Azure CLI.

```bash
az login
```


```bash
go run main.go apim backend list --resource-group rg-tarathec-poc-az-asse-sbx-001 --service-name apimpocazassesbx003
```


```bash
go run main.go apim backend create --resource-group rg-tarathec-poc-az-asse-sbx-001 --service-name apimpocazassesbx003 --backend-id hello --url https://tarathep.com --protocol http
```

```bash
apimtool parse --env dev --api-id digital-trading --resource-group rg-tarathec-poc-az-asse-sbx-001 --service-name apimpocazassesbx003
```

```bash
go run main.go template backend create --env dev --resource-group rg-tarathec-poc-az-asse-sbx-001 --service-name apimpocazassesbx003 --backend-id hello --url https://tarathep.com --protocol http
```
