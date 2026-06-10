package graylog

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	BaseURL    string
	Username   string
	Password   string
	HTTPClient *http.Client
}

func NewClient(baseURL, username, password string) *Client {
	return &Client{
		BaseURL:  baseURL,
		Username: username,
		Password: password,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type SearchResponse struct {
	TotalResults int       `json:"total_results"`
	Messages     []Message `json:"messages"`
	Fields       []string  `json:"fields"`
	From         string    `json:"from"`
	To           string    `json:"to"`
}

type Message struct {
	Index     string                 `json:"index"`
	ID        string                 `json:"id"`
	Message   map[string]interface{} `json:"message"`
	Relevance float64                `json:"relevance"`
}

func (c *Client) FetchEvents(query string, limit int, from, to string) (*SearchResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", strconv.Itoa(limit))
	params.Set("from", from)
	params.Set("to", to)

	u := fmt.Sprintf("%s/api/search/universal/absolute?%s", c.BaseURL, params.Encode())
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("graylog returned %d: %s", resp.StatusCode, string(body))
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) FetchAllEvents(query string, batchSize int, callback func([]Event) error) error {
	// Split Dec 21-22 2025 into hourly windows for pagination
	start, _ := time.Parse(time.RFC3339, "2025-12-21T00:00:00Z")
	end, _ := time.Parse(time.RFC3339, "2025-12-23T00:00:00Z")
	total := 0

	for windowStart := start; windowStart.Before(end); windowStart = windowStart.Add(time.Hour) {
		windowEnd := windowStart.Add(time.Hour)
		if windowEnd.After(end) {
			windowEnd = end
		}

		from := windowStart.Format(time.RFC3339)
		to := windowEnd.Format(time.RFC3339)

		resp, err := c.FetchEvents(query, batchSize, from, to)
		if err != nil {
			return fmt.Errorf("fetch batch %s to %s: %w", from, to, err)
		}

		if resp.TotalResults == 0 {
			continue
		}

		events := make([]Event, 0, len(resp.Messages))
		for _, msg := range resp.Messages {
			events = append(events, MessageToEvent(msg))
		}

		if err := callback(events); err != nil {
			return fmt.Errorf("process batch: %w", err)
		}

		total += len(events)
		fmt.Printf("  [%s] Fetched %d events (total: %d/%d)\n",
			windowStart.Format("Jan 02 15:00"), len(events), total, 182207)
	}

	return nil
}

func MessageToEvent(msg Message) Event {
	e := Event{
		RawJSON: msg.Message,
	}

	if ts, ok := msg.Message["timestamp"].(string); ok {
		e.Timestamp = ts
	}
	if v, ok := msg.Message["source"].(string); ok {
		e.Hostname = v
	}
	if v, ok := msg.Message["event_type"].(string); ok {
		e.EventType = v
	}
	if v, ok := msg.Message["user"].(string); ok {
		e.User = v
	}
	if v, ok := msg.Message["source_ip"].(string); ok {
		e.SourceIP = v
	}
	if v, ok := msg.Message["dest_ip"].(string); ok {
		e.DestIP = v
	}
	if v, ok := msg.Message["process_name"].(string); ok {
		e.ProcessName = v
	}
	if v, ok := msg.Message["command_line"].(string); ok {
		e.CommandLine = v
	}
	if v, ok := msg.Message["parent_process"].(string); ok {
		e.ParentProcess = v
	}
	if v, ok := msg.Message["event_id"].(string); ok {
		e.EventID = v
	}
	if v, ok := msg.Message["log_type"].(string); ok {
		e.LogType = v
	}
	if v, ok := msg.Message["session_id"].(string); ok {
		e.SessionID = v
	}
	if v, ok := msg.Message["department"].(string); ok {
		e.Department = v
	}
	if v, ok := msg.Message["location"].(string); ok {
		e.Location = v
	}
	if v, ok := msg.Message["device_type"].(string); ok {
		e.DeviceType = v
	}
	if v, ok := msg.Message["success"].(string); ok {
		e.Success = v == "true"
	}
	if v, ok := msg.Message["port"].(string); ok {
		e.Port = v
	}
	if v, ok := msg.Message["protocol"].(string); ok {
		e.Protocol = v
	}
	if v, ok := msg.Message["file_path"].(string); ok {
		e.FilePath = v
	}
	if v, ok := msg.Message["severity"].(string); ok {
		e.Severity = v
	}
	if v, ok := msg.Message["error"].(string); ok {
		e.Error = v
	}

	return e
}
