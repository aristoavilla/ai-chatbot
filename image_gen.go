package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// LocalImageGenRequest matches the payload expected by OpenAI-compatible image generation endpoints
type LocalImageGenRequest struct {
	Model        string `json:"model"`
	Prompt       string `json:"prompt"`
	N            int    `json:"n"`
	Size         string `json:"size,omitempty"`
	Quality      string `json:"quality,omitempty"`
	Background   string `json:"background,omitempty"`
	ImageDetail  string `json:"image_detail,omitempty"`
	OutputFormat string `json:"output_format,omitempty"`
}

// LocalImageGenResponse represents the API response format from /v1/images/generations
type LocalImageGenResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL     string `json:"url,omitempty"`
		B64JSON string `json:"b64_json,omitempty"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    any    `json:"code"`
	} `json:"error,omitempty"`
}

// FormatImageDataURI ensures the returned base64 string or URL is properly formatted for browser display
func FormatImageDataURI(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "data:") || strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	// Check header bytes for PNG or JPEG
	if strings.HasPrefix(raw, "iVBORw0KGgo") {
		return "data:image/png;base64," + raw
	}
	if strings.HasPrefix(raw, "/9j/") {
		return "data:image/jpeg;base64," + raw
	}
	return "data:image/png;base64," + raw
}

// CallLocalImageGen calls the local router's /v1/images/generations endpoint to produce an AI-generated image
func CallLocalImageGen(prompt string, modelOverride ...string) (string, error) {
	apiURL := os.Getenv("LOCAL_IMAGE_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:20128/v1/images/generations"
	}

	apiKey := os.Getenv("LOCAL_IMAGE_API_KEY")

	model := os.Getenv("IMAGE_MODEL")
	if len(modelOverride) > 0 && strings.TrimSpace(modelOverride[0]) != "" {
		model = strings.TrimSpace(modelOverride[0])
	}
	if model == "" {
		model = "ag/gemini-3.1-flash-image"
	}

	reqPayload := LocalImageGenRequest{
		Model:        model,
		Prompt:       prompt,
		N:            1,
		Size:         "auto",
		Quality:      "auto",
		Background:   "auto",
		ImageDetail:  "high",
		OutputFormat: "png",
	}

	jsonData, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to encode image request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request to image service failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("image API returned HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var parsedResp LocalImageGenResponse
	if err := json.Unmarshal(bodyBytes, &parsedResp); err != nil {
		return "", fmt.Errorf("failed to parse image response: %w", err)
	}

	if parsedResp.Error != nil {
		return "", fmt.Errorf("image generation error: %s", parsedResp.Error.Message)
	}

	if len(parsedResp.Data) == 0 {
		return "", fmt.Errorf("empty image data returned from generator")
	}

	item := parsedResp.Data[0]
	if item.B64JSON != "" {
		return FormatImageDataURI(item.B64JSON), nil
	}
	if item.URL != "" {
		return item.URL, nil
	}

	return "", fmt.Errorf("no usable image format in response")
}

// GenerateImageRequest is the incoming payload for /api/generate-image
type GenerateImageRequest struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model,omitempty"`
}

// GenerateImageResponse is returned to the client
type GenerateImageResponse struct {
	ImageResult string `json:"image_result"`
	Prompt      string `json:"prompt"`
	DurationMs  int64  `json:"duration_ms"`
	Error       string `json:"error,omitempty"`
}

// handleGenerateImage handles POST requests to /api/generate-image
func handleGenerateImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req GenerateImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GenerateImageResponse{Error: "Invalid JSON payload"})
		return
	}

	if strings.TrimSpace(req.Prompt) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(GenerateImageResponse{Error: "Prompt cannot be empty"})
		return
	}

	startTime := time.Now()
	imgData, err := CallLocalImageGen(req.Prompt, req.Model)
	if err != nil {
		log.Printf("[ImageGen Handler] Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(GenerateImageResponse{
			Error: fmt.Sprintf("Failed to generate image: %v", err),
		})
		return
	}

	json.NewEncoder(w).Encode(GenerateImageResponse{
		ImageResult: imgData,
		Prompt:      req.Prompt,
		DurationMs:  time.Since(startTime).Milliseconds(),
	})
}
