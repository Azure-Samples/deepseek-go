package main

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestHandlers() *handlers {
	config := &Config{
		AzureOpenAIKey:     "test-key",
		ModelDeploymentURL: "test-url",
		ModelName:          "test-model",
		Port:               "3000",
		Template:           template.Must(template.ParseFiles("static/index.html")),
	}
	return &handlers{config: config}
}

// TestHealthHandler tests the /health endpoint to ensure it returns the expected HTTP status code and response body
func TestHealthHandler(t *testing.T) {
	// Set up the handlers with a test configuration
	h := setupTestHandlers()

	// Create a new HTTP GET request targeting the /health endpoint
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err) // Fail the test if there's an issue creating the request
	}

	// Create a ResponseRecorder to record the response
	recorder := httptest.NewRecorder()

	// Wrap the healthHandler function in an http.HandlerFunc
	handler := http.HandlerFunc(h.healthHandler)

	// Serve the HTTP request
	handler.ServeHTTP(recorder, req)

	// Check that the returned HTTP status code is 200 OK
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Define the expected response body
	expected := "Healthy."
	// Verify that the actual response body matches the expected one
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

	// Create a ResponseRecorder to capture the HTTP response
	recorder := httptest.NewRecorder()
	// Wrap the indexHandler function into an http.HandlerFunc
	handler := http.HandlerFunc(h.indexHandler)

	// Serve the HTTP request using the handler
	handler.ServeHTTP(recorder, req)

	// Verify that the returned HTTP status code is 200 OK
	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
