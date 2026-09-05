// Package bootstrap holds one-time startup wiring for the example app.
package bootstrap

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/josuebrunel/ezauth"
	"github.com/josuebrunel/ezauth/pkg/service"
	"github.com/josuebrunel/gopkg/xlog"
)

// AdminRoleName is the RBAC role checked by RequireRole on every /admin
// route in this app (see cmd/server/main.go).
const AdminRoleName = "admin"

// SeedAdmin ensures adminEmail exists and holds AdminRoleName, creating the
// account with a random password on first run if it doesn't exist yet — this
// is a demo convenience so the app is usable immediately, not a pattern for
// a real deployment. It is meant to be called unconditionally on every
// server startup and is fully idempotent:
//   - The "admin" role is created once; if it already exists, RoleCreate's
//     own "role already exists" error is avoided by checking RoleGetByName
//     first, so it's never called a second time.
//   - The admin user is created once, with a random password logged at
//     INFO on that first run only — an existing account's password is never
//     touched on later runs.
//   - auth.UserRoleGrant is a verified no-op at the DB level when the user
//     already holds the role (see ezauth pkg/service/rbac.go) — no duplicate
//     audit event, no error.
func SeedAdmin(ctx context.Context, auth *ezauth.EzAuth, adminEmail string) error {
	if adminEmail == "" {
		xlog.Debug("admin bootstrap: APP_ADMIN_EMAIL not set, skipping")
		return nil
	}

	if _, err := auth.Repo.RoleGetByName(ctx, AdminRoleName); err != nil {
		xlog.Debug("admin bootstrap: admin role missing, creating", "role", AdminRoleName)
		if _, err := auth.RoleCreate(ctx, AdminRoleName, "Full administrative access"); err != nil {
			return fmt.Errorf("seed admin role: %w", err)
		}
	}

	user, err := auth.Repo.UserGetByEmail(ctx, adminEmail)
	if err != nil {
		password, genErr := generatePassword()
		if genErr != nil {
			return fmt.Errorf("generate admin password: %w", genErr)
		}
		user, err = auth.Service.UserCreate(ctx, &service.RequestBasicAuth{Email: adminEmail, Password: password})
		if err != nil {
			return fmt.Errorf("create admin user: %w", err)
		}
		xlog.Info("admin bootstrap: created admin user — save these credentials, they are shown only once",
			"email", adminEmail, "password", password)
	}

	if err := auth.UserRoleGrant(ctx, user.ID, AdminRoleName); err != nil {
		return fmt.Errorf("seed admin grant: %w", err)
	}
	xlog.Info("admin bootstrap: admin account ready", "email", adminEmail, "user_id", user.ID)
	return nil
}

// generatePassword returns a random hex string well within EzAuth's 8-128
// character password bounds.
func generatePassword() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
