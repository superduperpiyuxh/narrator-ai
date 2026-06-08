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
			Timeout: 30 * time.Second,
		},
	}
}

type SearchResponse struct {
	TotalResults int             `json:"total_results"`
	Messages     []Message       `json:"messages"`
	Fields       []string        `json:"fields"`
	From         string          `json:"from"`
	To           string          `json:"to"`
	After        json.RawMessage `json:"after,omitempty"`
}

type Message struct {
	Index     string                 `json:"index"`
	ID        string                 `json:"id"`
	Message   map[string]interface{} `json:"message"`
	Relevance float64                `json:"relevance"`
}

func (c *Client) doRequest(method, path string, params url.Values) (*http.Request, error) {
	u := fmt.Sprintf("%s%s", c.BaseURL, path)
	if params != nil {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequest(method, u, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (c *Client) FetchEvents(query string, limit int, after json.RawMessage) (*SearchResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", strconv.Itoa(limit))
	params.Set("range", "0")

	req, err := c.doRequest("GET", "/api/search/universal/relative", params)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

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
	var after json.RawMessage
	total := 0

	for {
		resp, err := c.FetchEvents(query, batchSize, after)
		if err != nil {
			return fmt.Errorf("fetch batch at offset %d: %w", total, err)
		}

		events := make([]Event, 0, len(resp.Messages))
		for _, msg := range resp.Messages {
			events = append(events, MessageToEvent(msg))
		}

		if err := callback(events); err != nil {
			return fmt.Errorf("process batch: %w", err)
		}

		total += len(events)
		fmt.Printf("  Fetched %d/%d events\n", total, resp.TotalResults)

		if len(resp.Messages) < batchSize || total >= resp.TotalResults {
			break
		}

		after = resp.After
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
	if v, ok := msg.Message["_event_type"].(string); ok {
		e.EventType = v
	}
	if v, ok := msg.Message["_user"].(string); ok {
		e.User = v
	}
	if v, ok := msg.Message["_source_ip"].(string); ok {
		e.SourceIP = v
	}
	if v, ok := msg.Message["_dest_ip"].(string); ok {
		e.DestIP = v
	}
	if v, ok := msg.Message["_process_name"].(string); ok {
		e.ProcessName = v
	}
	if v, ok := msg.Message["_command_line"].(string); ok {
		e.CommandLine = v
	}
	if v, ok := msg.Message["_parent_process"].(string); ok {
		e.ParentProcess = v
	}
	if v, ok := msg.Message["_event_id"].(string); ok {
		e.EventID = v
	}
	if v, ok := msg.Message["_log_type"].(string); ok {
		e.LogType = v
	}
	if v, ok := msg.Message["_session_id"].(string); ok {
		e.SessionID = v
	}
	if v, ok := msg.Message["_department"].(string); ok {
		e.Department = v
	}
	if v, ok := msg.Message["_location"].(string); ok {
		e.Location = v
	}
	if v, ok := msg.Message["_device_type"].(string); ok {
		e.DeviceType = v
	}
	if v, ok := msg.Message["_success"].(string); ok {
		e.Success = v == "true"
	}
	if v, ok := msg.Message["_port"].(string); ok {
		e.Port = v
	}
	if v, ok := msg.Message["_protocol"].(string); ok {
		e.Protocol = v
	}
	if v, ok := msg.Message["_file_path"].(string); ok {
		e.FilePath = v
	}
	if v, ok := msg.Message["_severity"].(string); ok {
		e.Severity = v
	}
	if v, ok := msg.Message["_error"].(string); ok {
		e.Error = v
	}

	return e
}
