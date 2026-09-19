// Package validate provides the input checking used by every request handler.
//
// The rule this package exists to enforce is simple: nothing arriving from a
// client is trusted. Values are checked for shape and size, normalised, and
// only then allowed near the database. The frontend does its own checking too,
// but purely so the forms feel responsive — it is never relied upon, because
// anyone can bypass it with a single curl command.
//
// Errors accumulate rather than short-circuiting, so a client gets every
// problem with a submission at once instead of discovering them one reload at
// a time.
package validate

import (
	"errors"
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"
)

// bcrypt refuses passwords longer than 72 bytes. Rejecting them up front with a
// clear message is far better than surfacing a confusing hashing error later.
const MaxPasswordBytes = 72

// FieldError names one problem with one field, in language safe to show a user.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Result collects field errors across a whole request payload.
type Result struct {
	Errors []FieldError `json:"errors"`
}

// New returns an empty Result ready to collect errors.
func New() *Result { return &Result{} }

// Add records a problem with a field.
func (r *Result) Add(field, message string) {
	r.Errors = append(r.Errors, FieldError{Field: field, Message: message})
}

// Check records a problem when cond is false. Reads as an assertion at the call
// site: Check(len(x) > 0, "name", "is required").
func (r *Result) Check(cond bool, field, message string) {
	if !cond {
		r.Add(field, message)
	}
}

// OK reports whether the payload passed every check.
func (r *Result) OK() bool { return len(r.Errors) == 0 }

// ErrInvalid is returned by Err when validation failed. Handlers translate it
// into a 422 response carrying the field list.
var ErrInvalid = errors.New("validation failed")

// Err returns ErrInvalid if anything failed, nil otherwise.
func (r *Result) Err() error {
	if r.OK() {
		return nil
	}
	return ErrInvalid
}

// Text normalises and length-checks a single-line text field, returning the
// cleaned value.
//
// Cleaning happens before measuring, so a string padded with spaces or
// zero-width characters cannot sneak past a minimum length check.
//
// Lengths are counted in runes, not bytes, so a question written in Tamil or
// Hindi gets the same allowance as one written in English. Counting bytes would
// quietly give non-Latin scripts a third of the space.
func Text(r *Result, field, value string, min, max int) string {
	cleaned := CleanText(value)
	n := utf8.RuneCountInString(cleaned)

	switch {
	case n == 0 && min > 0:
		r.Add(field, "is required")
	case n < min:
		r.Add(field, "is too short")
	case n > max:
		r.Add(field, "is too long")
	}
	return cleaned
}

// CleanText normalises a single-line string: every kind of whitespace collapses
// to a single plain space, invisible and control characters are dropped, and
// the ends are trimmed.
//
// The two cases are ordered deliberately. Whitespace is handled first so that a
// non-breaking space becomes a normal space rather than vanishing. Everything
// else that is not printable — C0 and C1 control codes, and the Unicode format
// characters such as the byte order mark and the zero-width spaces and joiners —
// is removed outright.
//
// Dropping zero-width characters matters for more than tidiness: they are the
// standard trick for making two different strings look identical, which is how
// you would sneak a poll option past a human reviewer while it reads as
// something else to a program.
func CleanText(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	prevSpace := false
	for _, ru := range s {
		switch {
		case unicode.IsSpace(ru):
			if !prevSpace {
				b.WriteRune(' ')
				prevSpace = true
			}
		case !unicode.IsPrint(ru):
			continue
		default:
			b.WriteRune(ru)
			prevSpace = false
		}
	}
	return strings.TrimSpace(b.String())
}

// Email validates and normalises an address, returning it lowercased.
//
// Normalising on the way in is what makes the unique index on email meaningful:
// without it "Vignesh@Example.com" and "vignesh@example.com" would be two
// separate accounts.
func Email(r *Result, field, value string) string {
	v := strings.ToLower(strings.TrimSpace(value))

	if v == "" {
		r.Add(field, "is required")
		return v
	}
	// The practical upper bound for an address, per RFC 5321.
	if len(v) > 254 {
		r.Add(field, "is too long")
		return v
	}
	addr, err := mail.ParseAddress(v)
	if err != nil || addr.Address != v {
		// The second condition rejects display-name forms such as
		// "Vignesh <v@example.com>", which ParseAddress happily accepts but
		// which a signup form has no business storing.
		r.Add(field, "is not a valid email address")
	}
	return v
}

// Password checks a plaintext password.
//
// It is deliberately never trimmed or cleaned. Leading and trailing spaces are
// legitimate characters that a password manager may well have generated, and
// silently stripping them at signup would lock the user out at login.
func Password(r *Result, field, value string) {
	switch {
	case value == "":
		r.Add(field, "is required")
	case utf8.RuneCountInString(value) < 8:
		r.Add(field, "must be at least 8 characters")
	case len(value) > MaxPasswordBytes:
		r.Add(field, "must be at most 72 bytes")
	}
}
