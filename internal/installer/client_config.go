package installer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"
)

var ErrConfigCredentialNotFound = errors.New("no persistent ERP credential for this endpoint")
var errConfigUnchanged = errors.New("MCP configuration unchanged")
var configUpdateMu sync.Mutex

// ClientConfig edits only the user-scoped erp-open connection. Secrets never
// travel through a client CLI's command-line arguments.
type ClientConfig struct {
	HomeDir         string
	CodexHome       string
	ClaudeConfigDir string
}

func (c ClientConfig) Path(agent string) (string, error) {
	if c.HomeDir == "" {
		return "", errors.New("user home directory is required")
	}
	switch agent {
	case "codex":
		home := c.CodexHome
		if home == "" {
			home = filepath.Join(c.HomeDir, ".codex")
		}
		return filepath.Join(home, "config.toml"), nil
	case "claude-code":
		home := c.HomeDir
		if c.ClaudeConfigDir != "" {
			home = c.ClaudeConfigDir
		}
		return filepath.Join(home, ".claude.json"), nil
	default:
		return "", errors.New("unsupported MCP configuration client")
	}
}

func (c ClientConfig) Register(agent, endpoint, secret string) error {
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil || parsed.Fragment != "" {
		return errors.New("invalid HTTP MCP endpoint")
	}
	if secret == "" || strings.ContainsAny(secret, "\x00\r\n") {
		return errors.New("invalid MCP credential")
	}
	return c.update(agent, func(servers map[string]any) error {
		server := map[string]any{"url": endpoint}
		headers := map[string]any{"Authorization": "Bearer " + secret}
		if agent == "codex" {
			server["http_headers"] = headers
		} else {
			server["type"] = "http"
			server["headers"] = headers
		}
		servers["erp-open"] = server
		return nil
	})
}

func (c ClientConfig) Unregister(agent, endpoint string) error {
	return c.update(agent, func(servers map[string]any) error {
		server, exists := servers["erp-open"]
		if !exists {
			return errConfigUnchanged
		}
		entry, ok := server.(map[string]any)
		if !ok || entry["url"] != endpoint {
			return errors.New("ERP MCP endpoint changed; refusing to remove another connection")
		}
		delete(servers, "erp-open")
		return nil
	})
}

func (c ClientConfig) Credential(agent, endpoint string) (string, error) {
	path, err := c.Path(agent)
	if err != nil {
		return "", err
	}
	data, err := readClientConfig(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", ErrConfigCredentialNotFound
	}
	if err != nil {
		return "", err
	}
	_, servers, err := decodeClientConfig(agent, data)
	if err != nil {
		return "", err
	}
	server, ok := servers["erp-open"].(map[string]any)
	if !ok || server["url"] != endpoint {
		return "", ErrConfigCredentialNotFound
	}
	headerKey := "headers"
	if agent == "codex" {
		headerKey = "http_headers"
	}
	headers, _ := server[headerKey].(map[string]any)
	for name, value := range headers {
		if strings.EqualFold(name, "Authorization") {
			text, _ := value.(string)
			if strings.HasPrefix(text, "Bearer ") && !strings.Contains(text, "${") && len(text) > len("Bearer ") {
				return strings.TrimPrefix(text, "Bearer "), nil
			}
		}
	}
	return "", ErrConfigCredentialNotFound
}

func (c ClientConfig) update(agent string, change func(map[string]any) error) error {
	configUpdateMu.Lock()
	defer configUpdateMu.Unlock()
	path, err := c.Path(agent)
	if err != nil {
		return err
	}
	previous, err := readClientConfig(path)
	existed := !errors.Is(err, os.ErrNotExist)
	if err != nil && existed {
		return err
	}
	config, servers, err := decodeClientConfig(agent, previous)
	if err != nil {
		return err
	}
	if err := change(servers); err != nil {
		if errors.Is(err, errConfigUnchanged) {
			return nil
		}
		return err
	}
	var data []byte
	if agent == "codex" {
		data, err = encodeCodexConfig(previous, config, servers)
	} else {
		data, err = json.MarshalIndent(config, "", "  ")
		data = append(data, '\n')
	}
	if err != nil {
		return errors.New("cannot encode user MCP configuration")
	}
	current, readErr := readClientConfig(path)
	if (readErr != nil && !errors.Is(readErr, os.ErrNotExist)) || existed != (readErr == nil) || !bytes.Equal(current, previous) {
		return errors.New("user MCP configuration changed during installation; retry")
	}
	return writeAtomic(path, data, 0600)
}

func readClientConfig(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 8*1024*1024 {
		return nil, errors.New("user MCP configuration must be a regular file below 8 MiB; symbolic links are not modified")
	}
	return os.ReadFile(path)
}

func decodeClientConfig(agent string, data []byte) (map[string]any, map[string]any, error) {
	config := make(map[string]any)
	if len(bytes.TrimSpace(data)) != 0 {
		var err error
		if agent == "codex" {
			err = toml.Unmarshal(data, &config)
		} else {
			decoder := json.NewDecoder(bytes.NewReader(data))
			decoder.UseNumber()
			err = decoder.Decode(&config)
			if err == nil {
				var trailing any
				if decoder.Decode(&trailing) != io.EOF {
					err = errors.New("trailing JSON data")
				}
			}
		}
		if err != nil || config == nil {
			// Parser diagnostics can contain an existing credential-bearing line.
			return nil, nil, errors.New("invalid user MCP configuration; existing file was not modified")
		}
	}
	key := "mcpServers"
	if agent == "codex" {
		key = "mcp_servers"
	}
	servers, ok := config[key].(map[string]any)
	if !ok {
		if _, exists := config[key]; exists {
			return nil, nil, errors.New("invalid MCP server map; existing file was not modified")
		}
		servers = make(map[string]any)
		config[key] = servers
	}
	return config, servers, nil
}

// Keep the bytes outside normal erp-open table definitions, including comments.
// For inline/dotted representations, re-encode only after semantic validation.
func encodeCodexConfig(previous []byte, config, servers map[string]any) ([]byte, error) {
	var parser unstable.Parser
	parser.Reset(previous)
	var result []byte
	start, end, copied := -1, -1, 0
	flush := func() {
		if start >= 0 {
			result = append(result, previous[copied:start]...)
			copied = end
		}
		start, end = -1, -1
	}
	lineEnd := func(offset int) int {
		if index := bytes.IndexByte(previous[offset:], '\n'); index >= 0 {
			return offset + index + 1
		}
		return len(previous)
	}
	for parser.NextExpression() {
		node := parser.Expression()
		if node.Kind == unstable.Table || node.Kind == unstable.ArrayTable {
			flush()
			var keys []string
			first := 0
			iterator := node.Key()
			for iterator.Next() {
				if len(keys) == 0 {
					first = int(iterator.Node().Raw.Offset)
				}
				keys = append(keys, string(iterator.Node().Data))
			}
			if len(keys) >= 2 && keys[0] == "mcp_servers" && keys[1] == "erp-open" {
				start = bytes.LastIndexByte(previous[:first], '\n') + 1
				end = lineEnd(first)
			}
		} else if start >= 0 && node.Kind == unstable.KeyValue {
			end = lineEnd(int(node.Raw.Offset + node.Raw.Length))
		}
	}
	if parser.Error() != nil {
		return nil, errors.New("invalid TOML configuration")
	}
	flush()
	result = append(result, previous[copied:]...)
	if server, ok := servers["erp-open"].(map[string]any); ok {
		url, _ := json.Marshal(server["url"])
		headers := server["http_headers"].(map[string]any)
		authorization, _ := json.Marshal(headers["Authorization"])
		result = append(result, []byte(fmt.Sprintf("\n[mcp_servers.erp-open]\nurl = %s\n[mcp_servers.erp-open.http_headers]\nAuthorization = %s\n", url, authorization))...)
	}
	var verified map[string]any
	if toml.Unmarshal(result, &verified) == nil && reflect.DeepEqual(verified, config) {
		return result, nil
	}
	return toml.Marshal(config)
}
