package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SessionCookie is the name of the cookie carrying the session token.
//
// The __Host- prefix is not used here because it forbids the Domain attribute
// and requires Secure, which would break plain-HTTP local development. Secure
// is set explicitly in production instead.
const SessionCookie = "livepoll_session"

const tokenIssuer = "livepoll"

// ErrInvalidToken covers every reason a token was not accepted.
//
// The reasons are collapsed into one error on purpose. Telling a caller whether
// a token was expired, forged or malformed hands useful information to someone
// probing the endpoint, and no legitimate client needs to know the difference.
var ErrInvalidToken = errors.New("invalid or expired session")

// Claims is the payload carried inside a session token.
//
// Name and email are included so the frontend can render a header without an
// extra round trip. Nothing sensitive goes in here: a JWT is signed, which
// makes it tamper-evident, but it is *not* encrypted, and anyone holding one
// can read its contents.
type Claims struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	jwt.RegisteredClaims
}

// TokenManager issues and verifies session tokens.
type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

// NewTokenManager returns a manager signing with the given secret.
func NewTokenManager(secret []byte, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: secret, ttl: ttl}
}

// TTL reports how long issued tokens remain valid. The HTTP layer uses it to
// set a matching cookie lifetime.
func (m *TokenManager) TTL() time.Duration { return m.ttl }

// Issue signs a new session token for a user and returns it with its expiry.
func (m *TokenManager) Issue(userID, email, name string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.ttl)

	claims := Claims{
		Name:  name,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    tokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign session token: %w", err)
	}
	return signed, expiresAt, nil
}

// Parse verifies a token and returns its claims.
//
// Two separate guards pin the algorithm to HMAC-SHA256. This is not belt and
// braces for its own sake: the classic JWT attack is to hand the server a token
// whose header says "alg":"none", or one signed with HMAC using the server's
// *public* key as the secret when the server expects RSA. A verifier that
// trusts the algorithm named in the token it is trying to verify can be talked
// into accepting anything. The algorithm must come from the server's own
// configuration, never from the token.
func (m *TokenManager) Parse(raw string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(raw, claims,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
			}
			return m.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(tokenIssuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.Subject == "" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
