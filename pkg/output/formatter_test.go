package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestResolveFormat(t *testing.T) {
	if f := ResolveFormat("json"); f != FormatJSON {
		t.Errorf("expected FormatJSON, got %v", f)
	}
	if f := ResolveFormat("table"); f != FormatTable {
		t.Errorf("expected FormatTable, got %v", f)
	}
	if f := ResolveFormat("csv"); f != FormatCSV {
		t.Errorf("expected FormatCSV, got %v", f)
	}
}

func TestRenderJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"name": "Apple Ads Campaign"}

	err := RenderJSON(&buf, data, false)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	expected := `{"name":"Apple Ads Campaign"}`
	if !strings.Contains(buf.String(), expected) {
		t.Errorf("expected output to contain %s, got %s", expected, buf.String())
	}
}

func TestRenderCSV(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"ID", "NAME"}
	rows := [][]string{
		{"101", "Summer Promo"},
		{"102", "Winter Launch"},
	}

	err := RenderCSV(&buf, headers, rows)
	if err != nil {
		t.Fatalf("RenderCSV failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "ID,NAME") {
		t.Errorf("expected CSV header, got %s", out)
	}
	if !strings.Contains(out, "101,Summer Promo") {
		t.Errorf("expected CSV row, got %s", out)
	}
}

func TestRenderTable(t *testing.T) {
	var buf bytes.Buffer
	headers := []string{"ID", "NAME"}
	rows := [][]string{
		{"101", "Summer Promo"},
	}

	err := RenderTable(&buf, headers, rows)
	if err != nil {
		t.Fatalf("RenderTable failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Summer Promo") {
		t.Errorf("expected table to contain 'Summer Promo', got %s", out)
	}
}
