package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-key")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.apiKey != "test-key" {
		t.Errorf("expected api key test-key, got %s", client.apiKey)
	}
}

func TestChatSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Bearer test-key, got %s", r.Header.Get("Authorization"))
		}

		resp := ChatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "test response"}},
			},
			Usage: struct {
				TotalTokens int `json:"total_tokens"`
			}{TotalTokens: 100},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: server.Client(),
	}

	messages := []Message{
		{Role: "system", Content: "You are a test assistant"},
		{Role: "user", Content: "Hello"},
	}

	content, tokens, err := client.ChatWithServer(server.URL, messages, 0.2, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != "test response" {
		t.Errorf("expected test response, got %s", content)
	}
	if tokens != 100 {
		t.Errorf("expected 100 tokens, got %d", tokens)
	}
}

func TestChatError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("API error"))
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: server.Client(),
	}

	_, _, err := client.Chat([]Message{{Role: "user", Content: "test"}}, 0.2, 100)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestChatNoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ChatResponse{Choices: []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{}, Usage: struct {
			TotalTokens int `json:"total_tokens"`
		}{}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: server.Client(),
	}

	_, _, err := client.Chat([]Message{{Role: "user", Content: "test"}}, 0.2, 100)
	if err == nil {
		t.Fatal("expected error for no choices")
	}
}
