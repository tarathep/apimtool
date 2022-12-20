package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
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
	ServiceName    string `short:"n" long:"name" description:"Name"`

	FilePath    string `long:"file-path" description:"File Path"`
	Environment string `long:"env" description:"Environment"`
	ApiID       string `long:"api-id"`

	Helps   bool   `long:"help" description:"help"`
	Token   string `long:"token" description:"Personal Access Token"`
	Logging bool   `long:"logging" description:"Console log"`
}

func main() {

	//engine.Execute()

	var options Options
	parser := flags.NewParser(&options, flags.PrintErrors|flags.PassDoubleDash)
	if _, err := parser.Parse(); err != nil {
		log.Fatal(err)
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
		case "config":
			{
				if len(os.Args) > 2 && os.Args[2] == "parser" {
					engine.ConfigParser(options.Environment, options.ApiID)

				}
			}
		}
	}
}
