package handler

import (
	"net/http"
	"net/url"

	"github.com/josuebrunel/gopkg/etr"
	"github.com/josuebrunel/gopkg/xlog"
	"github.com/labstack/echo/v5"

	"github.com/josuebrunel/ezauth-example/views"
)

// ProfilePage renders the profile-edit and change-password forms.
//
// EzAuth has no dedicated HTTP route for arbitrary profile-field updates —
// see ProfileUpdate below — so, unlike most pages in this app, this one
// isn't a thin wrapper around an EzAuth route; ProfileUpdate/
// ProfilePasswordUpdate own real business logic.
func (h *Handler) ProfilePage(c *echo.Context) error {
	ctx := c.Request().Context()
	user, err := h.Auth.GetSessionUser(ctx)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}
	return etr.Render(c, http.StatusOK, views.Profile(h.Auth, user, c.QueryParam("success"), c.QueryParam("error")), nil)
}

// ProfileUpdate saves the editable profile fields plus a user-editable
// "theme" UserMetadata preference, via auth.Service.UserUpdate — the only
// way to persist these fields, since EzAuth exposes no HTTP route for it.
func (h *Handler) ProfileUpdate(c *echo.Context) error {
	ctx := c.Request().Context()
	user, err := h.Auth.GetSessionUser(ctx)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}

	user.FirstName = c.FormValue("first_name")
	user.LastName = c.FormValue("last_name")
	user.Username = c.FormValue("username")
	user.AvatarURL = c.FormValue("avatar_url")
	user.Locale = c.FormValue("locale")
	user.Timezone = c.FormValue("timezone")
	user.SetMeta("theme", c.FormValue("theme"))

	if _, err := h.Auth.Service.UserUpdate(ctx, user); err != nil {
		xlog.Error("profile update failed", "user_id", user.ID, "err", err)
		return c.Redirect(http.StatusFound, "/profile?error="+url.QueryEscape(err.Error()))
	}
	xlog.Info("profile updated", "user_id", user.ID)
	return c.Redirect(http.StatusFound, "/profile?success="+url.QueryEscape("Profile updated"))
}

// ProfilePasswordUpdate changes the current user's password via
// auth.Service.UserUpdatePassword — EzAuth's password-reset routes are for
// the forgot-password flow (unauthenticated, token-based); changing your
// password while logged in has no dedicated HTTP route either.
func (h *Handler) ProfilePasswordUpdate(c *echo.Context) error {
	ctx := c.Request().Context()
	user, err := h.Auth.GetSessionUser(ctx)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}

	password := c.FormValue("password")
	if password != c.FormValue("password_confirm") {
		return c.Redirect(http.StatusFound, "/profile?error="+url.QueryEscape("passwords do not match"))
	}

	if _, err := h.Auth.Service.UserUpdatePassword(ctx, user, password); err != nil {
		xlog.Error("password update failed", "user_id", user.ID, "err", err)
		return c.Redirect(http.StatusFound, "/profile?error="+url.QueryEscape(err.Error()))
	}
	xlog.Info("password updated", "user_id", user.ID)
	return c.Redirect(http.StatusFound, "/profile?success="+url.QueryEscape("Password updated"))
}
