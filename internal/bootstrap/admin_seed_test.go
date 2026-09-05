package bootstrap

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/josuebrunel/ezauth"
	ezconfig "github.com/josuebrunel/ezauth/pkg/config"
	"github.com/josuebrunel/ezauth/pkg/service"

	_ "modernc.org/sqlite"
)

func newTestAuth(t *testing.T) *ezauth.EzAuth {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "ezauth_test.db")
	cfg := &ezconfig.Config{
		DB: ezconfig.Database{
			Dialect: "sqlite",
			DSN:     dsn,
		},
		JWTSecret: "test-secret",
	}
	auth, err := ezauth.New(cfg, "auth")
	if err != nil {
		t.Fatalf("failed to create ezauth instance: %v", err)
	}
	if err := auth.Migrate(); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return auth
}

func registerUser(t *testing.T, auth *ezauth.EzAuth, email, password string) string {
	t.Helper()
	user, err := auth.Service.UserCreate(context.Background(), &service.RequestBasicAuth{
		Email:    email,
		Password: password,
	})
	if err != nil {
		t.Fatalf("failed to register test user %s: %v", email, err)
	}
	return user.ID
}

func TestSeedAdmin_NoAdminEmail(t *testing.T) {
	auth := newTestAuth(t)
	if err := SeedAdmin(context.Background(), auth, ""); err != nil {
		t.Fatalf("expected no-op with empty admin email, got error: %v", err)
	}
	if _, err := auth.Repo.RoleGetByName(context.Background(), AdminRoleName); err == nil {
		t.Error("expected no admin role to be created when APP_ADMIN_EMAIL is unset")
	}
}

func TestSeedAdmin_CreatesAdminIfMissing(t *testing.T) {
	auth := newTestAuth(t)
	ctx := context.Background()

	if err := SeedAdmin(ctx, auth, "admin@example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := auth.Repo.RoleGetByName(ctx, AdminRoleName); err != nil {
		t.Error("expected the admin role to be created")
	}
	user, err := auth.Repo.UserGetByEmail(ctx, "admin@example.com")
	if err != nil {
		t.Fatalf("expected the admin user to be created, got: %v", err)
	}
	has, err := auth.UserHasRole(ctx, user.ID, AdminRoleName)
	if err != nil {
		t.Fatalf("UserHasRole: %v", err)
	}
	if !has {
		t.Fatal("expected the newly created admin user to hold the admin role")
	}
}

func TestSeedAdmin_GrantsRoleToExistingUser(t *testing.T) {
	auth := newTestAuth(t)
	ctx := context.Background()
	userID := registerUser(t, auth, "admin@example.com", "correct-horse-battery-staple")

	if err := SeedAdmin(ctx, auth, "admin@example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	has, err := auth.UserHasRole(ctx, userID, AdminRoleName)
	if err != nil {
		t.Fatalf("UserHasRole: %v", err)
	}
	if !has {
		t.Fatal("expected admin role to be granted")
	}
}

// TestSeedAdmin_PreservesExistingPassword guards against SeedAdmin ever
// resetting an already-registered admin's password when it only needs to
// grant a role — the random-password path is for first-run creation only.
func TestSeedAdmin_PreservesExistingPassword(t *testing.T) {
	auth := newTestAuth(t)
	ctx := context.Background()
	const password = "correct-horse-battery-staple"
	registerUser(t, auth, "admin@example.com", password)

	if err := SeedAdmin(ctx, auth, "admin@example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := auth.Service.UserAuthenticate(ctx, service.RequestBasicAuth{
		Email:    "admin@example.com",
		Password: password,
	}); err != nil {
		t.Fatalf("expected original password to still authenticate, got: %v", err)
	}
}

func TestSeedAdmin_IdempotentAcrossRestarts(t *testing.T) {
	auth := newTestAuth(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if err := SeedAdmin(ctx, auth, "admin@example.com"); err != nil {
			t.Fatalf("run %d: unexpected error: %v", i, err)
		}
	}

	user, err := auth.Repo.UserGetByEmail(ctx, "admin@example.com")
	if err != nil {
		t.Fatalf("UserGetByEmail: %v", err)
	}
	result, err := auth.Service.UsersList(ctx, service.ListUsersOptions{Search: "admin@example.com"})
	if err != nil {
		t.Fatalf("UsersList: %v", err)
	}
	if len(result.Users) != 1 {
		t.Fatalf("expected exactly one user for admin@example.com after repeated seeding, got %d", len(result.Users))
	}

	roles, err := auth.UserRolesList(ctx, user.ID)
	if err != nil {
		t.Fatalf("UserRolesList: %v", err)
	}
	count := 0
	for _, r := range roles {
		if r.Name == AdminRoleName {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one admin role grant after repeated seeding, got %d", count)
	}
}

func TestGeneratePassword(t *testing.T) {
	password, err := generatePassword()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(password) < 8 {
		t.Fatalf("generated password %q is shorter than EzAuth's 8-character minimum", password)
	}
	other, err := generatePassword()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if password == other {
		t.Fatal("expected two successive calls to generate different passwords")
	}
}
