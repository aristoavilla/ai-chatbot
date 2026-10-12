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

func TestCallGeminiMultimodalImageGenValidation(t *testing.T) {
	_, _, err := CallGeminiMultimodalImageGen("", "directive", "rawA", "image/png", "", "")
	if err == nil {
		t.Errorf("Expected error when API key is empty")
	}

	_, _, err = CallGeminiMultimodalImageGen("key", "directive", "", "image/png", "", "")
	if err == nil {
		t.Errorf("Expected error when subject image (rawA) is empty")
	}
}

func TestCallGeminiMultimodalImageGenWithMockServer(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("Failed to decode mock payload: %v", err)
		}

		contents, ok := payload["contents"].([]interface{})
		if !ok || len(contents) == 0 {
			t.Fatalf("Expected non-empty contents array")
		}

		genConfig, ok := payload["generationConfig"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected generationConfig in payload")
		}
		modalities, ok := genConfig["responseModalities"].([]interface{})
		if !ok || len(modalities) == 0 || modalities[0] != "IMAGE" {
			t.Errorf("Expected responseModalities to contain 'IMAGE'")
		}

		// Return simulated image in candidates parts
		mockResp := map[string]interface{}{
			"candidates": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"parts": []map[string]interface{}{
							{
								"inline_data": map[string]interface{}{
									"mime_type": "image/png",
									"data":      "mockGeneratedImageBase64PNG",
								},
							},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer mockServer.Close()

	imgResult, notice, err := CallGeminiMultimodalImageGen(
		"test_key",
		"Keep subject face intact and blend with style",
		"rawSubjectBase64",
		"image/png",
		"rawStyleBase64",
		"image/jpeg",
		mockServer.URL,
	)

	if err != nil {
		t.Fatalf("Unexpected error from CallGeminiMultimodalImageGen: %v", err)
	}
	if imgResult != "data:image/png;base64,mockGeneratedImageBase64PNG" {
		t.Errorf("Expected data URI with mock base64, got %s", imgResult)
	}
	if notice != "" {
		t.Errorf("Expected empty notice on successful generation, got %s", notice)
	}
}

func TestCallGeminiMultimodalImageGen429Notice(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResp := map[string]interface{}{
			"error": map[string]interface{}{
				"code":    429,
				"message": "Resource exhausted",
				"status":  "RESOURCE_EXHAUSTED",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer mockServer.Close()

	_, notice, err := CallGeminiMultimodalImageGen(
		"test_key",
		"directive",
		"rawSubject",
		"image/png",
		"",
		"",
		mockServer.URL,
	)

	if err == nil {
		t.Errorf("Expected error when all models return 429")
	}
	if !strings.Contains(notice, "Free Tier has a limit of 0") {
		t.Errorf("Expected quota limit explanation notice, got %q", notice)
	}
}


