// Package config loads and validates runtime configuration.
//
// Everything the server needs comes from the environment. Values are read and
// checked exactly once, at startup, and the process refuses to boot if anything
// required is missing or obviously unsafe. A deployment that is misconfigured
// should fail loudly and immediately rather than start up and serve a broken or
// insecure application.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

const devVoterSalt = "livepoll-dev-salt-change-me"

// Config holds every tunable the server uses.
type Config struct {
	Env       string        // "development" or "production"
	Port      string        // TCP port to listen on
	MongoURI  string        // full MongoDB connection string
	MongoDB   string        // database name within the cluster
	RedisURL  string        // redis:// or rediss:// connection string
	JWTSecret []byte        // HMAC signing key for session tokens
	JWTTTL    time.Duration // how long a session lasts
	VoterSalt string        // salt for hashing anonymous voter fingerprints
	StaticDir string        // directory holding the built React app
}

// IsProduction reports whether the server is running in production mode. It
// gates behaviour that must differ between environments, such as whether
// session cookies are marked Secure.
func (c *Config) IsProduction() bool { return c.Env == "production" }

// Load reads configuration from a .env file if one exists, then from the real
// process environment, which always wins. That ordering matters: a deploy
// platform injects real environment variables, and those must never be
// silently overridden by a stray .env file baked into an image.
func Load() (*Config, error) {
	// Try both, so the server runs whether it was started from the repo root
	// or from inside backend/.
	loadDotEnv(".env")
	loadDotEnv("backend/.env")

	c := &Config{
		Env:       getenv("APP_ENV", "development"),
		Port:      getenv("PORT", "8080"),
		MongoURI:  os.Getenv("MONGO_URI"),
		MongoDB:   getenv("MONGO_DB", "livepoll"),
		RedisURL:  os.Getenv("REDIS_URL"),
		VoterSalt: getenv("VOTER_SALT", devVoterSalt),
		StaticDir: getenv("STATIC_DIR", "./web"),
		JWTTTL:    24 * time.Hour,
	}

	secret := os.Getenv("JWT_SECRET")

	var missing []string
	if c.MongoURI == "" {
		missing = append(missing, "MONGO_URI")
	}
	if c.RedisURL == "" {
		missing = append(missing, "REDIS_URL")
	}
	if secret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	// A short signing secret makes the whole JWT scheme worthless, since it can
	// be brute forced offline. Refusing to start is the only safe response.
	if len(secret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters, got %d", len(secret))
	}
	c.JWTSecret = []byte(secret)

	// The voter salt is what stops anonymous voter fingerprints from being
	// reversible by anyone who knows the hashing scheme. Shipping the public
	// default to production would defeat it.
	if c.IsProduction() && c.VoterSalt == devVoterSalt {
		return nil, fmt.Errorf("VOTER_SALT must be set to a unique random value when APP_ENV=production")
	}

	if c.Env != "development" && c.Env != "production" {
		return nil, fmt.Errorf("APP_ENV must be \"development\" or \"production\", got %q", c.Env)
	}

	return c, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadDotEnv is a deliberately minimal .env reader.
//
// It exists so local development needs no extra dependency for something this
// small. It handles comments, blank lines, optional "export " prefixes and
// quoted values, which covers every .env file this project uses. Anything more
// exotic is out of scope on purpose.
//
// Existing environment variables are never overwritten.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // absent .env is the normal case in production
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")

		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key == "" {
			continue
		}

		if len(val) >= 2 {
			first, last := val[0], val[len(val)-1]
			if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
				val = val[1 : len(val)-1]
			}
		}

		if _, exists := os.LookupEnv(key); exists {
			continue // the real environment wins
		}
		_ = os.Setenv(key, val)
	}
}
