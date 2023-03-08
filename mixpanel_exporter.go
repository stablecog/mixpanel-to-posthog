package main

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
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

func readCsvFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	return records
}

type CSVHeaderIndex struct {
	index int
	name  string
}

// Dont translate these properties
const NOOP = "NOOP"

func LoadMixpanelUsersFromCSVFile(csvFile string) ([]MixpanelUser, error) {
	// read csv from string
	records := readCsvFile(csvFile)
	headers := records[0]
	headerIndex := make(map[int]string)
	for i, header := range headers {
		if strings.HasPrefix(header, "$") {
			switch header {
			case "$mp_first_event_time":
				headerIndex[i] = "NOOP"
			case "$timezone":
				headerIndex[i] = "$geoip_time_zone"
			case "$region":
				headerIndex[i] = "$geoip_subdivision_1_name"
			case "$country_code":
				headerIndex[i] = "$geoip_country_code"
			case "$city":
				headerIndex[i] = "$geoip_city_name"
			case "$email":
				headerIndex[i] = "email"
			default:
				headerIndex[i] = header
			}
		} else {
			headerIndex[i] = header
		}
	}

	ret := []MixpanelUser{}
	for _, record := range records[1:] {
		properties := make(map[string]interface{})
		distinctId := ""
		for i, value := range record {
			if headerIndex[i] == "$distinct_id" {
				distinctId = value
				continue
			}
			if headerIndex[i] != NOOP {
				properties[headerIndex[i]] = value
			}
		}
		ret = append(ret, MixpanelUser{
			DistinctID: distinctId,
			Properties: properties,
		})
	}

	return MergeMixpanelUsers(ret), nil
}

// We need to do a merge to find valid uuids for emails and replace them
// Some mixpanel IDs are not valid uuidv4
func MergeMixpanelUsers(users []MixpanelUser) []MixpanelUser {
	// Find users with invalid uuids
	invalidIdEmailMap := make(map[string]string)

	// Loop users
	for _, user := range users {
		// If this uuid is valid
		if _, err := uuid.Parse(user.DistinctID); err == nil {
			// Get email
			email, ok := user.Properties["email"]
			if !ok {
				continue
			}
			email, ok = email.(string)
			if !ok {
				continue
			}
			// Map email to valid UUID
			invalidIdEmailMap[email.(string)] = user.DistinctID
		}
	}

	// Replace all IDs for the valid one
	for i, user := range users {
		// If user has an email
		if email, ok := user.Properties["email"]; ok {
			// If we have a valid ID for this email
			if validId, ok := invalidIdEmailMap[email.(string)]; ok {
				if _, err := uuid.Parse(users[i].DistinctID); err != nil {
					log.Info("Replacing invalid ID with valid one", "invalid", user.DistinctID, "valid", validId)
					users[i].DistinctID = validId
				}
			}
		}
	}
	return users
}

type MixpanelUser struct {
	DistinctID string
	Properties map[string]interface{}
}
