// Package output renders CLI results as tables, JSON, or YAML.
package output

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// Format is one of "table", "json", or "yaml".
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

// ParseFormat validates s. allowTable rejects "table" for commands that don't
// have a table form (e.g. events describe, events diff).
func ParseFormat(s string, allowTable bool) (Format, error) {
	switch Format(s) {
	case FormatTable:
		if !allowTable {
			return "", fmt.Errorf("--format=%s not supported here (use json or yaml)", s)
		}
		return FormatTable, nil
	case FormatJSON:
		return FormatJSON, nil
	case FormatYAML:
		return FormatYAML, nil
	default:
		return "", fmt.Errorf("--format=%s invalid (must be table, json, or yaml)", s)
	}
}

// JSON writes v as indented JSON followed by a newline.
func JSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// YAML writes v as YAML.
func YAML(w io.Writer, v any) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		return err
	}
	return enc.Close()
}
