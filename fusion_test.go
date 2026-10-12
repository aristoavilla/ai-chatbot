package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAvailablePresets(t *testing.T) {
	if len(AvailablePresets) == 0 {
		t.Fatalf("Expected AvailablePresets to have at least one preset")
	}

	autoPreset := GetPresetByID("auto")
	if autoPreset.ID != "auto" {
		t.Errorf("Expected auto preset, got %s", autoPreset.ID)
	}

	originalPreset := GetPresetByID("original")
	if originalPreset.ID != "auto" {
		t.Errorf("Expected original to resolve to auto, got %s", originalPreset.ID)
	}

	cyberpunk := GetPresetByID("cyberpunk")
	if cyberpunk.ID != "cyberpunk" {
		t.Errorf("Expected cyberpunk preset, got %s", cyberpunk.ID)
	}

	// Test fallback
	fallback := GetPresetByID("unknown_preset_123")
	if fallback.ID != "auto" {
		t.Errorf("Expected fallback to auto, got %s", fallback.ID)
	}
}

func TestCleanBase64(t *testing.T) {
	dataURI := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
	raw, mime := CleanBase64(dataURI, "image/jpeg")

	if mime != "image/png" {
		t.Errorf("Expected mime image/png, got %s", mime)
	}
	if raw != "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==" {
		t.Errorf("Expected raw base64 string to be extracted correctly")
	}

	plain := "plainBase64String=="
	rawPlain, mimePlain := CleanBase64(plain, "image/webp")
	if mimePlain != "image/webp" {
		t.Errorf("Expected mime image/webp, got %s", mimePlain)
	}
	if rawPlain != plain {
		t.Errorf("Expected plain string to remain unchanged")
	}
}

func TestHandlePresetsEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/presets", nil)
	w := httptest.NewRecorder()

	handlePresets(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var presets []StylePreset
	if err := json.NewDecoder(resp.Body).Decode(&presets); err != nil {
		t.Fatalf("Failed to decode presets response: %v", err)
	}

	if len(presets) < 5 {
		t.Errorf("Expected at least 5 presets, got %d", len(presets))
	}
}

func TestHandleFusionMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/fusion", nil)
	w := httptest.NewRecorder()

	handleFusion(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 Method Not Allowed, got: %d", resp.StatusCode)
	}
}

func TestHandleFusionMissingSubject(t *testing.T) {
	fusionReq := FusionRequest{
		ImageAData: "", // Missing subject
		Preset:     "cyberpunk",
	}
	body, _ := json.Marshal(fusionReq)
	req := httptest.NewRequest(http.MethodPost, "/api/fusion", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Ensure API key is set for test
	t.Setenv("GEMINI_API_KEY", "test_mock_key")

	handleFusion(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 Bad Request when Image A is missing, got: %d", resp.StatusCode)
	}
}

func TestHarmonizeImagePreset(t *testing.T) {
	// 1x1 white pixel in PNG base64
	samplePNG := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="
	res, err := HarmonizeImagePreset(samplePNG, "cyberpunk", 70)
	if err != nil {
		t.Fatalf("HarmonizeImagePreset failed: %v", err)
	}
	if !strings.HasPrefix(res, "data:image/png;base64,") && !strings.HasPrefix(res, "data:image/jpeg;base64,") {
		t.Errorf("Expected valid data URI output, got: %s", res)
	}
}

func TestSynthesizeVisionPromptFallback(t *testing.T) {
	preset := GetPresetByID("cyberpunk")
	prompt, critique := SynthesizeVisionPrompt("invalid_key_for_testing", preset, "", "", "", "", "")
	if prompt != "" || critique != "" {
		t.Errorf("Expected empty outputs on invalid request, got prompt=%q, critique=%q", prompt, critique)
	}
}

