package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/go-saas/erp-agent-skill/internal/mcpclient"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type fakeCredentials struct {
	values map[string]string
}

func (s *fakeCredentials) Set(agent, endpoint, secret string) error {
	if s.values == nil {
		s.values = make(map[string]string)
	}
	s.values[agent+"|"+endpoint] = secret
	return nil
}

func (s *fakeCredentials) Get(agent, endpoint string) (string, error) {
	value, ok := s.values[agent+"|"+endpoint]
	if !ok {
		return "", errors.New("credential missing")
	}
	return value, nil
}

func (s *fakeCredentials) Delete(agent, endpoint string) error {
	delete(s.values, agent+"|"+endpoint)
	return nil
}

type fakeProfiles struct {
	values map[string]string
}

func (s *fakeProfiles) Set(agent, endpoint string) error {
	if s.values == nil {
		s.values = make(map[string]string)
	}
	s.values[agent] = endpoint
	return nil
}

func (s *fakeProfiles) Get(agent string) (string, error) {
	value, ok := s.values[agent]
	if !ok {
		return "", errors.New("profile missing")
	}
	return value, nil
}

func (s *fakeProfiles) Delete(agent string) error {
	delete(s.values, agent)
	return nil
}

type fakeDoctor struct {
	result mcpclient.DoctorResult
	err    error
}

func (d fakeDoctor) Doctor(context.Context, string, string) (mcpclient.DoctorResult, error) {
	return d.result, d.err
}

type fakeCommands struct {
	name string
	args []string
}

func (r *fakeCommands) Run(_ context.Context, name string, args ...string) error {
	r.name = name
	r.args = append([]string(nil), args...)
	return nil
}

type launchCall struct {
	command string
	args    []string
	envKey  string
	secret  string
}

func newTestApp(t *testing.T) (*App, *bytes.Buffer, *fakeCredentials, *fakeProfiles, *fakeCommands, *[]launchCall) {
	t.Helper()
	output := &bytes.Buffer{}
	credentials := &fakeCredentials{}
	profiles := &fakeProfiles{}
	commands := &fakeCommands{}
	launches := &[]launchCall{}
	home := t.TempDir()
	app := &App{
		Version:     "0.1.0",
		In:          strings.NewReader(""),
		Out:         output,
		Err:         &bytes.Buffer{},
		HomeDir:     home,
		CodexHome:   filepath.Join(home, ".codex"),
		Credentials: credentials,
		Profiles:    profiles,
		Doctor: fakeDoctor{result: mcpclient.DoctorResult{
			Status:        mcpclient.DoctorStatusVerified,
			ServerName:    "erp-open-platform",
			ServerVersion: "1.0.0",
			Instructions:  "instructions",
			Tools:         []string{"get_platform_info", "search_users"},
			Verification:  &mcpclient.VerificationResult{Tool: "get_platform_info", Output: map[string]any{"ok": true}},
		}},
		Commands: commands,
		Launch: func(_ context.Context, command string, args []string, envKey, secret string) error {
			*launches = append(*launches, launchCall{command: command, args: append([]string(nil), args...), envKey: envKey, secret: secret})
			return nil
		},
	}
	return app, output, credentials, profiles, commands, launches
}

func TestInstallCodexValidatesStoresAndRegisters(t *testing.T) {
	app, _, credentials, profiles, commands, _ := newTestApp(t)
	app.In = strings.NewReader("personal-secret\n")
	endpoint := "https://erp.example.com/mcp"
	if err := app.Run(t.Context(), []string{"install", "codex", "--url", endpoint, "--key-stdin"}); err != nil {
		t.Fatal(err)
	}
	if credentials.values["codex|"+endpoint] != "personal-secret" {
		t.Fatal("credential was not stored")
	}
	if profiles.values["codex"] != endpoint {
		t.Fatal("endpoint profile was not stored")
	}
	wantArgs := []string{"mcp", "add", "erp-open", "--url", endpoint, "--bearer-token-env-var", "ERP_OPEN_CODEX_KEY"}
	if commands.name != "codex" || !reflect.DeepEqual(commands.args, wantArgs) {
		t.Fatalf("registration command = %s %v", commands.name, commands.args)
	}
}

func TestDoctorAndLaunchUseStoredCredential(t *testing.T) {
	app, output, credentials, profiles, _, launches := newTestApp(t)
	endpoint := "https://erp.example.com/mcp"
	profiles.values = map[string]string{"codex": endpoint}
	credentials.values = map[string]string{"codex|" + endpoint: "personal-secret"}
	if err := app.Run(t.Context(), []string{"doctor", "codex", "--url", endpoint}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "search_users") {
		t.Fatalf("doctor output = %q", output.String())
	}
	if err := app.Run(t.Context(), []string{"launch", "codex", "--", "--model", "gpt"}); err != nil {
		t.Fatal(err)
	}
	if len(*launches) != 1 || (*launches)[0].secret != "personal-secret" || (*launches)[0].envKey != "ERP_OPEN_CODEX_KEY" {
		t.Fatalf("launches = %#v", *launches)
	}
	if !reflect.DeepEqual((*launches)[0].args, []string{"--model", "gpt"}) {
		t.Fatalf("launch args = %v", (*launches)[0].args)
	}
}

func TestDoctorJSONStatusesDoNotLeakCredential(t *testing.T) {
	const sentinel = "recognizable-sentinel-key"
	endpoint := "https://erp.example.com/mcp"
	for _, test := range []struct {
		name             string
		result           mcpclient.DoctorResult
		wantStatus       mcpclient.DoctorStatus
		wantVerification bool
		wantValue        string
	}{
		{
			name: "verified structured output",
			result: mcpclient.DoctorResult{
				Status:        mcpclient.DoctorStatusVerified,
				ServerName:    "erp-open-platform",
				ServerVersion: "1.2.3",
				Tools:         []string{"get_platform_info"},
				Verification: &mcpclient.VerificationResult{Tool: "get_platform_info", Output: map[string]any{
					"tenant":                 "demo",
					"credential":             sentinel,
					"credential-" + sentinel: "redacted key",
				}},
			},
			wantStatus:       mcpclient.DoctorStatusVerified,
			wantVerification: true,
			wantValue:        "demo",
		},
		{
			name: "verified unstructured output",
			result: mcpclient.DoctorResult{
				Status:        mcpclient.DoctorStatusVerified,
				ServerName:    "erp-open-platform",
				ServerVersion: "1.2.3",
				Tools:         []string{"get_platform_info"},
				Verification: &mcpclient.VerificationResult{
					Tool: "get_platform_info",
					Output: []mcp.Content{
						&mcp.TextContent{Text: "platform-ready " + sentinel},
					},
				},
			},
			wantStatus:       mcpclient.DoctorStatusVerified,
			wantVerification: true,
			wantValue:        "platform-ready",
		},
		{
			name: "authorization required",
			result: mcpclient.DoctorResult{
				Status:        mcpclient.DoctorStatusAuthorizationRequired,
				ServerName:    "erp-open-platform",
				ServerVersion: "1.2.3",
				Tools:         []string{},
			},
			wantStatus: mcpclient.DoctorStatusAuthorizationRequired,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			app, stdout, credentials, _, _, _ := newTestApp(t)
			stderr := app.Err.(*bytes.Buffer)
			credentials.values = map[string]string{"codex|" + endpoint: sentinel}
			app.Doctor = fakeDoctor{result: test.result}

			runErr := app.Run(t.Context(), []string{"doctor", "codex", "--url", endpoint})
			if runErr != nil {
				t.Fatalf("doctor error = %v", runErr)
			}
			var payload map[string]json.RawMessage
			if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
				t.Fatalf("doctor JSON = %q: %v", stdout.String(), err)
			}
			var status mcpclient.DoctorStatus
			if err := json.Unmarshal(payload["status"], &status); err != nil {
				t.Fatalf("doctor status JSON = %s: %v", payload["status"], err)
			}
			if status != test.wantStatus {
				t.Fatalf("doctor status = %q, want %q", status, test.wantStatus)
			}
			for _, key := range []string{"server_name", "server_version", "tools"} {
				if _, ok := payload[key]; !ok {
					t.Fatalf("doctor JSON missing %q: %s", key, stdout.String())
				}
			}
			_, hasVerification := payload["verification"]
			if hasVerification != test.wantVerification {
				t.Fatalf("doctor verification presence = %v, want %v: %s", hasVerification, test.wantVerification, stdout.String())
			}
			if test.wantValue != "" && !strings.Contains(stdout.String(), test.wantValue) {
				t.Fatalf("doctor JSON lost non-secret value %q: %s", test.wantValue, stdout.String())
			}
			if strings.Contains(stdout.String(), sentinel) || strings.Contains(stderr.String(), sentinel) || (runErr != nil && strings.Contains(runErr.Error(), sentinel)) {
				t.Fatalf("doctor leaked credential: stdout=%q stderr=%q error=%v", stdout.String(), stderr.String(), runErr)
			}
		})
	}

	t.Run("error path", func(t *testing.T) {
		app, stdout, credentials, _, _, _ := newTestApp(t)
		stderr := app.Err.(*bytes.Buffer)
		credentials.values = map[string]string{"codex|" + endpoint: sentinel}
		app.Doctor = fakeDoctor{err: errors.New("doctor failed with " + sentinel)}

		runErr := app.Run(t.Context(), []string{"doctor", "codex", "--url", endpoint})
		if runErr == nil {
			t.Fatal("doctor error = nil, want error")
		}
		if strings.Contains(stdout.String(), sentinel) || strings.Contains(stderr.String(), sentinel) || strings.Contains(runErr.Error(), sentinel) {
			t.Fatalf("doctor leaked credential: stdout=%q stderr=%q error=%q", stdout.String(), stderr.String(), runErr)
		}
	})
}

func TestCLIRejectsLiteralKeyAndRedactsFailures(t *testing.T) {
	app, _, _, _, _, _ := newTestApp(t)
	if err := app.Run(t.Context(), []string{"install", "codex", "--url", "https://erp.example.com/mcp", "--key", "literal-secret"}); err == nil || strings.Contains(err.Error(), "literal-secret") {
		t.Fatalf("literal key rejection error = %v", err)
	}
	const secret = "recognizable-secret-value"
	app.In = strings.NewReader(secret)
	app.Doctor = fakeDoctor{err: errors.New("doctor failed with " + secret)}
	err := app.Run(t.Context(), []string{"install", "codex", "--url", "https://erp.example.com/mcp", "--key-stdin"})
	if err == nil || strings.Contains(err.Error(), secret) {
		t.Fatalf("redacted error = %v", err)
	}
}

func TestCLIRejectsUnknownAgentAndMissingURL(t *testing.T) {
	app, _, _, _, _, _ := newTestApp(t)
	for _, args := range [][]string{
		{"install", "unknown", "--url", "https://erp.example.com/mcp", "--key-stdin"},
		{"install", "codex", "--key-stdin"},
	} {
		if err := app.Run(t.Context(), args); err == nil {
			t.Fatalf("Run(%v) error = nil", args)
		}
	}
}
