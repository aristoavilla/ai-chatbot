package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFormatImageDataURI(t *testing.T) {
	// Already formatted data URI
	uri := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUg=="
	if got := FormatImageDataURI(uri); got != uri {
		t.Errorf("Expected %s, got %s", uri, got)
	}

	// Web URL
	httpURL := "https://example.com/generated.png"
	if got := FormatImageDataURI(httpURL); got != httpURL {
		t.Errorf("Expected %s, got %s", httpURL, got)
	}

	// PNG base64
	pngRaw := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="
	expectedPNG := "data:image/png;base64," + pngRaw
	if got := FormatImageDataURI(pngRaw); got != expectedPNG {
		t.Errorf("Expected %s, got %s", expectedPNG, got)
	}

	// JPEG base64
	jpgRaw := "/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAP//////////////////////////////////////////////////////////////////////////////////////wgALCAABAAEBAREA/8QAFBABAAAAAAAAAAAAAAAAAAAAAP/aAAgBAQABPxA="
	expectedJPG := "data:image/jpeg;base64," + jpgRaw
	if got := FormatImageDataURI(jpgRaw); got != expectedJPG {
		t.Errorf("Expected %s, got %s", expectedJPG, got)
	}
}

func TestHandleGenerateImageValidation(t *testing.T) {
	// 1. Method Not Allowed (GET)
	req := httptest.NewRequest(http.MethodGet, "/api/generate-image", nil)
	w := httptest.NewRecorder()
	handleGenerateImage(w, req)
	if w.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405 Method Not Allowed, got: %d", w.Result().StatusCode)
	}

	// 2. Empty prompt
	emptyBody, _ := json.Marshal(GenerateImageRequest{Prompt: "   "})
	reqEmpty := httptest.NewRequest(http.MethodPost, "/api/generate-image", bytes.NewReader(emptyBody))
	reqEmpty.Header.Set("Content-Type", "application/json")
	wEmpty := httptest.NewRecorder()
	handleGenerateImage(wEmpty, reqEmpty)
	if wEmpty.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request on empty prompt, got: %d", wEmpty.Result().StatusCode)
	}

	// 3. Invalid JSON
	reqInvalid := httptest.NewRequest(http.MethodPost, "/api/generate-image", strings.NewReader("invalid-json"))
	reqInvalid.Header.Set("Content-Type", "application/json")
	wInvalid := httptest.NewRecorder()
	handleGenerateImage(wInvalid, reqInvalid)
	if wInvalid.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request on invalid JSON, got: %d", wInvalid.Result().StatusCode)
	}
}

func TestCallLocalImageGenWithMockServer(t *testing.T) {
	// Create mock 9router server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
			t.Errorf("Expected Authorization Bearer test-key, got %s", auth)
		}

		var req LocalImageGenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode mock request: %v", err)
		}

		resp := LocalImageGenResponse{
			Created: 1720000000,
			Data: []struct {
				URL     string `json:"url,omitempty"`
				B64JSON string `json:"b64_json,omitempty"`
			}{
				{
					B64JSON: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	t.Setenv("LOCAL_IMAGE_API_URL", mockServer.URL)
	t.Setenv("LOCAL_IMAGE_API_KEY", "test-key")
	t.Setenv("IMAGE_MODEL", "ag/gemini-3.1-flash-image")

	imgURI, err := CallLocalImageGen("A tiny orange kitten with green eyes")
	if err != nil {
		t.Fatalf("CallLocalImageGen failed: %v", err)
	}

	if !strings.HasPrefix(imgURI, "data:image/png;base64,") {
		t.Errorf("Expected data URI prefix data:image/png;base64,, got: %s", imgURI)
	}
}
