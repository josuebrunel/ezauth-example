package handler

import (
	"net/http"

	"github.com/josuebrunel/gopkg/etr"
	"github.com/labstack/echo/v5"

	"github.com/josuebrunel/ezauth-example/views"
)

// APIKeysPage lists the current user's API keys, fetched via
// auth.APIKeysList — EzAuth's own /auth/api-keys route responds with JSON
// only, so this is a direct Go call rather than a proxy to that route.
func (h *Handler) APIKeysPage(c *echo.Context) error {
	ctx := c.Request().Context()
	user, err := h.Auth.GetSessionUser(ctx)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}
	keys, err := h.Auth.APIKeysList(ctx, user.ID)
	if err != nil {
		return err
	}
	return etr.Render(c, http.StatusOK, views.ApiKeys(h.Auth, keys), nil)
}

// APIKeysFragment re-renders just the API key rows, used by the page to
// refresh its list after a create/revoke over HTMX.
func (h *Handler) APIKeysFragment(c *echo.Context) error {
	ctx := c.Request().Context()
	user, err := h.Auth.GetSessionUser(ctx)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}
	keys, err := h.Auth.APIKeysList(ctx, user.ID)
	if err != nil {
		return err
	}
	return etr.Render(c, http.StatusOK, views.ApiKeyRows(keys), nil)
}
