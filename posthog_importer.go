package main

import (
	"github.com/fatih/color"
	"github.com/posthog/posthog-go"
)

type PosthogImport struct {
	Posthog posthog.Client
}

func NewPosthogImporter(apiKey, personalKey, url string) (*PosthogImport, error) {
	client, err := posthog.NewWithConfig(apiKey, posthog.Config{
		Endpoint:       url,
		PersonalApiKey: personalKey,
	})
	if err != nil {
		return nil, err
	}
	return &PosthogImport{
		Posthog: client,
	}, nil
}

func (p *PosthogImport) Import(data []MixpanelDataLine) error {
	for _, line := range data {
		color.Cyan("Importing event: %s %s", line.Event, line.DistinctID)
		// Construct properties
		properties := posthog.NewProperties()
		for k, v := range line.Properties {
			properties.Set(k, v)
		}
		err := p.Posthog.Enqueue(posthog.Capture{
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
