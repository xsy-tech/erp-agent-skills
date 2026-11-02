package mcpclient

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type noInput struct{}
type testOutput struct {
	OK bool `json:"ok"`
}

func TestDoctorVerifiesPlatformInfoAfterListingAllTools(t *testing.T) {
	const key = "test-personal-key"
	var platformCalls atomic.Int32
	var writeCalls atomic.Int32
	arguments := make(chan string, 2)
	server := mcp.NewServer(
		&mcp.Implementation{Name: "erp-open-platform", Version: "1.2.3"},
		&mcp.ServerOptions{Instructions: "Use tools/list as runtime authority.", PageSize: 1},
	)
	mcp.AddTool(server, &mcp.Tool{Name: "z_read_tool", Description: "read"}, func(context.Context, *mcp.CallToolRequest, noInput) (*mcp.CallToolResult, testOutput, error) {
		return nil, testOutput{OK: true}, nil
	})
	mcp.AddTool(server, &mcp.Tool{Name: "get_platform_info", Description: "verify"}, func(_ context.Context, request *mcp.CallToolRequest, _ noInput) (*mcp.CallToolResult, testOutput, error) {
		platformCalls.Add(1)
		arguments <- string(request.Params.Arguments)
		return nil, testOutput{OK: true}, nil
	})
	mcp.AddTool(server, &mcp.Tool{Name: "create_order", Description: "write"}, func(context.Context, *mcp.CallToolRequest, noInput) (*mcp.CallToolResult, testOutput, error) {
		writeCalls.Add(1)
		return nil, testOutput{OK: true}, nil
	})
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
	if result.Status != DoctorStatusVerified {
		t.Fatalf("Doctor() status = %q, want %q", result.Status, DoctorStatusVerified)
	}
	if strings.Join(result.Tools, ",") != "create_order,get_platform_info,z_read_tool" {
		t.Fatalf("Doctor() tools = %v, want sorted tools", result.Tools)
	}
	if result.Verification == nil || result.Verification.Tool != "get_platform_info" {
		t.Fatalf("Doctor() verification = %#v", result.Verification)
	}
	output, err := json.Marshal(result.Verification.Output)
	if err != nil {
		t.Fatalf("marshal verification output: %v", err)
	}
	if string(output) != `{"ok":true}` {
		t.Fatalf("Doctor() verification output = %s", output)
	}
	if got := platformCalls.Load(); got != 1 {
		t.Fatalf("get_platform_info calls = %d, want 1", got)
	}
	select {
	case got := <-arguments:
		if got != "{}" {
			t.Fatalf("get_platform_info arguments = %q, want {}", got)
		}
	case <-time.After(time.Second):
		t.Fatal("get_platform_info arguments were not recorded")
	}
	if got := writeCalls.Load(); got != 0 {
		t.Fatalf("write tool calls = %d, want 0", got)
	}
}

func TestDoctorRequiresAuthorizationWithoutPlatformInfo(t *testing.T) {
	const key = "test-personal-key"
	for _, test := range []struct {
		name      string
		addTools  func(*mcp.Server, *atomic.Int32)
		wantTools string
	}{
		{name: "no visible tools"},
		{
			name: "platform info missing",
			addTools: func(server *mcp.Server, calls *atomic.Int32) {
				mcp.AddTool(server, &mcp.Tool{Name: "create_order", Description: "write"}, func(context.Context, *mcp.CallToolRequest, noInput) (*mcp.CallToolResult, testOutput, error) {
					calls.Add(1)
					return nil, testOutput{OK: true}, nil
				})
			},
			wantTools: "create_order",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var calls atomic.Int32
			server := mcp.NewServer(&mcp.Implementation{Name: "erp-open-platform", Version: "1.2.3"}, nil)
			if test.addTools != nil {
				test.addTools(server, &calls)
			}
			handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
			testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.Header.Get("Authorization") != "Bearer "+key {
					http.Error(writer, "unauthorized", http.StatusUnauthorized)
					return
				}
				handler.ServeHTTP(writer, request)
			}))
			defer testServer.Close()

			result, err := New().Doctor(t.Context(), testServer.URL, key)
			if err != nil {
				t.Fatalf("Doctor() error = %v", err)
			}
			if result.Status != DoctorStatusAuthorizationRequired {
				t.Fatalf("Doctor() status = %q, want %q", result.Status, DoctorStatusAuthorizationRequired)
			}
			if strings.Join(result.Tools, ",") != test.wantTools {
				t.Fatalf("Doctor() tools = %v, want %q", result.Tools, test.wantTools)
			}
			if result.Verification != nil {
				t.Fatalf("Doctor() verification = %#v, want nil", result.Verification)
			}
			if got := calls.Load(); got != 0 {
				t.Fatalf("tool calls = %d, want 0", got)
			}
		})
	}
}

func TestDoctorPreservesUnstructuredPlatformInfoContent(t *testing.T) {
	server := mcp.NewServer(&mcp.Implementation{Name: "erp-open-platform", Version: "1.2.3"}, nil)
	server.AddTool(&mcp.Tool{
		Name:        "get_platform_info",
		Description: "verify",
		InputSchema: json.RawMessage(`{"type":"object"}`),
	}, func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "platform-ready"}}}, nil
	})
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
	testServer := httptest.NewServer(handler)
	defer testServer.Close()

	result, err := New().Doctor(t.Context(), testServer.URL, "secret")
	if err != nil {
		t.Fatalf("Doctor() error = %v", err)
	}
	content, ok := result.Verification.Output.([]mcp.Content)
	if !ok || len(content) != 1 {
		t.Fatalf("Doctor() verification output = %#v, want MCP content", result.Verification.Output)
	}
	text, ok := content[0].(*mcp.TextContent)
	if !ok || text.Text != "platform-ready" {
		t.Fatalf("Doctor() verification content = %#v", content[0])
	}
}

func TestDoctorRedactsPlatformInfoFailures(t *testing.T) {
	const sentinel = "recognizable-platform-error-secret"
	for _, test := range []struct {
		name    string
		handler mcp.ToolHandler
	}{
		{
			name: "tool result error",
			handler: func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: sentinel}}}, nil
			},
		},
		{
			name: "protocol error",
			handler: func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				return nil, errors.New("protocol failed with " + sentinel)
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := mcp.NewServer(&mcp.Implementation{Name: "erp-open-platform", Version: "1.2.3"}, nil)
			server.AddTool(&mcp.Tool{
				Name:        "get_platform_info",
				Description: "verify",
				InputSchema: json.RawMessage(`{"type":"object"}`),
			}, test.handler)
			handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})
			testServer := httptest.NewServer(handler)
			defer testServer.Close()

			_, err := New().Doctor(t.Context(), testServer.URL, sentinel)
			if err == nil || !errors.Is(err, ErrUnavailable) {
				t.Fatalf("Doctor() error = %v, want unavailable", err)
			}
			if strings.Contains(err.Error(), sentinel) {
				t.Fatalf("Doctor() leaked platform error: %v", err)
			}
		})
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

func TestDoctorReportsUnauthorizedWithoutLeakingKey(t *testing.T) {
	const secret = "recognizable-unauthorized-key"
	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "unauthorized with "+request.Header.Get("Authorization"), http.StatusUnauthorized)
	}))
	defer testServer.Close()

	_, err := New().Doctor(t.Context(), testServer.URL, secret)
	if err == nil || !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("Doctor() error = %v, want unauthorized", err)
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("Doctor() leaked key: %v", err)
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
