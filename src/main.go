package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/joho/godotenv"
)

// Config holds your application configuration
type Config struct {
	ModelDeploymentURL string
	ModelName          string
	Port               string
	Template           *template.Template
}

// Represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Represents the body for the REST call
type ChatRequest struct {
	Messages []Message `json:"messages"`
	Model    string    `json:"model"`
}

// Error message structure that will be encoded as JSON
type ErrorResponse struct {
	Error string `json:"error"`
}

// Groups together HTTP handler functions and holds a reference to the application's config
type handlers struct {
	config *Config
}

// Mutex to prevent concurrent processing
var (
	mu           sync.Mutex
	isProcessing bool
)

var cred azcore.TokenCredential

func main() {

	// Load configuration
	config, err := VarsConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		return
	}

	clientID := os.Getenv("AZURE_CLIENT_ID")
	productionEnvironment := os.Getenv("RUNNING_IN_PRODUCTION")

	if productionEnvironment == "true" {
		log.Println("Running in production environment.")
		c, err := azidentity.NewManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{
			ID: azidentity.ClientID(clientID),
		})
		if err != nil {
			// return nil, trace.Wrap(err)
			return
		}
		cred = c
	} else {
		c, err := azidentity.NewAzureDeveloperCLICredential(&azidentity.AzureDeveloperCLICredentialOptions{
			TenantID: os.Getenv("AZURE_TENANT_ID"),
		})
		if err != nil {
			// return nil, trace.Wrap(err)
			return
		}
		cred = c
		log.Println("Using Azure AD authentication with AZD credentials." + os.Getenv("AZURE_TENANT_ID"))
	}

	// Serve static files from the "static" directory (index.html and styles.css)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Initialize handlers with config
	handlers := initHandlers(config)

	// Set up routes
	http.HandleFunc("/", handlers.indexHandler)
	http.HandleFunc("/chat", handlers.chatHandler)
	http.HandleFunc("/health", handlers.healthHandler)

	// Start the server
	fmt.Printf("Starting server on port %s...\n", config.Port)
	log.Fatal(http.ListenAndServe(":"+config.Port, nil))

}

// VarsConfig creates and validates a new configuration
// It loads the .env file, reads necessary environment variables, and prepares the HTML template
func VarsConfig() (*Config, error) {
	// Load .env file and log a message if the file is not found but continue
	if err := godotenv.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading .env file\n")
	}

	// load inference endpoint
	azureInferenceEndpoint := os.Getenv("AZURE_INFERENCE_ENDPOINT")
	if azureInferenceEndpoint == "" {
		return nil, fmt.Errorf("AZURE_INFERENCE_ENDPOINT not set")
	}

	modelDeploymentURL := azureInferenceEndpoint + "/chat/completions?api-version=2024-05-01-preview"

	// Load the model name
	modelName := os.Getenv("AZURE_DEEPSEEK_DEPLOYMENT")
	// Set the default model name if it is not provided.
	if modelName == "" {
		modelName = "DeepSeek-R1" // default model name
	}

	// Parse and load the HTML template from the static/index.html file.
	tmpl := template.Must(template.ParseFiles("static/index.html"))

	// Return the config struct populated with the settings
	return &Config{
		ModelDeploymentURL: modelDeploymentURL,
		ModelName:          modelName,
		Port:               "3000", // Default port number
		Template:           tmpl,
	}, nil
}

// initHandlers initializes and returns a new handlers instance with the given config
func initHandlers(config *Config) *handlers {
	return &handlers{config: config}
}

// indexHandler renders the main page using the template defined in the config (index.html)
func (h *handlers) indexHandler(w http.ResponseWriter, r *http.Request) {
	if err := h.config.Template.Execute(w, nil); err != nil {
		// If there is an error during template execution, send an HTTP 500 response with the error message
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// chatHandler handles incoming POST requests for chat messages, forwarding them to the REST API
func (h *handlers) chatHandler(w http.ResponseWriter, r *http.Request) {
	/* Set headers for CORS and content type - uncomment if needed and client is on a different domain
	** w.Header().Set("Access-Control-Allow-Origin", "*")
	** w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	** w.Header().Set("Content-Type", "application/json")
	 */

	// Handle OPTIONS preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Ensure only POST method is allowed
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Ensure only one chat request is processed at a time
	mu.Lock()
	if isProcessing {
		mu.Unlock()
		// Return 'http 429 too many requests' if a request is already being processed
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "A request is currently in progress, please retry later."})
		return
	}
	isProcessing = true
	mu.Unlock()

	// Reset the processing flag
	defer func() {
		mu.Lock()
		isProcessing = false
		mu.Unlock()
	}()

	// Decode the incoming JSON chat request
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request"})
		return
	}

	// Prepend the system message to the list of messages
	req.Messages = append([]Message{{Role: "system", Content: "You are a helpful assistant"}}, req.Messages...)

	// Make the REST call to the model deployment endpoint
	response, err := h.makeRESTCall(req.Messages)
	if err != nil {
		log.Printf("REST call failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "REST call failed"})
		return
	}

	// Send back the response from the REST call
	fmt.Fprint(w, response)
}

// healthHandler provides a simple health check endpoint.
func (h *handlers) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Healthy.")
}

// makeRESTCall sends a POST request to the configured model deployment URL with the given messages
// It returns the response body as a string or an error if the call fails
func (h *handlers) makeRESTCall(messages []Message) (string, error) {
	reqBodyStruct := ChatRequest{
		Messages: messages,
		Model:    h.config.ModelName,
	}

	// Marshal the request body into JSON
	reqBodyBytes, err := json.Marshal(reqBodyStruct)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	log.Printf("REST API call to %s | Model: %s", h.config.ModelDeploymentURL, h.config.ModelName)

	// Create a new POST request with the JSON payload
	req, err := http.NewRequest("POST", h.config.ModelDeploymentURL, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Get token for Azure Cognitive Services
	ctx := context.Background()
	token, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{"https://cognitiveservices.azure.com/.default"},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get authentication token: %w", err)
	}

	// Check if token has expired and refresh if necessary
	// Note: The token expiration time is not always accurate, so we check if the token is expired
	now := time.Now()
	if now.After(token.ExpiresOn) {
		log.Printf("Token has expired, getting a new one")
		newToken, err := cred.GetToken(ctx, policy.TokenRequestOptions{
			Scopes: []string{"https://cognitiveservices.azure.com/.default"},
		})
		if err != nil {
			return "", fmt.Errorf("failed to refresh authentication token: %w", err)
		}
		token = newToken
	}

	// Log token expiration info for debugging purposes
	// log.Printf("Using token that expires at: %v", token.ExpiresOn)

	// Set the headers for the request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer"+" "+token.Token)

	// Initialize a new HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	// Ensure the response body is closed properly
	defer resp.Body.Close()

	// Read and return the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}
