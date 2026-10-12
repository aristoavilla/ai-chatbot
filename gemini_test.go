package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestGeminiRequestSerialization(t *testing.T) {
	req := GeminiRequest{
		SystemInstruction: &Content{
			Parts: []Part{{Text: "You are a Go reviewer"}},
		},
		Contents: []Content{
			{
				Role:  "user",
				Parts: []Part{{Text: "Hello"}},
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"system_instruction"`) {
		t.Errorf("Expected JSON to contain system_instruction, got: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"You are a Go reviewer"`) {
		t.Errorf("Expected JSON to contain prompt text, got: %s", jsonStr)
	}
}

func TestPersonasMap(t *testing.T) {
	expectedPersonas := []string{"default", "go_reviewer", "english_tutor", "eli5", "pirate"}
	for _, p := range expectedPersonas {
		prompt, exists := personas[p]
		if !exists || strings.TrimSpace(prompt) == "" {
			t.Errorf("Expected persona %q to exist and have non-empty prompt", p)
		}
	}
}

func TestHandleChatMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/chat", nil)
	w := httptest.NewRecorder()

	handleChat(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 Method Not Allowed, got: %d", resp.StatusCode)
	}
}

func TestHandleChatEmptyMessage(t *testing.T) {
	chatReq := ChatRequest{
		Message: "   ",
		Persona: "default",
	}
	body, _ := json.Marshal(chatReq)
	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handleChat(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request, got: %d", resp.StatusCode)
	}
}

func TestHandleChatMissingAPIKey(t *testing.T) {
	// Temporarily unset GEMINI_API_KEY for this test
	origKey := os.Getenv("GEMINI_API_KEY")
	os.Unsetenv("GEMINI_API_KEY")
	defer func() {
		if origKey != "" {
			os.Setenv("GEMINI_API_KEY", origKey)
		}
	}()

	chatReq := ChatRequest{
		Message: "Hello",
		Persona: "go_reviewer",
	}
	body, _ := json.Marshal(chatReq)
	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handleChat(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500 Internal Server Error, got: %d", resp.StatusCode)
	}

	var res ChatResponse
	json.NewDecoder(w.Body).Decode(&res)
	if !strings.Contains(res.Error, "GEMINI_API_KEY is not set") {
		t.Errorf("Expected error about missing GEMINI_API_KEY, got: %s", res.Error)
	}
}
