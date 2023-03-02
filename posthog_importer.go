package main

import (
	"time"

	"github.com/fatih/color"
	"github.com/posthog/posthog-go"
)

func PosthogImport(client posthog.Client, data []MixpanelDataLine) error {
	for _, line := range data {
		// Construct properties
		properties := posthog.NewProperties()
		for k, v := range line.Properties {
			properties.Set(k, v)
		}
		properties.Set("$geoip_disable", true)
		err := client.Enqueue(posthog.Capture{
			DistinctId: line.DistinctID,
			Event:      line.Event,
			Properties: properties,
			Timestamp:  line.Time,
		})
		if err != nil {
			color.Red("\nError importing event: %s", line.Event)
			return err
		}
		// Sleep in between to avoid overloading the API
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func PosthogImportUsers(client posthog.Client, users []MixpanelUser) error {
	for _, user := range users {
		// Construct properties
		properties := posthog.NewProperties()
		for k, v := range user.Properties {
			if v != "undefined" {
				properties.Set(k, v)
			}
		}
		properties.Set("$geoip_disable", true)
		properties.Set("$lib", "mixpanel-importer")
		err := client.Enqueue(posthog.Identify{
			DistinctId: user.DistinctID,
			Properties: properties,
		})
		if err != nil {
			color.Red("\nError importing user: %s", user.DistinctID)
			return err
		}
		// Sleep in between to avoid overloading the API
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}
