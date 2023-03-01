package main

import (
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
	"github.com/manifoldco/promptui"
)

func main() {
	godotenv.Load(".env")

	// Verify env is set
	if os.Getenv("MIXPANEL_API_URL") == "" || os.Getenv("MIXPANEL_USERNAME") == "" || os.Getenv("MIXPANEL_PASSWORD") == "" {
		log.Fatal("Please set MIXPANEL_API_URL, MIXPANEL_USERNAME, and MIXPANEL_PASSWORD in your environment")
		os.Exit(1)
	}

	validateDate := func(input string) error {
		// Validate date is in the format 2006-01-02
		_, err := time.Parse("2006-01-02", input)
		return err
	}

	// Prompt for from_date and to_date in the format 2006-01-02
	fromDtPrompt := promptui.Prompt{
		Label:    "Enter from_date in the format YYYY-MM-DD",
		Validate: validateDate,
	}
	fromDtResult, err := fromDtPrompt.Run()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	to_date := promptui.Prompt{
		Label:    "Enter to_date in the format YYYY-MM-DD",
		Validate: validateDate,
	}
	toDtResult, err := to_date.Run()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Parse dates
	fromDt, _ := time.Parse("2006-01-02", fromDtResult)
	toDt, _ := time.Parse("2006-01-02", toDtResult)

	// Create mixpanel exporter
	exporter := NewExporter(os.Getenv("MIXPANEL_API_URL"), os.Getenv("MIXPANEL_USERNAME"), os.Getenv("MIXPANEL_PASSWORD"), fromDt, toDt)

	err = exporter.Export()
	if err != nil {
		log.Fatal("Failed to export data from Mixpanel", "err", err)
	}

	os.Exit(0)
}
