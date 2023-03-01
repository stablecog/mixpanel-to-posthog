package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Mixpanel struct {
	APIUrl   string
	Token    string
	FromDate time.Time
	ToDate   time.Time
	Client   *http.Client
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// Create a new mixpanel client
func NewExporter(apiUrl string, user string, password string, fromDate time.Time, toDate time.Time) *Mixpanel {
	return &Mixpanel{
		APIUrl:   apiUrl,
		Token:    basicAuth(user, password),
		FromDate: fromDate,
		ToDate:   toDate,
		Client:   http.DefaultClient,
	}
}

func (c *Mixpanel) Export() error {
	// Format times to yyyy-mm-dd
	fromDate := c.FromDate.Format("2006-01-02")
	toDate := c.ToDate.Format("2006-01-02")
	url := c.APIUrl + fmt.Sprintf("/export?from_date=%s&to_date=%s", fromDate, toDate)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Basic %s", c.Token))
	resp, err := c.Client.Do(request)
	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("status=%s; httpCode=%d Export failed", resp.Status, resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	type verboseResponse struct {
		Error  string `json:"error"`
		Status int    `json:"status"`
	}

	var jsonBody map[string]interface{}
	json.Unmarshal(body, &jsonBody)

	// Print json as string
	fmt.Println(jsonBody)

	return nil
}
