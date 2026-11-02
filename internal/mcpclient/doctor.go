package mcpclient

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	bundle "github.com/go-saas/erp-agent-skill"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	ErrUnauthorized = errors.New("MCP endpoint rejected the API key")
	ErrNoTools      = errors.New("MCP endpoint returned no visible tools")
	ErrUnavailable  = errors.New("MCP endpoint is unavailable")
)

type DoctorStatus string

const (
	DoctorStatusVerified              DoctorStatus = "VERIFIED"
	DoctorStatusAuthorizationRequired DoctorStatus = "INSTALLED_AUTHORIZATION_REQUIRED"
)

type VerificationResult struct {
	Tool   string `json:"tool"`
	Output any    `json:"output"`
}

type DoctorResult struct {
	Status        DoctorStatus        `json:"status"`
	ServerName    string              `json:"server_name"`
	ServerVersion string              `json:"server_version"`
	Instructions  string              `json:"instructions,omitempty"`
	Tools         []string            `json:"tools"`
	Verification  *VerificationResult `json:"verification,omitempty"`
}

func (c *Client) Doctor(ctx context.Context, endpoint, key string) (DoctorResult, error) {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return DoctorResult{}, errors.New("MCP endpoint must be an absolute URL")
	}
	if strings.TrimSpace(key) == "" {
		return DoctorResult{}, errors.New("MCP API key is required")
	}
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	doctorContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client := mcp.NewClient(&mcp.Implementation{Name: "erp-agent-skill", Version: bundle.Version}, nil)
	session, err := client.Connect(doctorContext, &mcp.StreamableClientTransport{
		Endpoint: endpoint, HTTPClient: c.authenticatedHTTPClient(key), DisableStandaloneSSE: true, MaxRetries: -1,
	}, nil)
	if err != nil {
		return DoctorResult{}, doctorError(doctorContext, parsed.Host, err)
	}
	defer session.Close()

	initialize := session.InitializeResult()
	if initialize == nil {
		return DoctorResult{}, fmt.Errorf("MCP endpoint %s returned no initialize result", parsed.Host)
	}
	result := DoctorResult{
		ServerName:    initialize.ServerInfo.Name,
		ServerVersion: initialize.ServerInfo.Version,
		Instructions:  initialize.Instructions,
		Tools:         []string{},
	}
	hasPlatformInfo := false
	seenCursors := make(map[string]struct{})
	seenTools := make(map[string]struct{})
	params := &mcp.ListToolsParams{}
	for {
		page, listErr := session.ListTools(doctorContext, params)
		if listErr != nil {
			return DoctorResult{}, doctorError(doctorContext, parsed.Host, listErr)
		}
		for _, tool := range page.Tools {
			if _, exists := seenTools[tool.Name]; exists {
				continue
			}
			seenTools[tool.Name] = struct{}{}
			result.Tools = append(result.Tools, tool.Name)
			if tool.Name == "get_platform_info" {
				hasPlatformInfo = true
			}
		}
		if page.NextCursor == "" {
			break
		}
		if _, exists := seenCursors[page.NextCursor]; exists {
			return DoctorResult{}, doctorError(doctorContext, parsed.Host, errors.New("tools list returned a repeated cursor"))
		}
		seenCursors[page.NextCursor] = struct{}{}
		params.Cursor = page.NextCursor
	}
	sort.Strings(result.Tools)
	if !hasPlatformInfo {
		result.Status = DoctorStatusAuthorizationRequired
		return result, nil
	}
	verification, callErr := session.CallTool(doctorContext, &mcp.CallToolParams{
		Name:      "get_platform_info",
		Arguments: map[string]any{},
	})
	if callErr != nil {
		return DoctorResult{}, doctorError(doctorContext, parsed.Host, callErr)
	}
	if verification == nil || verification.IsError {
		return DoctorResult{}, doctorError(doctorContext, parsed.Host, errors.New("platform verification failed"))
	}
	output := verification.StructuredContent
	if output == nil {
		output = verification.Content
	}
	result.Status = DoctorStatusVerified
	result.Verification = &VerificationResult{Tool: "get_platform_info", Output: output}
	return result, nil
}

func doctorError(ctx context.Context, host string, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "401") || strings.Contains(message, "unauthorized") {
		return fmt.Errorf("MCP endpoint %s: %w", host, ErrUnauthorized)
	}
	return fmt.Errorf("MCP endpoint %s: %w", host, ErrUnavailable)
}
