package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/log"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/manifoldco/promptui"
	"github.com/posthog/posthog-go"
)

var version = "dev"

// Delay between posthog queue events to avoid overloading the API
const DELAY_MS = 1

func getPosthogClient() posthog.Client {
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
		posthogEndpoint = os.Getenv("POSTHOG_ENDPOINT")
	} else {
		posthogApiKeyPrompt := promptui.Prompt{
			Label: "Enter Posthog API Endpoint",
			Validate: func(input string) error {
				_, err := url.Parse(input)
				return err
			},
		}
		pR, _ := posthogApiKeyPrompt.Run()
		posthogEndpoint = pR
	}

	// Create posthog client
	posthogClient, err := posthog.NewWithConfig(posthogApiKey, posthog.Config{
		Endpoint:       posthogEndpoint,
		PersonalApiKey: posthogPersonalApiKey,
	})
	if err != nil {
		color.Red("\nEncountered an error while creating Posthog client: %v", err)
		os.Exit(1)
	}
	return posthogClient
}

func main() {
	godotenv.Load(".env")

	fmt.Println("------------------------------------")
	color.Green("SC Mixpanel to Posthog Data Migrator")
	fmt.Println("------------------------------------")

	// They can optionally just identify users
	csvFile := flag.String("users-csv-file", "", "Path to CSV file to import users")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("\nVersion: %v\n", color.GreenString(version))
		os.Exit(0)
	}

	if *csvFile != "" {
		// See if file exists
		if _, err := os.Stat(*csvFile); os.IsNotExist(err) {
			color.Red("CSV %s file does not exist or cannot be read", *csvFile)
			os.Exit(1)
		}
		// Load from MP
		users, err := LoadMixpanelUsersFromCSVFile(*csvFile)
		if err != nil {
			color.Red("Error loading users from CSV file: %v", err)
			os.Exit(1)
		}

		// Import users
		// Calculate duration
		totalMs := DELAY_MS * 5 * len(users)
		totalDuration := time.Duration(totalMs) * time.Millisecond
		color.Cyan("Importing users from %s (This will take approximately %d minutes, the current time is %s)", *csvFile, int(totalDuration.Minutes()), time.Now().Format("15:04:05"))
		s := spinner.New(spinner.CharSets[43], 100*time.Millisecond)
		s.Start()
		posthogClient := getPosthogClient()
		defer posthogClient.Close()

		err = PosthogImportUsers(posthogClient, users)
		if err != nil {
			color.Red("Error importing users: %v", err)
			os.Exit(1)
		}
		s.Stop()
		color.Green("Successfully imported %d users", len(users))
		// Block until user presses control C
		color.Red("It's recommended to wait several minutes for posthog to process the users.")
		color.Red("Once you see all users in posthog console, you can exit this program.")
		color.Red("Press control C to exit...")
		select {}
	}

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
			Label:     "Enter Mixpanel API URL (for EU, use the EU-specific URL):",
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

	posthogClient := getPosthogClient()
	defer posthogClient.Close()

	// ** Mixpanel Export ** //

	// Create mixpanel exporter
	exporter := NewExporter(version, apiUrlResult, serviceUsernameResult, servicePasswordResult, projectIdResult, fromDt, toDt)
	color.Blue("Exporting data from Mixpanel (This may take awhile)")
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

	// ** Posthog Import ** //

	totalMs := DELAY_MS * len(data)
	totalDuration := time.Duration(totalMs) * time.Millisecond
	color.Green("\nImporting data into Posthog (This will take approximately %s, the current time is %s)", totalDuration.String(), time.Now().Format("15:04:05"))
	s.Reverse()
	s.Start()
	err = PosthogImport(posthogClient, data)
	if err != nil {
		color.Red("\nEncountered an error while importing data into Posthog: %v", err)
		os.Exit(1)
	}
	s.Stop()
	if err != nil {
		color.Red("\nEncountered an error while closing Posthog client: %v", err)
		os.Exit(1)
	}

	color.Green("Success!")
	color.Green("Imported %d events into Posthog", len(data))

	// Block until user presses control C
	color.Red("It's recommended to wait several minutes for posthog to process the events.")
	color.Red("Once you see all events in posthog, you can exit this program.")
	color.Red("Press control C to exit...")
	select {}
}
