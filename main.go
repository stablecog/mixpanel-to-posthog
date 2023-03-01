package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/log"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/manifoldco/promptui"
)

func main() {
	godotenv.Load(".env")

	fmt.Println("------------------------------------")
	color.Green("SC Mixpanel to Posthog Data Migrator")
	fmt.Println("------------------------------------")

	// User inputs

	// ** Get mixpanel credentials ** //

	if os.Getenv("MIXPANEL_API_URL") == "" || os.Getenv("MIXPANEL_PROJECT_ID") == "" || os.Getenv("MIXPANEL_USERNAME") == "" || os.Getenv("MIXPANEL_PASSWORD") == "" {
		color.Cyan("\nMixpanel Credentials")
		color.Cyan("See the README for reference on what these are and how to get them.\n\n")
	}
	// If in env, don't ask
	var apiUrlResult string
	if os.Getenv("MIXPANEL_API_URL") != "" {
		apiUrlResult = os.Getenv("MIXPANEL_API_URL")
	} else {
		apiUrlPrompt := promptui.Prompt{
			Label:     "Enter Mixpanel API URL",
			AllowEdit: false,
			Default:   "https://data.mixpanel.com/api/2.0",
			Validate: func(input string) error {
				// Validate URL
				_, err := url.ParseRequestURI(input)
				return err
			},
		}
		pR, _ := apiUrlPrompt.Run()
		apiUrlResult = pR
	}

	// If in env, don't ask
	var projectIdResult string
	if os.Getenv("MIXPANEL_PROJECT_ID") != "" {
		projectIdResult = os.Getenv("MIXPANEL_PROJECT_ID")
	} else {
		projectIdPrompt := promptui.Prompt{
			Label: "Enter Mixpanel Project ID",
		}
		pR, _ := projectIdPrompt.Run()
		projectIdResult = pR
	}

	// If in env, don't ask
	var serviceUsernameResult string
	if os.Getenv("MIXPANEL_USERNAME") != "" {
		serviceUsernameResult = os.Getenv("MIXPANEL_USERNAME")
	} else {
		serviceUsernamePrompt := promptui.Prompt{
			Label: "Enter Mixpanel Username (Service Account)",
		}
		pR, _ := serviceUsernamePrompt.Run()
		serviceUsernameResult = pR
	}

	// If in env, don't ask
	var servicePasswordResult string
	if os.Getenv("MIXPANEL_PASSWORD") != "" {
		servicePasswordResult = os.Getenv("MIXPANEL_PASSWORD")
	} else {
		servicePasswordPrompt := promptui.Prompt{
			Label: "Enter Mixpanel Password (Service Account)",
			Mask:  '*',
		}
		pR, _ := servicePasswordPrompt.Run()
		servicePasswordResult = pR
	}

	// ** Get Mixpanel date range ** //

	color.Yellow("\nWARNING: If you have a large dataset, consider entering smaller date ranges at a time.")
	color.Yellow("You may crash your machine if you try to export too much data at once.\n\n")

	// Prompt for from_date and to_date in the format 2006-01-02
	fromDtPrompt := promptui.Prompt{
		Label: "Enter from_date in the format YYYY-MM-DD",
		Validate: func(input string) error {
			// Validate date is in the format 2006-01-02
			_, err := time.Parse("2006-01-02", input)
			return err
		},
	}
	fromDtResult, err := fromDtPrompt.Run()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	to_date := promptui.Prompt{
		Label: "Enter to_date in the format YYYY-MM-DD",
		Validate: func(input string) error {
			// Validate date is in the format 2006-01-02
			_, err := time.Parse("2006-01-02", input)
			return err
		},
	}
	toDtResult, err := to_date.Run()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Parse dates
	fromDt, _ := time.Parse("2006-01-02", fromDtResult)
	toDt, _ := time.Parse("2006-01-02", toDtResult)

	// ** Get Posthog credentials ** //
	if os.Getenv("POSTHOG_API_KEY") == "" || os.Getenv("POSTHOG_ENDPOINT") == "" || os.Getenv("POSTHOG_PROJECT_KEY") == "" {
		color.Cyan("\nPosthog Credentials")
		color.Cyan("See the README for reference on what these are and how to get them.\n\n")
	}

	// If in env, don't ask
	var posthogApiKey string
	if os.Getenv("POSTHOG_PROJECT_KEY") != "" {
		posthogApiKey = os.Getenv("POSTHOG_PROJECT_KEY")
	} else {
		posthogApiKeyPrompt := promptui.Prompt{
			Label: "Enter Posthog Project API Key",
			Mask:  '*',
		}
		pR, _ := posthogApiKeyPrompt.Run()
		posthogApiKey = pR
	}

	var posthogPersonalApiKey string
	if os.Getenv("POSTHOG_API_KEY") != "" {
		posthogPersonalApiKey = os.Getenv("POSTHOG_API_KEY")
	} else {
		posthogApiKeyPrompt := promptui.Prompt{
			Label: "Enter Posthog Personal API Key",
			Mask:  '*',
		}
		pR, _ := posthogApiKeyPrompt.Run()
		posthogPersonalApiKey = pR
	}

	// If in env, don't ask
	var posthogEndpoint string
	if os.Getenv("POSTHOG_ENDPOINT") != "" {
		posthogApiKey = os.Getenv("POSTHOG_API_KEY")
	} else {
		posthogApiKeyPrompt := promptui.Prompt{
			Label: "Enter Posthog API Endpoint",
			Validate: func(input string) error {
				_, err := url.Parse(input)
				return err
			},
		}
		pR, _ := posthogApiKeyPrompt.Run()
		posthogApiKey = pR
	}

	// Create importer
	importer, err := NewPosthogImporter(posthogApiKey, posthogPersonalApiKey, posthogEndpoint)
	if err != nil {
		color.Red("\nEncountered an error while creating Posthog importer: %v", err)
		os.Exit(1)
	}

	// ** Mixpanel Export ** //

	// Create mixpanel exporter
	exporter := NewExporter(apiUrlResult, serviceUsernameResult, servicePasswordResult, projectIdResult, fromDt, toDt)

	color.Blue("Exporting data from Mixpanel")
	s := spinner.New(spinner.CharSets[43], 100*time.Millisecond)
	s.Reverse()
	s.Start()
	data, err := exporter.Export()
	if err != nil {
		color.Red("\nEncountered an error while exporting data from Mixpanel: %v", err)
		os.Exit(1)
	}
	s.Stop()
	color.Green("Exported %d events from Mixpanel", len(data))

	// ** Posthog Import **//

	color.Green("\nImporting data into Posthog")
	s = spinner.New(spinner.CharSets[43], 100*time.Millisecond)
	s.Start()
	err = importer.Import(data)
	if err != nil {
		color.Red("\nEncountered an error while importing data into Posthog: %v", err)
		os.Exit(1)
	}
	s.Stop()
	err = importer.Posthog.Close()
	if err != nil {
		color.Red("\nEncountered an error while closing Posthog client: %v", err)
		os.Exit(1)
	}

	os.Exit(0)
}
