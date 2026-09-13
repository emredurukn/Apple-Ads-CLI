package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/olekukonko/tablewriter"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatCSV   Format = "csv"
)

// IsTTY returns true if standard output is connected to an interactive terminal.
func IsTTY() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// ResolveFormat determines the format to use given the user flag or environmental default.
func ResolveFormat(requested string) Format {
	req := strings.ToLower(strings.TrimSpace(requested))
	switch req {
	case "table":
		return FormatTable
	case "json":
		return FormatJSON
	case "csv":
		return FormatCSV
	default:
		if IsTTY() {
			return FormatTable
		}
		return FormatJSON
	}
}

// Render outputs data in the requested format.
func Render(w io.Writer, format Format, data any, headers []string, rows [][]string) error {
	switch format {
	case FormatJSON:
		return RenderJSON(w, data, true)
	case FormatCSV:
		return RenderCSV(w, headers, rows)
	case FormatTable:
		fallthrough
	default:
		return RenderTable(w, headers, rows)
	}
}

// RenderJSON writes indented JSON representation of data.
func RenderJSON(w io.Writer, data any, pretty bool) error {
	var bytes []byte
	var err error
	if pretty {
		bytes, err = json.MarshalIndent(data, "", "  ")
	} else {
		bytes, err = json.Marshal(data)
	}
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	_, err = fmt.Fprintln(w, string(bytes))
	return err
}

// RenderTable formats and renders an ASCII table to writer.
func RenderTable(w io.Writer, headers []string, rows [][]string) error {
	var opts []tablewriter.Option
	if len(headers) > 0 {
		opts = append(opts, tablewriter.WithHeader(headers))
	}

	table := tablewriter.NewTable(w, opts...)
	if len(rows) > 0 {
		if err := table.Bulk(rows); err != nil {
			return fmt.Errorf("failed to append rows to table: %w", err)
		}
	}

	return table.Render()
}

// RenderCSV writes header and rows in comma-separated values format.
func RenderCSV(w io.Writer, headers []string, rows [][]string) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	if len(headers) > 0 {
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("failed to write CSV header: %w", err)
		}
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}
