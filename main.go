package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// ChatRequest is what the browser sends to our Go server
type ChatRequest struct {
	Message string    `json:"message"`
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

	// 2. Append the new user message to the conversation history
	userContent := Content{
		Role:  "user",
		Parts: []Part{{Text: req.Message}},
	}
	updatedHistory := append(req.History, userContent)

	// 3. Call the Gemini API with the full conversation history
	reply, err := CallGemini(apiKey, updatedHistory)
	if err != nil {
		log.Printf("Error calling Gemini: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ChatResponse{Error: fmt.Sprintf("AI error: %v", err)})
		return
	}

	// 4. Append the model's reply to the history so future turns have context
	modelContent := Content{
		Role:  "model",
		Parts: []Part{{Text: reply}},
	}
	updatedHistory = append(updatedHistory, modelContent)

	// 5. Send the reply and updated history back to the browser
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
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// 2. Chat API endpoint
	http.HandleFunc("/api/chat", handleChat)

	fmt.Printf("🚀 Server running at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
