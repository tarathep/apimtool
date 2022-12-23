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

const version string = "v0.0.1"
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
		fmt.Print(version)
	}

	if options.Helps {
		fmt.Print(label)
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "parse":
			{
				engine.ConfigParser(options.Environment, options.ApiID)
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
							apim.ListAPI(options.ResourceGroup, options.ServiceName, options.FilterDisplayName, 1)
							return
						}

						color.New(color.FgHiWhite).Print("\n\nExamples from AI knowledge base:\n")
						color.New(color.FgHiBlue).Print("apimtool apim api list --resource-group ")
						color.New(color.FgHiWhite).Print("myresourcegroup ")
						color.New(color.FgHiBlue).Print("--service-name ")
						color.New(color.FgHiWhite).Print(" myservice\n\n")

						color.New(color.FgHiBlue).Print("apimtool apim api list --resource-group ")
						color.New(color.FgHiWhite).Print("myresourcegroup ")
						color.New(color.FgHiBlue).Print("--service-name ")
						color.New(color.FgHiWhite).Print(" myservice ")
						color.New(color.FgHiBlue).Print("--filter-display-name ")
						color.New(color.FgHiWhite).Print(" myfilterdisplay\n\n")

						color.New(color.FgCyan).Print("https://github.com/tarathep/apimtool\n")
						color.New(color.FgBlack).Print("Read more about the command in reference docs\n")

					}
				}
				if len(os.Args) > 2 && os.Args[2] == "backend" {
					if len(os.Args) > 3 && os.Args[3] == "list" {
						if options.ResourceGroup != "" && options.ServiceName != "" {
							apim.ListBackend(options.ResourceGroup, options.ServiceName)
						}
					}
				}
			}
		case "config":
			{
				if len(os.Args) > 2 && os.Args[2] == "parser" {
					engine.ConfigParser(options.Environment, options.ApiID)
					return
				}
			}

		}

	}

	fmt.Println(`
	 _   ___ ___ __  __   _____ ___   ___  _    
	/_\ | _ \_ _|  \/  | |_   _/ _ \ / _ \| |   
       / _ \|  _/| || |\/| |   | || (_) | (_) | |__ 
      /_/ \_\_| |___|_|  |_|   |_| \___/ \___/|____|
												`)

	color.New(color.FgHiWhite).Print("Welcome to the APIM Tool CLI!\n")

	color.New(color.FgHiWhite).Print("To support configuration Microsoft Azure API Managment\nUse `apimtool --version` to display the current version.\n")
	color.New(color.FgHiWhite).Print("Here are the base commands:\n\n")

	color.New(color.FgHiWhite).Print("\tparse \t: Generage Configuration files to Source files for support Deploy ARM templates via pipeline\n")
	color.New(color.FgHiWhite).Print("\tapim \t: Manage Azure API Management services.\n\n")

	color.New(color.FgCyan).Print("https://github.com/tarathep/apimtool\n")
	color.New(color.FgBlack).Print("Read more about the command in reference docs\n")
}
