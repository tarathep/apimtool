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

const version string = "0.0.1"
const label string = `Azure API Management Tool ` + version + `
Repository : https://github.com/tarathep/apimtool`

type Options struct {
	Version bool `short:"v" long:"version" description:"Version"`

	SubscriptionID string `long:"subscription-id" description:"Subscription ID"`
	ResourceGroup  string `short:"g" long:"resource-group" description:"Resource group"`
	Location       string `short:"l" long:"location" description:"Location"`
	ServiceName    string `short:"n" long:"service-name" description:"Name"`

	FilterDisplayName string `long:"filter-display-name" description:"Filter of APIs by displayName."`
	Top               string `long:"top" description:"Number of records to return."`

	FilePath    string `long:"file-path" description:"File Path"`
	Environment string `long:"env" description:"Environment"`
	ApiID       string `long:"api-id"`

	Helps   bool   `long:"help" description:"help"`
	Token   string `long:"token" description:"Personal Access Token"`
	Logging bool   `long:"logging" description:"Console log"`

	Option string `short:"o" long:"option" description:"Option"`
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
					e.ConfigParser(options.Environment, options.ApiID, options.ResourceGroup, options.ServiceName)
					return
				}

				printExCommand("--resource-group/-g, --service-name/-n --env --api-id\nthe directories and config files are required: ./apis/dev/{api-id}.json ./apim-dev/sources/ ./apim-dev/templates/backends.template.json", true, "apimtool parse --resource-group", "myresourcegroup", "--service-name", "myservice", "--env", "dev", "--api-id", "api-name-id")

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
					if len(os.Args) > 3 && os.Args[3] == "list" {
						if options.ResourceGroup != "" && options.ServiceName != "" {
							apim.ListBackend(options.ResourceGroup, options.ServiceName, options.FilterDisplayName, options.Option)
							return
						}
						printExCommand("--resource-group/-g, --service-name/-n", true, "apimtool apim backend list --resource-group", "myresourcegroup", "--service-name", "myservice")
						printExCommand("", false, "apimtool apim backend list --resource-group", "myresourcegroup", "--service-name", "myservice", "--filter-display-name", "myfilterdisplay")
						printExCommand("", false, "apimtool apim backend list --resource-group", "myresourcegroup", "--service-name", "myservice", "--filter-display-name", "myfilterdisplay", "--option", "table/list")
					}
				}
				printLast()
				return
			}
		case "config":
			{
				if len(os.Args) > 2 && os.Args[2] == "parser" {
					//engine.ConfigParser(options.Environment, options.ApiID)
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

	fmt.Print("\tparse \t: Parsing Configuration files to Source files to support Azure API Management DevOps Resource Kit,\n\t\t please refer https://github.com/Azure/azure-api-management-devops-resource-kit\n")
	fmt.Print("\tapim \t: Manage Azure API Management services.\n\n")

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
