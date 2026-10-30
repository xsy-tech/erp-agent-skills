package mcpclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type noInput struct{}
type testOutput struct {
	OK bool `json:"ok"`
}

func TestDoctorInitializesAndListsSortedTools(t *testing.T) {
	const key = "test-personal-key"
	server := mcp.NewServer(
		&mcp.Implementation{Name: "erp-open-platform", Version: "1.2.3"},
		&mcp.ServerOptions{Instructions: "Use tools/list as runtime authority."},
	)
	for _, name := range []string{"z_tool", "a_tool"} {
		mcp.AddTool(server, &mcp.Tool{Name: name, Description: "test"}, func(context.Context, *mcp.CallToolRequest, noInput) (*mcp.CallToolResult, testOutput, error) {
			return nil, testOutput{OK: true}, nil
		})
	}
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
	protected := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer "+key {
			http.Error(writer, "unauthorized", http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(writer, request)
	})
	testServer := httptest.NewServer(protected)
	defer testServer.Close()

	result, err := New().Doctor(t.Context(), testServer.URL, key)
	if err != nil {
		t.Fatalf("Doctor() error = %v", err)
	}
	if result.ServerName != "erp-open-platform" || result.ServerVersion != "1.2.3" {
		t.Fatalf("unexpected server identity: %+v", result)
	}
	if result.Instructions == "" {
		t.Fatal("Doctor() instructions are empty")
	}
	if strings.Join(result.Tools, ",") != "a_tool,z_tool" {
		t.Fatalf("Doctor() tools = %v, want sorted tools", result.Tools)
	}
}

func TestDoctorAppliesTimeout(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer testServer.Close()
	client := New()
	client.Timeout = 20 * time.Millisecond
	_, err := client.Doctor(t.Context(), testServer.URL, "secret")
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Doctor() error = %v, want deadline exceeded", err)
	}
}

type failingTransport struct {
	secret string
}

func (t failingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("transport failed with " + t.secret)
}

func TestDoctorRedactsKeyFromErrors(t *testing.T) {
	const secret = "recognizable-secret-value"
	client := New()
	client.HTTPClient = &http.Client{Transport: failingTransport{secret: secret}}
	_, err := client.Doctor(t.Context(), "https://erp.example.com/mcp", secret)
	if err == nil {
		t.Fatal("Doctor() error = nil, want error")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("Doctor() leaked key: %v", err)
	}
}
