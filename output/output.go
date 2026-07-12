// Package output provides consistent machine- and human-readable CLI rendering.
package output

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

var ErrOutput = errors.New("output error")

const (
	FormatTable = "table"
	FormatText  = "text"
	FormatJSON  = "json"
	FormatJSONL = "jsonl"
	FormatCSV   = "csv"
)

// NormaliseFormat makes supported format names lowercase and defaults empty to table.
func NormaliseFormat(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", FormatTable:
		return FormatTable
	case FormatText:
		return FormatText
	case FormatJSON:
		return FormatJSON
	case FormatJSONL:
		return FormatJSONL
	case FormatCSV:
		return FormatCSV
	default:
		return value
	}
}

func ValidateSnapshotFormat(value string) error {
	switch NormaliseFormat(value) {
	case FormatTable, FormatText, FormatJSON, FormatCSV:
		return nil
	case FormatJSONL:
		return fmt.Errorf("snapshot commands produce one result; use --format json instead of --format jsonl")
	default:
		return fmt.Errorf("invalid snapshot output format %q; expected table, text, json, or csv", value)
	}
}

func ValidateStreamFormat(value string) error {
	switch NormaliseFormat(value) {
	case FormatText, FormatJSONL, FormatCSV:
		return nil
	case FormatJSON:
		return fmt.Errorf("stream commands use line-delimited output; use --format jsonl instead of --format json")
	case FormatTable:
		return fmt.Errorf("stream commands do not support table output; use text, jsonl, or csv")
	default:
		return fmt.Errorf("invalid stream output format %q; expected text, jsonl, or csv", value)
	}
}

func WriteJSON(w io.Writer, value any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(value); err != nil {
		return wrap(err)
	}
	return nil
}

func WriteJSONLine(w io.Writer, value any) error {
	if err := json.NewEncoder(w).Encode(value); err != nil {
		return wrap(err)
	}
	return nil
}

func WriteText(w io.Writer, value any) error {
	if _, err := fmt.Fprintln(w, value); err != nil {
		return wrap(err)
	}
	return nil
}

func WriteTable(w io.Writer, headers []string, rows [][]string) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if len(headers) > 0 {
		if _, err := fmt.Fprintln(tw, strings.Join(headers, "\t")); err != nil {
			return wrap(err)
		}
	}
	for _, row := range rows {
		if _, err := fmt.Fprintln(tw, strings.Join(row, "\t")); err != nil {
			return wrap(err)
		}
	}
	if err := tw.Flush(); err != nil {
		return wrap(err)
	}
	return nil
}

func WriteCSV(w io.Writer, headers []string, rows [][]string) error {
	cw := csv.NewWriter(w)
	if len(headers) > 0 {
		if err := cw.Write(headers); err != nil {
			return wrap(err)
		}
	}
	for _, row := range rows {
		if err := cw.Write(row); err != nil {
			return wrap(err)
		}
	}
	cw.Flush()
	if err := cw.Error(); err != nil {
		return wrap(err)
	}
	return nil
}

func WriteCSVRows(w io.Writer, rows [][]string) error { return WriteCSV(w, nil, rows) }

// Stream writes repeated results using a stream-safe format.
type Stream struct {
	w      io.Writer
	format string
	csv    *csv.Writer
}

// NewStream initializes a stream. CSV headers are written exactly once.
func NewStream(w io.Writer, format string, headers []string) (*Stream, error) {
	format = NormaliseFormat(format)
	if err := ValidateStreamFormat(format); err != nil {
		return nil, err
	}
	s := &Stream{w: w, format: format}
	if format == FormatCSV {
		s.csv = csv.NewWriter(w)
		if len(headers) > 0 {
			if err := s.csv.Write(headers); err != nil {
				return nil, wrap(err)
			}
			s.csv.Flush()
			if err := s.csv.Error(); err != nil {
				return nil, wrap(err)
			}
		}
	}
	return s, nil
}

// Write emits one stream record. row is used for text and CSV; value is used for JSONL.
func (s *Stream) Write(row []string, value any) error {
	switch s.format {
	case FormatText:
		return WriteText(s.w, strings.Join(row, "\t"))
	case FormatJSONL:
		return WriteJSONLine(s.w, value)
	case FormatCSV:
		if err := s.csv.Write(row); err != nil {
			return wrap(err)
		}
		s.csv.Flush()
		if err := s.csv.Error(); err != nil {
			return wrap(err)
		}
		return nil
	default:
		return fmt.Errorf("%w: invalid stream format %q", ErrOutput, s.format)
	}
}

func wrap(err error) error { return fmt.Errorf("%w: %v", ErrOutput, err) }
