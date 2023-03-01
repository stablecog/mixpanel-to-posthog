package main

import (
	"github.com/fatih/color"
	"github.com/posthog/posthog-go"
)

func PosthogImport(client posthog.Client, data []MixpanelDataLine) error {
	for _, line := range data {
		color.Cyan("Importing event: %s %s", line.Event, line.DistinctID)
		// Construct properties
		properties := posthog.NewProperties()
		for k, v := range line.Properties {
			properties.Set(k, v)
		}
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
	}
	return nil
}
