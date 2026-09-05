package bootstrap

import (
	"context"
	"testing"

	"github.com/josuebrunel/ezauth/pkg/service"
)

func TestSeedDemoUsers_CreatesAll(t *testing.T) {
	auth := newTestAuth(t)
	ctx := context.Background()

	if err := SeedDemoUsers(ctx, auth); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, du := range DemoUsers {
		user, err := auth.Repo.UserGetByEmail(ctx, du.Email)
		if err != nil {
			t.Errorf("expected demo user %s to be created, got: %v", du.Email, err)
			continue
		}
		if user.FirstName != du.FirstName || user.LastName != du.LastName || user.Username != du.Username {
			t.Errorf("demo user %s has unexpected profile fields: %+v", du.Email, user)
		}
	}
}

func TestSeedDemoUsers_IdempotentAndPreservesEdits(t *testing.T) {
	auth := newTestAuth(t)
	ctx := context.Background()

	if err := SeedDemoUsers(ctx, auth); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Simulate an admin editing one of the demo users.
	alice, err := auth.Repo.UserGetByEmail(ctx, "alice@example.com")
	if err != nil {
		t.Fatalf("UserGetByEmail: %v", err)
	}
	alice.LastName = "Edited"
	if _, err := auth.Service.UserUpdate(ctx, alice); err != nil {
		t.Fatalf("UserUpdate: %v", err)
	}

	if err := SeedDemoUsers(ctx, auth); err != nil {
		t.Fatalf("unexpected error on second run: %v", err)
	}

	result, err := auth.Service.UsersList(ctx, service.ListUsersOptions{Limit: 100})
	if err != nil {
		t.Fatalf("UsersList: %v", err)
	}
	count := 0
	for _, u := range result.Users {
		for _, du := range DemoUsers {
			if u.Email == du.Email {
				count++
			}
		}
	}
	if count != len(DemoUsers) {
		t.Fatalf("expected exactly %d demo users after repeated seeding, got %d", len(DemoUsers), count)
	}

	alice, err = auth.Repo.UserGetByEmail(ctx, "alice@example.com")
	if err != nil {
		t.Fatalf("UserGetByEmail: %v", err)
	}
	if alice.LastName != "Edited" {
		t.Fatalf("expected re-seeding to leave an already-existing demo user's edits alone, got LastName=%q", alice.LastName)
	}
}
