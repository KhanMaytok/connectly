package connectly

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

// BatchSendMessageRequest represents a single message to send to the Connectly API
type BatchSendMessageRequest struct {
	Number       string             `json:"number"`
	TemplateName string             `json:"templateName"`
	Language     string             `json:"language"`
	Parameters   []MessageParameter `json:"parameters"`
}

// MessageParameter the parameters to send to the API such as body_1, body_2, etc.
type MessageParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// BatchSendCampaignRequest represents the campagn to send
type BatchSendCampaignRequest struct {
	TemplateName string                    `json:"templateName"`
	Language     string                    `json:"language"`
	BusinessID   string                    `json:"businessId"`
	APIKey       string                    `json:"apiKey"`
	Messages     []BatchSendMessageRequest `json:"messages"`
	CsvFile      string                    `json:"csvFile"`
}

type APIResponse struct {
	Id  *string `json:"id"`
	Err string  `json:"error_message"`
}

// The report from the entire campaign a report per api request is good enough for now
type BatchSendCampaignResponse struct {
	Err    string        `json:"error_message"`
	Report []APIResponse `json:"report"`
}

func SendAPIMessage(message BatchSendMessageRequest, apiUrl string, apiKey string, wg *sync.WaitGroup, bscr *BatchSendCampaignResponse) {
	defer wg.Done()
	apiRes := APIResponse{}
	messageJSON, err := json.Marshal(message)
	if err != nil {
		apiRes.Err = fmt.Sprintf("Error marshalling message: %s", err)
		apiRes.Id = nil
		bscr.Report = append(bscr.Report, apiRes)
		return
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(messageJSON))
	if err != nil {
		apiRes.Err = fmt.Sprintf("Error creating request: %s", err)
		apiRes.Id = nil
		bscr.Report = append(bscr.Report, apiRes)
		return
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("x-mock-response-code", "201") // Change this to 401 to see the error mockup

	// Send the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		apiRes.Id = nil
		bscr.Report = append(bscr.Report, apiRes)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		apiRes.Err = fmt.Sprintf("Error reading response body: %s", err)
		apiRes.Id = nil
		bscr.Report = append(bscr.Report, apiRes)
		return
	}

	// for the excercise, we receive always a 201
	if resp.StatusCode != http.StatusCreated {
		apiRes.Err = fmt.Sprintf("Error sending message to %s: %s", message.Number, string(body))
		apiRes.Id = nil
		bscr.Report = append(bscr.Report, apiRes)
		return
	}

	if err := json.Unmarshal(body, &apiRes); err != nil {
		apiRes.Err = fmt.Sprintf("Error unmarshalling message: %s", err)
		apiRes.Id = nil
		bscr.Report = append(bscr.Report, apiRes)
		return
	}
	bscr.Report = append(bscr.Report, apiRes)
}

// BatchSendCampaign get the csv file and send the messages // Not async calls for now
func BatchSendCampaign(req *BatchSendCampaignRequest) *BatchSendCampaignResponse {
	var wg sync.WaitGroup

	// Initialization
	report := []APIResponse{}
	bscr := BatchSendCampaignResponse{
		Report: report,
	}
	// Download and save the csv file
	if err := DownloadCSV(req.CsvFile); err != nil {
		bscr.Err = fmt.Sprintf("Error downloading csv file: %s", err)
		return &bscr
	}

	file, err := os.Open("./campaign.csv")
	if err != nil {
		bscr.Err = fmt.Sprintf("Error opening csv file: %s", err)
	}
	defer file.Close()

	businessID := req.BusinessID
	// I know it's a static Business ID, but let's make it dynamic for the sake of the exercise
	apiUrl := fmt.Sprintf("https://cde176f9-7913-4af7-b352-75e26f94fbe3.mock.pstmn.io/v1/businesses/%s/send/whatsapp_templated_messages", businessID)

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil
	}

	for idx, record := range records {
		// Skip the header row
		if idx == 0 {
			continue
		}
		message := BatchSendMessageRequest{
			Number:       record[1],
			TemplateName: req.TemplateName, // Is record[0] the template name? csv header says channel_type and it's for session messages in the docs
			Language:     req.Language,
			Parameters: []MessageParameter{
				{Name: "body_1", Value: record[1]},
				{Name: "body_2", Value: record[2]},
			},
		}

		// One message for row now, add async support later
		wg.Add(1)
		go SendAPIMessage(message, apiUrl, req.APIKey, &wg, &bscr)
	}
	wg.Wait()
	return &bscr
}

// DownloadCSV get the csv campaign and save it locally
func DownloadCSV(url string) error {
	filepath := "./campaign.csv"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
