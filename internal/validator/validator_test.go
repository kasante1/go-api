package validator

import (
	"regexp"
	"testing"
)

func TestNew(t *testing.T) {
	v := New()
	if v.Errors == nil {
		t.Error("Expected Errors map to be initialized, got nil")
	}
	if len(v.Errors) != 0 {
		t.Error("Expected Errors map to be empty, got non-empty")
	}
}

func TestValidator_Valid(t *testing.T) {
	v := New()
	if !v.Valid() {
		t.Error("Expected Valid to be true for a new Validator with no errors")
	}

	v.AddError("test", "test error")
	if v.Valid() {
		t.Error("Expected Valid to be false when errors are present")
	}
}

func TestValidator_AddError(t *testing.T) {
	v := New()
	v.AddError("test", "test error")
	if v.Errors["test"] != "test error" {
		t.Errorf("Expected error message 'test error', got %v", v.Errors["test"])
	}

	v.AddError("test", "another error")
	if v.Errors["test"] != "test error" {
		t.Errorf("Expected error message to remain 'test error', got %v", v.Errors["test"])
	}
}

func TestValidator_Check(t *testing.T) {
	v := New()
	v.Check(false, "test", "error message")
	if v.Errors["test"] != "error message" {
		t.Errorf("Expected error message 'error message', got %v", v.Errors["test"])
	}

	v.Check(true, "test", "another message")
	if v.Errors["test"] != "error message" {
		t.Errorf("Expected error message to remain 'error message', got %v", v.Errors["test"])
	}
}

func TestIn(t *testing.T) {
	if !In("a", "a", "b", "c") {
		t.Error("Expected In to return true for value 'a' in the list")
	}

	if In("d", "a", "b", "c") {
		t.Error("Expected In to return false for value 'd' not in the list")
	}
}

func TestMatches(t *testing.T) {
	rx := regexp.MustCompile("^[a-z]+$")

	if !Matches("test", rx) {
		t.Error("Expected Matches to return true for matching pattern")
	}

	if Matches("Test123", rx) {
		t.Error("Expected Matches to return false for non-matching pattern")
	}
}

func TestUnique(t *testing.T) {
	values := []string{"a", "b", "c"}
	if !Unique(values) {
		t.Error("Expected Unique to return true for a list of unique values")
	}

	values = []string{"a", "b", "a"}
	if Unique(values) {
		t.Error("Expected Unique to return false for a list with duplicates")
	}
}
