package validate_test

import (
	"strings"
	"testing"

	"github.com/vignesh/livepoll/internal/validate"
)

func TestValidationResult(t *testing.T) {
	r := validate.New()
	if !r.OK() {
		t.Fatal("new validator should be OK")
	}

	r.Check(true, "field1", "should not error")
	if !r.OK() {
		t.Fatal("validator with true check should be OK")
	}

	r.Check(false, "field2", "failed condition")
	if r.OK() {
		t.Fatal("validator with false check should not be OK")
	}

	if len(r.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(r.Errors))
	}
	if r.Errors[0].Field != "field2" {
		t.Errorf("expected field2, got %s", r.Errors[0].Field)
	}

	if r.Err() == nil {
		t.Fatal("expected non-nil error when errors are present")
	}
}

func TestEmailValidation(t *testing.T) {
	// Valid emails
	validEmails := []string{
		"user@example.com",
		"User.Name+tag@example.co.uk",
		"  clean.me@example.com  ",
	}

	for _, raw := range validEmails {
		r := validate.New()
		cleaned := validate.Email(r, "email", raw)
		if !r.OK() {
			t.Errorf("expected %q to be valid, got: %v", raw, r.Errors)
		}
		if cleaned != strings.ToLower(strings.TrimSpace(raw)) {
			t.Errorf("expected lowercase cleaned email, got %s", cleaned)
		}
	}

	// Invalid emails
	invalidEmails := []string{
		"",
		"not-an-email",
		"@example.com",
		"user@",
		"Vignesh <vignesh@example.com>", // Display-name form rejected
	}

	for _, raw := range invalidEmails {
		r := validate.New()
		validate.Email(r, "email", raw)
		if r.OK() {
			t.Errorf("expected %q to fail validation", raw)
		}
	}
}

func TestPasswordValidation(t *testing.T) {
	// Valid password
	r := validate.New()
	validate.Password(r, "password", "validPass123")
	if !r.OK() {
		t.Fatalf("expected password to pass, got: %v", r.Errors)
	}

	// Empty password
	r = validate.New()
	validate.Password(r, "password", "")
	if r.OK() {
		t.Fatal("expected error for empty password")
	}

	// Too short (< 8 runes)
	r = validate.New()
	validate.Password(r, "password", "short")
	if r.OK() {
		t.Fatal("expected error for short password")
	}

	// Too long (> 72 bytes)
	longPass := strings.Repeat("A", validate.MaxPasswordBytes+1)
	r = validate.New()
	validate.Password(r, "password", longPass)
	if r.OK() {
		t.Fatal("expected error for password exceeding max bcrypt bytes")
	}
}

func TestTextValidationAndCleaning(t *testing.T) {
	r := validate.New()
	// Test rune count for non-Latin script
	tamilText := "வணக்கம்" // 7 runes
	cleaned := validate.Text(r, "question", tamilText, 3, 200)
	if !r.OK() {
		t.Fatalf("expected valid text, got %v", r.Errors)
	}
	if cleaned != tamilText {
		t.Errorf("expected %s, got %s", tamilText, cleaned)
	}

	// Collapsing multiple whitespace
	dirty := "  Hello    World \t\n  "
	cleaned = validate.CleanText(dirty)
	if cleaned != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", cleaned)
	}

	// Too short check
	r = validate.New()
	validate.Text(r, "title", "Hi", 3, 10)
	if r.OK() {
		t.Fatal("expected error for text below min length")
	}

	// Too long check
	r = validate.New()
	validate.Text(r, "title", "A very long title here", 1, 10)
	if r.OK() {
		t.Fatal("expected error for text exceeding max length")
	}
}
