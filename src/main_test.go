package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

func setupTestHandlers() *handlers {
	config := &Config{
		ModelDeploymentURL: "test-url",
		ModelName:          "DeepSeek-R1",
		Port:               "3000",
		Template:           template.Must(template.ParseFiles("static/index.html")),
	}
	return &handlers{config: config}
}

// TestHealthHandler tests the /health endpoint to ensure it returns the expected HTTP status code and response body
func TestHealthHandler(t *testing.T) {
	// Set up the handlers with a test configuration
	h := setupTestHandlers()

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	handler := http.HandlerFunc(h.healthHandler)

	// Serve the HTTP request
	handler.ServeHTTP(recorder, req)

	// Check that the returned HTTP status code is 200 OK
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "Healthy."
	if recorder.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", recorder.Body.String(), expected)
	}
}

// TestIndexHandler tests the index ("/") endpoint to ensure it returns HTTP 200 OK
func TestIndexHandler(t *testing.T) {
	// Set up the test handlers with a test configuration
	h := setupTestHandlers()

	// Create a new HTTP GET request targeting the index endpoint
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	handler := http.HandlerFunc(h.indexHandler)

	handler.ServeHTTP(recorder, req)

	// Verify that the returned HTTP status code is 200 OK
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

// Create a test implementation of the makeRESTCall function
var originalMakeRESTCall func(h *handlers, messages []Message) (string, error)

// Save the original function before tests and restore after
func init() {
	originalMakeRESTCall = makeRESTCall
}

// Mock function for makeRESTCall that can be set during tests
var mockRESTCallFunc func(messages []Message) (string, error)

// Override the makeRESTCall function for testing
func makeRESTCall(h *handlers, messages []Message) (string, error) {
	if mockRESTCallFunc != nil {
		return mockRESTCallFunc(messages)
	}
	return originalMakeRESTCall(h, messages)
}

// mockTokenCredential implements azcore.TokenCredential for testing purposes
type mockTokenCredential struct {
	token azcore.AccessToken
	err   error
}

func (m *mockTokenCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return m.token, m.err
}

// TestMakeRESTCall_Success tests the makeRESTCall function for a successful REST API call
func TestMakeRESTCall_Success(t *testing.T) {
	// Setup mock HTTP server
	mockResponse := `{"choices":[{"message":{"content":"Hello, how can I help you?"}}]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, mockResponse)
	}))
	defer server.Close()

	// Setup handlers with mock config
	h := &handlers{
		config: &Config{
			ModelDeploymentURL: server.URL,
			ModelName:          "DeepSeek-R1",
		},
	}

	// Setup mock credential
	cred = &mockTokenCredential{
		token: azcore.AccessToken{
			Token:     "fake-token",
			ExpiresOn: time.Now().Add(1 * time.Hour),
		},
		err: nil,
	}

	// Execute makeRESTCall
	messages := []Message{{Role: "user", Content: "Hello"}}
	resp, err := h.makeRESTCall(messages)
	if err != nil {
		t.Fatalf("makeRESTCall returned error: %v", err)
	}

	if resp != mockResponse {
		t.Errorf("Expected response %v, got %v", mockResponse, resp)
	}
}

// TestMakeRESTCall_InvalidURL tests the makeRESTCall function for an invalid URL
func TestMakeRESTCall_InvalidJSONResponse(t *testing.T) {
	// Setup mock HTTP server to return invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "invalid-json")
	}))
	defer server.Close()

	h := &handlers{
		config: &Config{
			ModelDeploymentURL: server.URL,
			ModelName:          "DeepSeek-R1",
		},
	}

	cred = &mockTokenCredential{
		token: azcore.AccessToken{
			Token:     "fake-token",
			ExpiresOn: time.Now().Add(1 * time.Hour),
		},
		err: nil,
	}

	messages := []Message{{Role: "user", Content: "Hello"}}
	resp, err := h.makeRESTCall(messages)
	if err != nil {
		t.Fatalf("Expected no error from makeRESTCall, got %v", err)
	}

	if resp != "invalid-json" {
		t.Errorf("Expected response 'invalid-json', got %v", resp)
	}
}

// TestMakeRESTCall_TokenError tests the makeRESTCall function for token retrieval errors
func TestMakeRESTCall_TokenError(t *testing.T) {
	h := &handlers{
		config: &Config{
			ModelDeploymentURL: "http://example.com",
			ModelName:          "DeepSeek-R1",
		},
	}

	cred = &mockTokenCredential{
		token: azcore.AccessToken{},
		err:   fmt.Errorf("token error"),
	}

	messages := []Message{{Role: "user", Content: "Hello"}}
	_, err := h.makeRESTCall(messages)
	if err == nil || !strings.Contains(err.Error(), "token error") {
		t.Fatalf("Expected token error from makeRESTCall, got %v", err)
	}
}
