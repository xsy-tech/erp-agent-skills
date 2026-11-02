package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"strings"
)

var errAmbiguousRedactedJSON = errors.New("doctor output contains ambiguous redacted object keys")

func (a *App) runDoctor(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("doctor requires an agent")
	}
	agent := args[0]
	if agent != "codex" && agent != "claude-code" && agent != "workbuddy" {
		return errors.New("doctor supports codex, claude-code, or workbuddy")
	}
	flags := flag.NewFlagSet("doctor "+agent, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	endpoint := flags.String("url", "", "MCP endpoint")
	keyStdin := flags.Bool("key-stdin", false, "read API key from stdin")
	if err := flags.Parse(args[1:]); err != nil || flags.NArg() != 0 || strings.TrimSpace(*endpoint) == "" {
		return errors.New("doctor requires --url")
	}
	var secret string
	var err error
	if agent == "workbuddy" {
		if !*keyStdin {
			return errors.New("WorkBuddy doctor requires --key-stdin")
		}
		secret, err = readSecret(a.In)
	} else {
		if *keyStdin {
			return errors.New("local agent doctor reads the stored OS credential")
		}
		secret, err = a.Credentials.Get(agent, *endpoint)
	}
	if err != nil {
		return err
	}
	result, err := a.Doctor.Doctor(ctx, *endpoint, secret)
	if err != nil {
		return redactError(err, secret)
	}
	return encodeRedactedJSON(a.Out, result, secret)
}

func encodeRedactedJSON(writer io.Writer, value any, secret string) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return redactError(err, secret)
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	var payload any
	if err := decoder.Decode(&payload); err != nil {
		return redactError(err, secret)
	}
	redacted, err := redactJSONValue(payload, secret)
	if err != nil {
		return redactError(err, secret)
	}
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return redactError(encoder.Encode(redacted), secret)
}

func redactJSONValue(value any, secret string) (any, error) {
	if secret == "" {
		return value, nil
	}
	switch typed := value.(type) {
	case string:
		return strings.ReplaceAll(typed, secret, "[REDACTED]"), nil
	case []any:
		for index := range typed {
			redacted, err := redactJSONValue(typed[index], secret)
			if err != nil {
				return nil, err
			}
			typed[index] = redacted
		}
		return typed, nil
	case map[string]any:
		redacted := make(map[string]any, len(typed))
		for key, item := range typed {
			redactedKey := strings.ReplaceAll(key, secret, "[REDACTED]")
			if _, exists := redacted[redactedKey]; exists {
				return nil, errAmbiguousRedactedJSON
			}
			redactedItem, err := redactJSONValue(item, secret)
			if err != nil {
				return nil, err
			}
			redacted[redactedKey] = redactedItem
		}
		return redacted, nil
	default:
		return value, nil
	}
}
