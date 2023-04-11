package main

import (
	"context"
	"fmt"
	"time"

	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/fatih/color"
	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/tarathep/apimtool/apim"
	"github.com/tarathep/apimtool/engine"
)

const version string = "v1.0.0"
const label string = `Azure API Management Tool ` + version + `
Repository : https://github.com/tarathep/apimtool`

type Options struct {
	Version bool `short:"v" long:"version" description:"Version"`

	SubscriptionID string `long:"subscription-id" description:"Subscription ID"`
	ResourceGroup  string `short:"g" long:"resource-group" description:"Resource group"`
	Location       string `short:"l" long:"location" description:"Location"`
	ServiceName    string `short:"n" long:"service-name" description:"Name"`

	Filter            string `long:"filter" description:"Filter"`
	FilterDisplayName string `long:"filter-display-name" description:"Filter of APIs by displayName."`
	Top               string `long:"top" description:"Number of records to return."`

	BackendID string `long:"backend-id" description:"Backend ID on APIM."`
	URL       string `long:"url" description:"URL endpoint"`
	Protocol  string `long:"protocol" description:"protocol to communcation"`

	FilePath    string `long:"file-path" description:"File Path"`
	Environment string `long:"env" description:"Environment"`
	ApiID       string `long:"api-id"`

	Helps   bool   `long:"help" description:"help"`
	Token   string `long:"token" description:"Personal Access Token"`
	Logging bool   `long:"logging" description:"Console log"`

	Option string `short:"o" long:"option" description:"Option"`

	Confirm bool `short:"y"`
}

func main() {
	//init log
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	log.Logger = log.With().Caller().Logger()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	var options Options
	parser := flags.NewParser(&options, flags.PrintErrors|flags.PassDoubleDash)
	if _, err := parser.Parse(); err != nil {
		log.Error().Err(err)
		color.New(color.FgHiRed).Println("Error")

	}

	flags.NewIniParser(parser)

	if options.Version {
		fmt.Print("apimtool version " + version)
		return
	}

	if options.Helps {
		fmt.Print(label)
	}

	if len(os.Args) > 1 {

		switch os.Args[1] {
		case "parse":
			{
				// PREPARATION and AUTH
				apimEnv := apim.Env()

				cred, err := azidentity.NewDefaultAzureCredential(nil)
				if err != nil {
					log.Error().Err(err).Msg("apim azidentity error")
					os.Exit(-1)
				}

				e := engine.Engine{APIM: apim.APIM{
					SubscriptionID: apimEnv.SubscriptionID,
					Location:       apimEnv.Location,
					Credential:     cred,
					Context:        context.Background(),
				}}

				if options.ResourceGroup != "" && options.ServiceName != "" && options.Environment != "" && options.ApiID != "" {
					//go run main.go parse --env dev --api-id digital-trading --resource-group rg-tarathec-poc-az-asse-sbx-001 --service-name apimpocazassesbx003
					e.ConfigParser(options.Environment, options.ApiID, options.ResourceGroup, options.ServiceName, options.FilePath)
					return
				}

				printExCommand("--resource-group/-g, --service-name/-n --env --api-id\nthe directories and config files are required: ./apis/dev/{api-id}.json ./apim-dev/sources/ ./apim-dev/templates/backends.template.json or use --file-path", true, "apimtool parse --resource-group", "myresourcegroup", "--service-name", "myservice", "--env", "dev", "--api-id", "api-name-id")
				printExCommand("", false, "apimtool parse --resource-group", "myresourcegroup", "--service-name", "myservice", "--env", "dev", "--api-id", "api-name-id", "--file-path", "./path-to-api/api.json")
				printLast()
				return
			}
		case "apim":
			{
				// PREPARATION and AUTH
				apimEnv := apim.Env()
				cred, err := azidentity.NewDefaultAzureCredential(nil)
				if err != nil {
					log.Error().Err(err).Msg("apim azidentity error")
					os.Exit(-1)
				}

				apim := apim.APIM{
					SubscriptionID: apimEnv.SubscriptionID,
					Location:       apimEnv.Location,
					Credential:     cred,
					Context:        context.Background(),
				}

				if len(os.Args) > 2 && os.Args[2] == "api" {
					if len(os.Args) > 3 && os.Args[3] == "list" {
						if options.ResourceGroup != "" && options.ServiceName != "" {
							apim.ListAPI(options.ResourceGroup, options.ServiceName, options.FilterDisplayName, options.Option)
							return
						}

						printExCommand("--resource-group/-g, --service-name/-n", true, "apimtool apim api list --resource-group", "myresourcegroup", "--service-name", "myservice")
						printExCommand("", false, "apimtool apim api list --resource-group", "myresourcegroup", "--service-name", "myservice", "--filter-display-name", "myfilterdisplay")
						printExCommand("", false, "apimtool apim api list --resource-group", "myresourcegroup", "--service-name", "myservice", "--filter-display-name", "myfilterdisplay", "--option", "table/list")

					}
				}
				if len(os.Args) > 2 && os.Args[2] == "backend" {

					if len(os.Args) > 3 && os.Args[3] == "api" {
						if len(os.Args) > 4 && os.Args[4] == "depend" {
							if len(os.Args) > 5 && os.Args[5] == "list" {
								if (options.ResourceGroup != "" && options.ServiceName != "") && (options.BackendID != "" || options.URL != "") {
									apim.ListAPIsDependingOnBackend(options.ResourceGroup, options.ServiceName, options.BackendID, options.URL)
									return
								}

								printExCommand("--resource-group/-g, --service-name/-n --backend-id or --url", true, "apimtool apim backend api depend list --resource-group", "myresourcegroup", "--service-name", "myservice", "--backend-id", "mybackend-id", "--url", "https://127.0.0.1")
								printExCommand("", false, "apimtool apim backend api depend list --resource-group", "myresourcegroup", "--service-name", "myservice", "--backend-id", "mybackend-id")
								printExCommand("", false, "apimtool apim backend api depend list --resource-group", "myresourcegroup", "--service-name", "myservice", "--url", "https://127.0.0.1")
							}
							if len(os.Args) > 5 && os.Args[5] == "update" {
								//UPDATE BACKEND TO APIS
							}
						}
					}

					if len(os.Args) > 3 && os.Args[3] == "list" {
						if options.ResourceGroup != "" && options.ServiceName != "" {
							apim.ListBackend(options.ResourceGroup, options.ServiceName, options.FilterDisplayName, options.Option)
							return
						}

						printExCommand("--resource-group/-g, --service-name/-n", true, "apimtool apim backend list --resource-group", "myresourcegroup", "--service-name", "myservice")
						printExCommand("", false, "apimtool apim backend list --resource-group", "myresourcegroup", "--service-name", "myservice", "--filter-display-name", "myfilterdisplay")
						printExCommand("", false, "apimtool apim backend list --resource-group", "myresourcegroup", "--service-name", "myservice", "--filter-display-name", "myfilterdisplay", "--option", "table/list")
					}

					//Create or Update Backend URL directly to APIM (Not update at backends.template.json)
					if len(os.Args) > 3 && os.Args[3] == "create" {
						if options.ResourceGroup != "" && options.ServiceName != "" && options.BackendID != "" {
							apim.CreateOrUpdateBackend(options.ResourceGroup, options.ServiceName, options.BackendID, options.URL, options.Protocol)
							return
						}
						//go run main.go apim backend create --resource-group rg-tarathec-poc-az-asse-sbx-001 --service-name apimpocazassesbx003 --backend-id hello --url https://tarathep.com --protocol soap
						printExCommand("--resource-group/-g, --service-name/-n --backend-id --url --protocol {http/soap}", true, "apimtool apim backend create --resource-group", "myresourcegroup", "--service-name", "myservice", "--backend-id", "my-backend-id", "--url", "https://127.0.0.1:8081", "--protocol", "http")
					}
				}
				printLast()
				return
			}
		case "template":
			{
				//trust
				if len(os.Args) > 2 && os.Args[2] == "backend" {
					if len(os.Args) > 3 && os.Args[3] == "export" {

						// PREPARATION and AUTH
						apimEnv := apim.Env()
						cred, err := azidentity.NewDefaultAzureCredential(nil)
						if err != nil {
							log.Error().Err(err).Msg("apim azidentity error")
							os.Exit(-1)
						}

						apim := apim.APIM{
							SubscriptionID: apimEnv.SubscriptionID,
							Location:       apimEnv.Location,
							Credential:     cred,
							Context:        context.Background(),
						}

						if options.ResourceGroup != "" && options.ServiceName != "" {
							apim.ExportBackendsTemplate(options.ResourceGroup, options.ServiceName, options.FilePath)
							return
						}
						printExCommand("--resource-group/-g, --service-name/-n", true, "apimtool template backend export --resource-group", "myresourcegroup", "--service-name", "myservice")
						printExCommand("--resource-group/-g, --service-name/-n", false, "apimtool template backend export --resource-group", "myresourcegroup", "--service-name", "myservice", "--file-path", "./templates/")
					}

					if len(os.Args) > 3 && os.Args[3] == "create" {
						if options.BackendID != "" && options.URL != "" && options.Protocol != "" {
							e := engine.Engine{}
							e.AddBackendTemplateJSON(options.BackendID, options.URL, options.Protocol)
							return
						}
						printExCommand("--backend-id --url --protocol {http/soap}\nthe directories and config files are required: ./templates/backends.template.json", true, "apimtool template backend create", "--backend-id", "my-backend-id", "--url", "https://127.0.0.1:8081", "--protocol", "http")
					}

					// bug
					if len(os.Args) > 3 && os.Args[3] == "delete" {
						if options.BackendID != "" {
							e := engine.Engine{}
							e.DeleteBackendTemplateJSONByID(options.BackendID)
							return
						}
						printExCommand("--env --backend-id\nthe directories and config files are required: ./templates/backends.template.json", true, "apimtool template backend delete", "--backend-id", "my-backend-id")
					}
					printLast()
					return
				}
			}
		}
	}

	color.New(color.FgHiBlue).Println(`
    _   ___ ___ __  __   _____ ___   ___  _
   /_\ | _ \_ _|  \/  | |_   _/ _ \ / _ \| |
  / _ \|  _/| || |\/| |   | || (_) | (_) | |__
 /_/ \_\_| |___|_|  |_|   |_| \___/ \___/|____|
												`)

	fmt.Print("\nWelcome to the APIM Tool CLI!\n")

	fmt.Print("To support configuration of Microsoft Azure API Management\nUse `apimtool --version` to display the current version.\n")
	fmt.Print("Here are the base commands:\n\n")

	fmt.Print("\tparse \t\t: Parsing Configuration files to Source files to support Azure API Management DevOps Resource Kit,\n\t\t\t please refer https://github.com/Azure/azure-api-management-devops-resource-kit\n")
	fmt.Print("\tapim \t\t: Manage Azure API Management services.\n")
	fmt.Print("\ttemplate \t: Manage template files configuration to support Azure Resource Manager template.\n\n")

	printLast()

}

func printExCommand(req string, header bool, options ...string) {
	if req != "" {
		color.New(color.FgHiRed).Print("the following arguments are required: " + req + "\n")
	}

	if header {
		fmt.Print("\nExamples:\n")
	}
	for i, option := range options {
		if (i % 2) == 0 {
			color.New(color.FgHiBlue).Print(option)
		} else {
			fmt.Print(option)
		}
		fmt.Print(" ")
	}
	fmt.Print("\n\n")
}

func printLast() {
	color.New(color.FgCyan).Print("https://github.com/tarathep/apimtool\n")
	color.New(color.FgHiBlack).Print("Read more about the command in reference docs\n")
}
