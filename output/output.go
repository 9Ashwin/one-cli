// Copyright (c) 2026 9Ashwin. SPDX-License-Identifier: MIT

// Package output provides the output formatting abstraction for one-cli.
//
// Implementations render command results as JSON, pretty text, tables, etc.
package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// Format names supported output formats.
type Format string

const (
	FormatJSON   Format = "json"
	FormatPretty Format = "pretty"
	FormatTable  Format = "table"
	FormatNDJSON Format = "ndjson"
	FormatCSV    Format = "csv"
)

// Formatter renders structured data to a writer.
type Formatter interface {
	// Format writes the given value using the configured format.
	Format(w io.Writer, v any) error
}

// JSONFormatter writes compact JSON.
type JSONFormatter struct{}

// Format implements Formatter.
func (JSONFormatter) Format(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// PrettyFormatter writes a human-readable representation.
type PrettyFormatter struct{}

// Format implements Formatter.
func (PrettyFormatter) Format(w io.Writer, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(b))
	return err
}

// New returns a Formatter for the given format.
func New(f Format) Formatter {
	switch f {
	case FormatPretty:
		return PrettyFormatter{}
	case FormatJSON, FormatTable, FormatNDJSON, FormatCSV:
		// Table/NDJSON/CSV fall back to JSON in this scaffold; expanded later.
		return JSONFormatter{}
	default:
		return JSONFormatter{}
	}
}
