// Package idgen generates the short, random identifiers used for share links.
package idgen

import (
	"crypto/rand"
	"fmt"
)

// alphabet excludes characters that are easy to confuse when a code is read off
// a screen or typed from a phone: 0/O, 1/l/I. A share code will get transcribed
// by hand at some point, and unambiguous characters cost nothing.
const alphabet = "23456789abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"

// Code returns a cryptographically random identifier of n characters.
//
// crypto/rand is used rather than math/rand deliberately. math/rand is
// predictable from a handful of outputs, which would let someone enumerate
// other people's poll links; crypto/rand is not. With this 56-character
// alphabet, an 8-character code carries roughly 46 bits of entropy, which makes
// guessing a valid code impractical.
func Code(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("idgen: length must be positive, got %d", n)
	}

	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("idgen: read random bytes: %w", err)
	}

	// 256 is not a multiple of 56, so plain modulo would make the first few
	// characters of the alphabet very slightly more likely. Any byte landing in
	// the short final block is redrawn so the distribution stays uniform.
	limit := byte(256 - (256 % len(alphabet)))
	out := make([]byte, 0, n)

	for len(out) < n {
		for _, b := range buf {
			if b >= limit {
				continue
			}
			out = append(out, alphabet[int(b)%len(alphabet)])
			if len(out) == n {
				break
			}
		}
		if len(out) < n {
			if _, err := rand.Read(buf); err != nil {
				return "", fmt.Errorf("idgen: read random bytes: %w", err)
			}
		}
	}
	return string(out), nil
}
