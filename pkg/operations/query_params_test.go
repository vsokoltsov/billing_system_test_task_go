package operations

import (
	"net/url"
	"strings"
	"testing"
)

// Test success query params parsing
func TestSuccessQueryParamsParser(t *testing.T) {
	params := make(url.Values)
	params.Set("format", "json")
	params.Set("page", "1")
	params.Set("per_page", "10")
	params.Set("date", "2020-01-01")
	qpr := QueryParamsReader{}
	queryParams, err := qpr.Parse(params)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if queryParams.format != "json" {
		t.Errorf("Formats mismatch")
	}

	if queryParams.listParams.page != 1 {
		t.Errorf("Page mismatch")
	}

	if queryParams.listParams.perPage != 10 {
		t.Errorf("Per page mismatch")
	}

	if queryParams.listParams.date != "2020-01-01" {
		t.Errorf("Date mismatch")
	}
}

// Test success query params parsing with empty format
func TestSuccessQueryParamsParserEmptyFormat(t *testing.T) {
	params := make(url.Values)
	params.Set("page", "1")
	params.Set("per_page", "10")
	qpr := QueryParamsReader{}
	queryParams, err := qpr.Parse(params)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if queryParams.format != "json" {
		t.Errorf("Formats mismatch")
	}

	if queryParams.listParams.page != 1 {
		t.Errorf("Page mismatch")
	}

	if queryParams.listParams.perPage != 10 {
		t.Errorf("Per page mismatch")
	}
}

// Test failed query params parsing with wrong page type
func TestFailedQueryParamsParserWrongPageFormat(t *testing.T) {
	params := make(url.Values)
	params.Set("page", "awdaw")
	params.Set("per_page", "10")
	qpr := QueryParamsReader{}
	_, err := qpr.Parse(params)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "error of 'page' attribute converting") {
		t.Errorf("Wrong message in error")
	}
}

// Test failed query params parsing with wrong per_page type
func TestFailedQueryParamsParserWrongPerPageFormat(t *testing.T) {
	params := make(url.Values)
	params.Set("page", "1")
	params.Set("per_page", "awdaw")
	qpr := QueryParamsReader{}
	_, err := qpr.Parse(params)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "error of 'per_page' attribute converting") {
		t.Errorf("Wrong message in error")
	}
}

// Benchmark parameters parsing
func BenchmarkParse(b *testing.B) {
	params := make(url.Values)
	params.Set("format", "json")
	params.Set("page", "1")
	params.Set("per_page", "10")
	params.Set("date", "2020-01-01")
	qpr := QueryParamsReader{}

	for i := 0; i < b.N; i++ {
		_, _ = qpr.Parse(params)
	}
}
