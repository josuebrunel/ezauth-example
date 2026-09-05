package handler

import (
	"net/http"

	"github.com/josuebrunel/ezauth"
	"github.com/josuebrunel/ezauth/pkg/service"
	"github.com/josuebrunel/gopkg/etr"
	"github.com/josuebrunel/gopkg/xlog"
	"github.com/labstack/echo/v5"

	"github.com/josuebrunel/ezauth-example/internal/bootstrap"
	"github.com/josuebrunel/ezauth-example/views"
)

// Home redirects to the dashboard if logged in, otherwise to the login page.
func (h *Handler) Home(c *echo.Context) error {
	if ezauth.IsAuthenticated(c.Request().Context()) {
		return c.Redirect(http.StatusFound, "/dashboard")
	}
	return c.Redirect(http.StatusFound, "/login")
}

// LoginPage renders the app-owned login form, which POSTs to /auth/login.
func (h *Handler) LoginPage(c *echo.Context) error {
	if ezauth.IsAuthenticated(c.Request().Context()) {
		return c.Redirect(http.StatusFound, "/dashboard")
	}
	return etr.Render(c, http.StatusOK, views.Login(h.Auth), nil)
}

// RegisterPage renders the app-owned registration form, which POSTs to
// /auth/register.
func (h *Handler) RegisterPage(c *echo.Context) error {
	if ezauth.IsAuthenticated(c.Request().Context()) {
		return c.Redirect(http.StatusFound, "/dashboard")
	}
	return etr.Render(c, http.StatusOK, views.Register(h.Auth), nil)
}

// ForgotPasswordPage renders the request-a-reset-link form, which POSTs to
// /auth/password-reset/request.
func (h *Handler) ForgotPasswordPage(c *echo.Context) error {
	return etr.Render(c, http.StatusOK, views.ForgotPassword(h.Auth), nil)
}

// ResetPasswordPage renders the choose-a-new-password form. The token comes
// from the query string — see PasswordResetConfirmRedirect for how a user
// following the emailed link actually lands here.
func (h *Handler) ResetPasswordPage(c *echo.Context) error {
	token := c.QueryParam("token")
	return etr.Render(c, http.StatusOK, views.ResetPassword(h.Auth, token), nil)
}

// PasswordResetConfirmRedirect intercepts GET /auth/password-reset/confirm.
//
// EzAuth's password-reset email always links to
// "{BaseURL}/auth/password-reset/confirm?token=...", but that route is
// POST-only (see ezauth pkg/handler/handler.go) — EzAuth expects the app to
// present its own confirmation page. Registering this exact path on the app
// router, ahead of the "/auth/*" wildcard mount, lets Echo's router match it
// first (an exact path always wins over a "*" wildcard, regardless of
// registration order) and redirect to our own /reset-password page instead
// of falling through to EzAuth's handler and getting a 405.
func (h *Handler) PasswordResetConfirmRedirect(c *echo.Context) error {
	token := c.QueryParam("token")
	xlog.Debug("intercepted GET /auth/password-reset/confirm, redirecting to app page", "has_token", token != "")
	return c.Redirect(http.StatusFound, "/reset-password?token="+token)
}

// Dashboard shows the current user's roles and quick links into the rest of
// the example, plus platform-wide stats for admins.
func (h *Handler) Dashboard(c *echo.Context) error {
	ctx := c.Request().Context()
	user, err := h.Auth.GetSessionUser(ctx)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}
	roles, err := h.Auth.UserRolesList(ctx, user.ID)
	if err != nil {
		xlog.Error("failed to list user roles", "user_id", user.ID, "err", err)
	}
	sessions, err := h.Auth.Sessions(ctx, user.ID)
	if err != nil {
		xlog.Error("failed to list sessions", "user_id", user.ID, "err", err)
	}

	stats := views.DashboardStats{SessionCount: len(sessions)}
	if isAdmin, _ := h.Auth.UserHasRole(ctx, user.ID, bootstrap.AdminRoleName); isAdmin {
		stats.IsAdmin = true
		if result, err := h.Auth.Service.UsersList(ctx, service.ListUsersOptions{Limit: 100}); err == nil {
			stats.UserCount, stats.UserCountMore = len(result.Users), result.HasMore
		} else {
			xlog.Error("failed to count users", "err", err)
		}
		if allRoles, err := h.Auth.Service.RolesList(ctx); err == nil {
			stats.RoleCount = len(allRoles)
		} else {
			xlog.Error("failed to count roles", "err", err)
		}
		if perms, err := h.Auth.Service.PermissionsList(ctx); err == nil {
			stats.PermissionCount = len(perms)
		} else {
			xlog.Error("failed to count permissions", "err", err)
		}
	}

	return etr.Render(c, http.StatusOK, views.Dashboard(h.Auth, user, roles, stats), nil)
}
