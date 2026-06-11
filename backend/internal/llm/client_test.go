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
		models:     []string{"test-model"},
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
		models:     []string{"test-model"},
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
		models:     []string{"test-model"},
	}

	_, _, err := client.Chat([]Message{{Role: "user", Content: "test"}}, 0.2, 100)
	if err == nil {
		t.Fatal("expected error for no choices")
	}
}

func TestModelRotation(t *testing.T) {
	var requestedModels []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		json.NewDecoder(r.Body).Decode(&req)
		requestedModels = append(requestedModels, req.Model)

		resp := ChatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "ok"}},
			},
			Usage: struct {
				TotalTokens int `json:"total_tokens"`
			}{TotalTokens: 10},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: server.Client(),
		models:     []string{"model-a", "model-b", "model-c"},
	}

	for i := 0; i < 6; i++ {
		client.ChatWithServer(server.URL, []Message{{Role: "user", Content: "test"}}, 0.2, 10)
	}

	// Should rotate: a, b, c, a, b, c
	expected := []string{"model-a", "model-b", "model-c", "model-a", "model-b", "model-c"}
	for i, want := range expected {
		if requestedModels[i] != want {
			t.Errorf("request %d: expected %s, got %s", i, want, requestedModels[i])
		}
	}
}

func TestChatWithServer_Headers(t *testing.T) {
	var headersReceived map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headersReceived = map[string]string{
			"Authorization": r.Header.Get("Authorization"),
			"HTTP-Referer":  r.Header.Get("HTTP-Referer"),
			"X-Title":       r.Header.Get("X-Title"),
			"Content-Type":  r.Header.Get("Content-Type"),
		}

		resp := ChatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: "ok"}},
			},
			Usage: struct {
				TotalTokens int `json:"total_tokens"`
			}{TotalTokens: 5},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("my-api-key")
	client.httpClient = server.Client()

	client.ChatWithServer(server.URL, []Message{{Role: "user", Content: "test"}}, 0.5, 50)

	if headersReceived["Authorization"] != "Bearer my-api-key" {
		t.Errorf("expected Bearer my-api-key, got %s", headersReceived["Authorization"])
	}
	if headersReceived["HTTP-Referer"] != "https://narrator-ai.dev" {
		t.Errorf("expected narrator-ai.dev referer, got %s", headersReceived["HTTP-Referer"])
	}
	if headersReceived["X-Title"] != "NarratorAI" {
		t.Errorf("expected NarratorAI title, got %s", headersReceived["X-Title"])
	}
	if headersReceived["Content-Type"] != "application/json" {
		t.Errorf("expected application/json, got %s", headersReceived["Content-Type"])
	}
}

func TestChatWithServer_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not valid json at all"))
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: server.Client(),
		models:     []string{"test-model"},
	}

	_, _, err := client.ChatWithServer(server.URL, []Message{{Role: "user", Content: "test"}}, 0.2, 10)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestNewClient_Models(t *testing.T) {
	client := NewClient("key")
	if len(client.models) != 3 {
		t.Errorf("expected 3 models, got %d", len(client.models))
	}
	if client.models[0] != "openai/gpt-4o-mini" {
		t.Errorf("expected gpt-4o-mini, got %s", client.models[0])
	}
}

func TestNextModel_RoundRobin(t *testing.T) {
	client := &Client{
		models: []string{"a", "b", "c"},
	}

	for i := 0; i < 6; i++ {
		model := client.nextModel()
		expected := []string{"a", "b", "c", "a", "b", "c"}[i]
		if model != expected {
			t.Errorf("iteration %d: expected %s, got %s", i, expected, model)
		}
	}
}
