// Package auth handles password hashing and session tokens.
//
// It deliberately knows nothing about HTTP or about Gin. Keeping it free of
// framework types means the security-critical logic can be read, reasoned about
// and unit tested on its own, and the HTTP layer is left with nothing to do but
// move values in and out of requests.
package auth

import "golang.org/x/crypto/bcrypt"

// BcryptCost is the work factor for password hashing.
//
// 12 is a deliberate choice, not a default. bcrypt's cost is a power of two, so
// each step doubles the time taken: the goal is to be slow enough that offline
// brute forcing a stolen database is impractical, while staying fast enough
// (tens of milliseconds) that logging in feels instant. This is also why
// password hashing must never use SHA-256 or similar — those are built to be
// fast, which is exactly the wrong property here.
const BcryptCost = 12

// HashPassword returns the bcrypt hash of a plaintext password. The salt is
// generated per call and stored inside the returned string, so two users with
// the same password still get different hashes and a precomputed rainbow table
// is useless.
func HashPassword(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

// CheckPassword reports whether plain matches the stored hash.
//
// bcrypt's own comparison is used rather than hashing and comparing strings,
// because it reads the cost and salt back out of the stored hash and compares
// in constant time.
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// BurnTiming spends roughly the same amount of CPU as a real password check.
//
// It is called on the login path when no account exists for the submitted
// email. Without it, a missing account would return noticeably faster than a
// wrong password, and that difference alone is enough to let someone discover
// which email addresses are registered. Returning an identical error message is
// not sufficient on its own — the timing has to match too.
func BurnTiming() {
	_, _ = bcrypt.GenerateFromPassword([]byte("timing-equalisation-only"), BcryptCost)
}
