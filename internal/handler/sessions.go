package handler

import (
	"net/http"

	"github.com/josuebrunel/gopkg/etr"
	"github.com/labstack/echo/v5"

	"github.com/josuebrunel/ezauth-example/views"
)

// SessionsPage lists the current user's active sessions, fetched via
// auth.Sessions — same JSON-only-route rationale as APIKeysPage.
func (h *Handler) SessionsPage(c *echo.Context) error {
	ctx := c.Request().Context()
	user, err := h.Auth.GetSessionUser(ctx)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}
	sessions, err := h.Auth.Sessions(ctx, user.ID)
	if err != nil {
		return err
	}
	return etr.Render(c, http.StatusOK, views.Sessions(h.Auth, sessions), nil)
}

// SessionsFragment re-renders just the session rows, used to refresh the
// list after a revoke over HTMX.
func (h *Handler) SessionsFragment(c *echo.Context) error {
	ctx := c.Request().Context()
	user, err := h.Auth.GetSessionUser(ctx)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}
	sessions, err := h.Auth.Sessions(ctx, user.ID)
	if err != nil {
		return err
	}
	return etr.Render(c, http.StatusOK, views.SessionRows(sessions), nil)
}
