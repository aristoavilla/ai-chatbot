package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Part holds the text content inside a message
type Part struct {
	Text string `json:"text"`
}

// Content represents a single turn in the conversation
// Role is either "user" or "model"
type Content struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

// GeminiRequest matches the JSON payload Gemini expects
type GeminiRequest struct {
	Contents []Content `json:"contents"`
}

// Candidate represents one possible response returned by Gemini
type Candidate struct {
	Content Content `json:"content"`
}

// GeminiResponse matches the JSON returned from Gemini
type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
	Error      *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

// CallGemini sends the conversation history to the Gemini API and returns the AI's reply text
func CallGemini(apiKey string, history []Content) (string, error) {
	// We use gemini-2.5-flash for fast and smart responses
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", apiKey)

	reqBody := GeminiRequest{
		Contents: history,
	}

	// 1. Convert Go struct into JSON bytes
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to encode request: %w", err)
	}

	// 2. Prepare HTTP POST request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// 3. Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// 4. Parse Gemini's JSON response
	var geminiResp GeminiResponse
	if err := json.Unmarshal(bodyBytes, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Check if Gemini returned an API error
	if geminiResp.Error != nil {
		return "", fmt.Errorf("Gemini API error (%d): %s", geminiResp.Error.Code, geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("Gemini returned an empty response")
	}

	// Return the reply text
	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}
