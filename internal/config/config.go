// Package config holds the app-level configuration for ezauth-example.
//
// This is deliberately minimal: EzAuth reads its own EZAUTH_* environment
// variables directly via its own config.LoadConfig(), so this struct only
// covers what the app itself needs (listen address, admin bootstrap email,
// log level) — see internal/bootstrap for how AdminEmail is used.
package config

import "github.com/josuebrunel/gopkg/xenv"

// Config is the app's own configuration, loaded from APP_* env vars.
type Config struct {
	// Addr is the address the Echo server listens on.
	Addr string `env:"ADDR" default:":8080"`

	// AdminEmail, if set, is created as an admin user (with a random
	// password logged once, on first run) if it doesn't exist yet, and
	// granted the "admin" RBAC role on every server startup. Safe to leave
	// set permanently (idempotent) — see internal/bootstrap.SeedAdmin.
	AdminEmail string `env:"ADMIN_EMAIL"`

	// LogLevel controls verbosity for both the app's own logs and EzAuth's
	// internal logs (they share the same xlog package-level logger).
	// One of DEBUG, INFO, WARN, ERROR. Defaults to DEBUG so this example is
	// verbose out of the box — that's the point of an example project.
	LogLevel string `env:"LOG_LEVEL" default:"DEBUG"`
}

// Load reads Config from APP_* environment variables.
func Load() (Config, error) {
	var cfg Config
	if err := xenv.LoadWithOptions(&cfg, xenv.Options{Prefix: "APP_"}); err != nil {
		return cfg, err
	}
	return cfg, nil
}
