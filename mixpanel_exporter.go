package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
)

type Mixpanel struct {
	APIUrl    string
	Token     string
	FromDate  time.Time
	ToDate    time.Time
	ProjectID string
	Client    *http.Client
	Version   string
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// Create a new mixpanel client
func NewExporter(version string, apiUrl string, user string, password string, projectId string, fromDate time.Time, toDate time.Time) *Mixpanel {
	return &Mixpanel{
		Version:   version,
		APIUrl:    apiUrl,
		Token:     basicAuth(user, password),
		FromDate:  fromDate,
		ToDate:    toDate,
		ProjectID: projectId,
		Client:    http.DefaultClient,
	}
}

func (c *Mixpanel) Export() ([]MixpanelDataLine, error) {
	// Format times to yyyy-mm-dd
	fromDate := c.FromDate.Format("2006-01-02")
	toDate := c.ToDate.Format("2006-01-02")
	url := c.APIUrl + fmt.Sprintf("/export?from_date=%s&to_date=%s&project_id=%s", fromDate, toDate, c.ProjectID)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Basic %s", c.Token))
	resp, err := c.Client.Do(request)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status=%s; httpCode=%d Export failed", resp.Status, resp.StatusCode)
	}
	defer resp.Body.Close()

	// Custom decoder since they have a wonky format
	dec := json.NewDecoder(resp.Body)
	ret := []MixpanelDataLine{}

	for {
		var line MixpanelDataLineRaw
		if err := dec.Decode(&line); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// Format the data
		formattedDataLine := MixpanelDataLine{}

		// Some events have internal names in posthog
		switch line.Event {
		case "Pageview":
			formattedDataLine.Event = "$pageview"
		default:
			formattedDataLine.Event = line.Event
		}

		// Parse properties
		formattedDataLine.Properties = make(map[string]interface{})
		formattedDataLine.Properties["$lib_version"] = fmt.Sprintf("stablecog/mp-to-ph@%s", c.Version)

		for k, v := range line.Properties {
			if k == "distinct_id" {
				formattedDataLine.DistinctID = v.(string)
			} else if k == "time" {
				// Seconds since epoch to time.Time
				formattedDataLine.Time = time.Unix(int64(v.(float64)), 0)
			} else {
				switch k {
				case "mp_lib":
					formattedDataLine.Properties["$lib"] = fmt.Sprintf("%s-imported", v)
				// Do nothing with these
				case "$mp_api_endpoint", "$mp_api_timestamp_ms", "mp_processing_time_ms":
				default:
					formattedDataLine.Properties[k] = v
				}
			}
		}

		if formattedDataLine.DistinctID == "" || formattedDataLine.Time.IsZero() {
			log.Info("Skipping event with no distinct_id or time", "event", formattedDataLine.Event)
			continue
		}
		ret = append(ret, formattedDataLine)
	}

	return ret, nil
}

type MixpanelDataLineRaw struct {
	Event      string                 `json:"event"`
	Properties map[string]interface{} `json:"properties"`
}

type MixpanelDataLine struct {
	Event      string                 `json:"event"`
	DistinctID string                 `json:"distinct_id"`
	Time       time.Time              `json:"time"`
	Properties map[string]interface{} `json:"properties"`
}
