package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// Predefined personas mapping keys to their Gemini system instructions
var personas = map[string]string{
	"default":           "You are a helpful, friendly, and knowledgeable AI assistant. Provide concise, clear, and well-structured answers.",
	"creative_director": "You are an elite AI Creative Director and Visual Stylist. Guide users on aesthetic compositions, color harmony, visual style fusion prompts, and lighting design.",
	"go_reviewer":       "You are an expert, strict Go (Golang) code reviewer. Follow Effective Go and Go best practices. Thoroughly inspect code for idiomatic patterns, error handling, concurrency safety, memory allocations, and proper formatting. Always provide constructive critiques and refactored Go examples.",
	"english_tutor":     "You are a warm, encouraging English language tutor. Review the user's messages for grammatical accuracy, spelling, and tone. Point out mistakes politely, suggest natural phrasing or idiomatic expressions, and explain the reasoning clearly.",
	"eli5":              "You explain everything like I'm 10 years old (ELI5). Break down complex concepts using simple everyday analogies, clear language, and relatable examples. Keep explanations fun, engaging, and jargon-free.",
	"pirate":            "Ahoy! You are a witty, boisterous pirate captain sailing the high seas of code and knowledge. Speak with nautical flair and pirate vocabulary while still delivering accurate, helpful answers.",
}

// ChatRequest is what the browser sends to our Go server
type ChatRequest struct {
	Message string    `json:"message"`
	Persona string    `json:"persona,omitempty"`
	History []Content `json:"history"`
}

// ChatResponse is what our Go server sends back to the browser
type ChatResponse struct {
	Reply   string    `json:"reply"`
	History []Content `json:"history"`
	Error   string    `json:"error,omitempty"`
}

// Simple helper to load key=value from a .env file into os.Getenv
func loadEnv(filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return // File doesn't exist, ignore
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			// Strip surrounding quotes if present
			v = strings.Trim(v, `"'`)
			if os.Getenv(k) == "" {
				os.Setenv(k, v)
			}
		}
	}
}

func handlePresets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AvailablePresets)
}

func handleFusion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(FusionResponse{
			Error: "GEMINI_API_KEY is not set. Please add it to your .env file.",
		})
		return
	}

	var req FusionRequest
	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Limit to 25MB total multipart upload
		if err := r.ParseMultipartForm(25 << 20); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(FusionResponse{Error: "Failed to parse multipart form"})
			return
		}

		req.Preset = r.FormValue("preset")
		req.Prompt = r.FormValue("prompt")

		// Read Image A
		fileA, headerA, errA := r.FormFile("image_a")
		if errA == nil && fileA != nil {
			defer fileA.Close()
			bytesA, _ := io.ReadAll(fileA)
			req.ImageAMime = headerA.Header.Get("Content-Type")
			if req.ImageAMime == "" {
				req.ImageAMime = "image/jpeg"
			}
			req.ImageAData = base64.StdEncoding.EncodeToString(bytesA)
		}

		// Read Image B
		fileB, headerB, errB := r.FormFile("image_b")
		if errB == nil && fileB != nil {
			defer fileB.Close()
			bytesB, _ := io.ReadAll(fileB)
			req.ImageBMime = headerB.Header.Get("Content-Type")
			if req.ImageBMime == "" {
				req.ImageBMime = "image/jpeg"
			}
			req.ImageBData = base64.StdEncoding.EncodeToString(bytesB)
		}
	} else {
		// JSON payload
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(FusionResponse{Error: "Invalid JSON request payload"})
			return
		}
	}

	if strings.TrimSpace(req.ImageAData) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(FusionResponse{Error: "Subject image (Image A) is required for fusion"})
		return
	}

	resp, err := CallGeminiFusion(apiKey, req)
	if err != nil {
		log.Printf("Error during fusion: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(FusionResponse{
			Error: fmt.Sprintf("AI Fusion error: %v", err),
		})
		return
	}

	json.NewEncoder(w).Encode(resp)
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// 1. Decode the JSON request sent by the browser
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ChatResponse{Error: "Invalid request payload"})
		return
	}

	if strings.TrimSpace(req.Message) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ChatResponse{Error: "Message cannot be empty"})
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ChatResponse{
			Error: "GEMINI_API_KEY is not set. Please add it to your .env file.",
		})
		return
	}

	// 2. Resolve system instruction from persona
	systemPrompt, exists := personas[req.Persona]
	if !exists || strings.TrimSpace(systemPrompt) == "" {
		systemPrompt = personas["default"]
	}

	// 3. Append the new user message to the conversation history
	userContent := Content{
		Role:  "user",
		Parts: []Part{{Text: req.Message}},
	}
	updatedHistory := append(req.History, userContent)

	// 4. Call the Gemini API with the system prompt and updated conversation history
	reply, err := CallGemini(apiKey, systemPrompt, updatedHistory)
	if err != nil {
		log.Printf("Error calling Gemini: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ChatResponse{Error: fmt.Sprintf("AI error: %v", err)})
		return
	}

	// 5. Append the model's reply to the history so future turns have context
	modelContent := Content{
		Role:  "model",
		Parts: []Part{{Text: reply}},
	}
	updatedHistory = append(updatedHistory, modelContent)

	// 6. Send the reply and updated history back to the browser
	json.NewEncoder(w).Encode(ChatResponse{
		Reply:   reply,
		History: updatedHistory,
	})
}

func main() {
	// Load .env if present
	loadEnv(".env")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 1. Serve frontend static files (HTML, CSS, JS) from the "static" folder
	// Cache-Control: no-cache forces the browser to revalidate, so edits to
	// app.js / style.css are always picked up on reload during development.
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		fs.ServeHTTP(w, r)
	}))

	// 2. Fusion Studio API endpoints
	http.HandleFunc("/api/presets", handlePresets)
	http.HandleFunc("/api/fusion", handleFusion)

	// 3. Chat API endpoint (Creative Director / Assistant)
	http.HandleFunc("/api/chat", handleChat)

	// 4. Standalone Image Generation API endpoint (Local Router / ag/gemini-3.1-flash-image)
	http.HandleFunc("/api/generate-image", handleGenerateImage)

	fmt.Printf("🚀 ChromaFuse AI Studio running at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
