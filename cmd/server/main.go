// Command server runs the ezauth-example app: an Echo v5 + Templ + HTMX
// showcase of EzAuth's authentication, RBAC, API key, session and
// impersonation features, backed by SQLite.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	csrf "filippo.io/csrf/gorilla"
	"github.com/josuebrunel/ezauth"
	ezconfig "github.com/josuebrunel/ezauth/pkg/config"
	"github.com/josuebrunel/gopkg/xlog"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"github.com/josuebrunel/ezauth-example/internal/bootstrap"
	appconfig "github.com/josuebrunel/ezauth-example/internal/config"
	"github.com/josuebrunel/ezauth-example/internal/handler"
)

func main() {
	migrateAction := flag.String("migrate", "", "up|down|revert — run EzAuth's own migrations and exit, instead of starting the server")
	flag.Parse()

	appCfg, err := appconfig.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "load app config:", err)
		os.Exit(1)
	}
	xlog.Setup(xlog.Config{Level: appCfg.LogLevel, Format: xlog.FormatText, Color: true})

	ezCfg, err := ezconfig.LoadConfig()
	if err != nil {
		xlog.Error("load ezauth config", "err", err)
		os.Exit(1)
	}

	auth, err := ezauth.New(&ezCfg, "auth")
	if err != nil {
		xlog.Error("initialize ezauth", "err", err)
		os.Exit(1)
	}

	switch *migrateAction {
	case "up":
		exitOnErr(auth.Migrate())
	case "down":
		exitOnErr(auth.MigrateDown())
	case "revert":
		exitOnErr(auth.MigrateRevert())
	case "":
		runServer(auth, appCfg)
	default:
		xlog.Error("unknown -migrate value, want up|down|revert", "value", *migrateAction)
		os.Exit(1)
	}
}

func exitOnErr(err error) {
	if err != nil {
		xlog.Error("migration failed", "err", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func runServer(auth *ezauth.EzAuth, appCfg appconfig.Config) {
	if err := auth.Migrate(); err != nil {
		xlog.Error("migrate ezauth schema", "err", err)
		os.Exit(1)
	}
	if err := bootstrap.SeedAdmin(context.Background(), auth, appCfg.AdminEmail); err != nil {
		xlog.Error("seed admin", "err", err)
		os.Exit(1)
	}
	if err := bootstrap.SeedDemoUsers(context.Background(), auth); err != nil {
		xlog.Error("seed demo users", "err", err)
		os.Exit(1)
	}

	h := handler.New(auth)
	e := echo.New()
	e.HTTPErrorHandler = errorHandler
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(echo.WrapMiddleware(auth.SessionMiddleware))
	e.Static("/static", "static")

	// csrfMW protects the app's own state-changing routes using the same
	// same-origin/Fetch-Metadata check EzAuth applies to /auth/* — see
	// filippo.io/csrf's docs. Safe methods (GET/HEAD/OPTIONS) are always
	// allowed, so it's only attached to POST routes below.
	csrfMW := echo.WrapMiddleware(csrf.Protect(nil))

	// Public pages — deliberately NOT behind LoginRequiredMiddleware (would
	// cause a redirect loop with EzAuth's own login/register redirects).
	e.GET("/", h.Home)
	e.GET("/login", h.LoginPage)
	e.GET("/register", h.RegisterPage)
	e.GET("/forgot-password", h.ForgotPasswordPage)
	e.GET("/reset-password", h.ResetPasswordPage)

	// Intercepts EzAuth's password-reset email link (GET, but that route is
	// POST-only on EzAuth's side) — must be registered before the "/auth/*"
	// wildcard mount matches it. Echo's router prefers an exact path match
	// over a wildcard regardless of registration order, but keeping the
	// specific route first here documents that intent.
	e.GET("/auth/password-reset/confirm", h.PasswordResetConfirmRedirect)

	// Every other EzAuth route: register/login/logout, password-reset,
	// impersonate, api-keys, sessions, admin/* — all respond with JSON or a
	// redirect, never HTML (see internal/handler package doc).
	e.Any("/auth/*", echo.WrapHandler(auth.Handler))

	authed := e.Group("", echo.WrapMiddleware(auth.LoginRequiredMiddleware))
	authed.GET("/dashboard", h.Dashboard)
	authed.GET("/profile", h.ProfilePage)
	authed.POST("/profile", h.ProfileUpdate, csrfMW)
	authed.POST("/profile/password", h.ProfilePasswordUpdate, csrfMW)
	authed.GET("/api-keys", h.APIKeysPage)
	authed.GET("/api-keys/fragment", h.APIKeysFragment)
	authed.GET("/sessions", h.SessionsPage)
	authed.GET("/sessions/fragment", h.SessionsFragment)

	admin := e.Group("/admin",
		echo.WrapMiddleware(auth.LoginRequiredMiddleware),
		echo.WrapMiddleware(auth.RequireRole(bootstrap.AdminRoleName)),
	)
	admin.GET("/users", h.AdminUsersList)
	admin.GET("/users/fragment", h.AdminUsersFragment)
	admin.GET("/users/:id", h.AdminUserDetail)
	admin.GET("/users/:id/roles-fragment", h.AdminUserRolesFragment)
	admin.POST("/users/:id/notes", h.AdminUserNotesUpdate, csrfMW)
	admin.GET("/roles", h.AdminRolesPage)
	admin.GET("/roles/fragment", h.AdminRolesFragment)
	admin.GET("/permissions", h.AdminPermissionsPage)
	admin.GET("/permissions/fragment", h.AdminPermissionsFragment)

	xlog.Info("starting ezauth-example", "addr", appCfg.Addr)
	if err := e.Start(appCfg.Addr); err != nil {
		xlog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func errorHandler(c *echo.Context, err error) {
	code := http.StatusInternalServerError
	message := "internal server error"
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		message = he.Message
	}
	xlog.Error("request failed", "path", c.Request().URL.Path, "status", code, "err", err)
	_ = c.HTML(code, fmt.Sprintf(`<!doctype html><html><head><link rel="stylesheet" href="/static/pico-daisy.css"></head>
<body><main class="container"><article><header>Error %d</header><p>%s</p><a href="/" role="button">Go home</a></article></main></body></html>`, code, message))
}
