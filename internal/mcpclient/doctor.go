package mcpclient

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	ErrUnauthorized = errors.New("MCP endpoint rejected the API key")
	ErrNoTools      = errors.New("MCP endpoint returned no visible tools")
	ErrUnavailable  = errors.New("MCP endpoint is unavailable")
)

type DoctorResult struct {
	ServerName    string
	ServerVersion string
	Instructions  string
	Tools         []string
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

	client := mcp.NewClient(&mcp.Implementation{Name: "erp-agent-skill", Version: "0.1.0"}, nil)
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
	}
	params := &mcp.ListToolsParams{}
	for {
		page, listErr := session.ListTools(doctorContext, params)
		if listErr != nil {
			return DoctorResult{}, doctorError(doctorContext, parsed.Host, listErr)
		}
		for _, tool := range page.Tools {
			result.Tools = append(result.Tools, tool.Name)
		}
		if page.NextCursor == "" {
			break
		}
		params.Cursor = page.NextCursor
	}
	if len(result.Tools) == 0 {
		return DoctorResult{}, fmt.Errorf("MCP endpoint %s: %w", parsed.Host, ErrNoTools)
	}
	sort.Strings(result.Tools)
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
