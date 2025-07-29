package services

import (
	"testing"
)

func TestParseLanguageStats_ValidJSON(t *testing.T) {
	raw := `{"Go":70,"JavaScript":20,"TypeScript":10}`
	expected := map[string]int64{"Go": 70, "JavaScript": 20, "TypeScript": 10}
	parsed, err := parseLanguageStats(raw)
	if err != nil {
		t.Fatalf("Unexpected error parsing valid JSON: %v", err)
	}
	if len(parsed) != len(expected) {
		t.Fatalf("Parsed map length = %d; want %d", len(parsed), len(expected))
	}
	for lang, expVal := range expected {
		if val, ok := parsed[lang]; !ok || val != expVal {
			t.Errorf("parsed[%q] = %d; want %d", lang, val, expVal)
		}
	}
}

func TestParseLanguageStats_InvalidJSON(t *testing.T) {
	raw := `not a json`
	if _, err := parseLanguageStats(raw); err == nil {
		t.Fatal("Expected error parsing invalid JSON, got nil")
	}
}

func TestParseLanguageStats_JSONWithFloatValue(t *testing.T) {
	raw := `{"Go":70.5}`
	if _, err := parseLanguageStats(raw); err == nil {
		t.Fatal("Expected error parsing JSON with float values into map[string]int64, got nil")
	}
}
