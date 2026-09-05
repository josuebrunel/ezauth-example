package bootstrap

import (
	"context"
	"fmt"

	"github.com/josuebrunel/ezauth"
	"github.com/josuebrunel/ezauth/pkg/service"
	"github.com/josuebrunel/gopkg/xlog"
)

// DemoUser describes one of the sample accounts SeedDemoUsers creates.
type DemoUser struct {
	Email     string
	FirstName string
	LastName  string
	Username  string
}

// DemoUsers is the fixed set of sample accounts this example seeds so the
// admin pages (user list, impersonation, suspend/reactivate, role grants)
// have real targets to try out immediately, without registering accounts
// by hand first.
var DemoUsers = []DemoUser{
	{Email: "alice@example.com", FirstName: "Alice", LastName: "Johnson", Username: "alice"},
	{Email: "bob@example.com", FirstName: "Bob", LastName: "Smith", Username: "bob"},
	{Email: "carol@example.com", FirstName: "Carol", LastName: "Davis", Username: "carol"},
	{Email: "dave@example.com", FirstName: "Dave", LastName: "Wilson", Username: "dave"},
}

// SeedDemoUsers creates each of DemoUsers with a random password if it
// doesn't exist yet, logging the credentials once at creation time — same
// demo-convenience rationale and idempotency guarantee as SeedAdmin. An
// account that already exists (e.g. an admin edited it, or a previous run
// already created it) is left completely untouched, including its
// password.
func SeedDemoUsers(ctx context.Context, auth *ezauth.EzAuth) error {
	for _, du := range DemoUsers {
		if _, err := auth.Repo.UserGetByEmail(ctx, du.Email); err == nil {
			continue
		}

		password, err := generatePassword()
		if err != nil {
			return fmt.Errorf("generate demo password for %s: %w", du.Email, err)
		}
		if _, err := auth.Service.UserCreate(ctx, &service.RequestBasicAuth{
			Email:     du.Email,
			Password:  password,
			FirstName: du.FirstName,
			LastName:  du.LastName,
			Username:  du.Username,
		}); err != nil {
			return fmt.Errorf("create demo user %s: %w", du.Email, err)
		}
		xlog.Info("demo bootstrap: created sample user — save these credentials, they are shown only once",
			"email", du.Email, "password", password)
	}
	return nil
}
