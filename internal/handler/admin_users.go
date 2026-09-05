package handler

import (
	"net/http"

	"github.com/josuebrunel/ezauth/pkg/db/models"
	"github.com/josuebrunel/ezauth/pkg/service"
	"github.com/josuebrunel/gopkg/etr"
	"github.com/josuebrunel/gopkg/xlog"
	"github.com/labstack/echo/v5"

	"github.com/josuebrunel/ezauth-example/views"
)

// AdminUsersList lists/searches users, via auth.Service.UsersList — EzAuth's
// own /auth/admin/users route responds with JSON only, so this calls the
// underlying Go method directly rather than proxying that route.
func (h *Handler) AdminUsersList(c *echo.Context) error {
	search := c.QueryParam("search")
	result, err := h.Auth.Service.UsersList(c.Request().Context(), service.ListUsersOptions{Search: search})
	if err != nil {
		return err
	}
	return etr.Render(c, http.StatusOK, views.AdminUsersList(h.Auth, result.Users, search), nil)
}

// AdminUsersFragment re-renders just the user rows, used by the search box
// to refresh the list over HTMX.
func (h *Handler) AdminUsersFragment(c *echo.Context) error {
	search := c.QueryParam("search")
	result, err := h.Auth.Service.UsersList(c.Request().Context(), service.ListUsersOptions{Search: search})
	if err != nil {
		return err
	}
	return etr.Render(c, http.StatusOK, views.AdminUserRows(result.Users), nil)
}

// AdminUserDetail shows one user's roles, auth history and audit log, and
// exposes suspend/reactivate/impersonate/grant-role actions that POST/DELETE
// straight to EzAuth's existing /auth/admin/* routes.
func (h *Handler) AdminUserDetail(c *echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	user, err := h.Auth.Repo.UserGetByID(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	userRoles, err := h.Auth.UserRolesList(ctx, id)
	if err != nil {
		xlog.Error("failed to list user roles", "user_id", id, "err", err)
	}
	allRoles, err := h.Auth.Service.RolesList(ctx)
	if err != nil {
		xlog.Error("failed to list roles", "err", err)
	}
	history, err := h.Auth.Service.UserAuthHistory(ctx, id, 0)
	if err != nil {
		xlog.Error("failed to load auth history", "user_id", id, "err", err)
	}
	auditResult, err := h.Auth.Service.AuditLogs(ctx, id, service.ListAuditLogsOptions{})
	if err != nil {
		xlog.Error("failed to load audit log", "user_id", id, "err", err)
	}
	var auditLogs []*models.AuditLog
	if auditResult != nil {
		auditLogs = auditResult.Events
	}
	permissions, err := h.Auth.Repo.PermissionsByUserID(ctx, id)
	if err != nil {
		xlog.Error("failed to list effective permissions", "user_id", id, "err", err)
	}

	return etr.Render(c, http.StatusOK, views.AdminUserDetail(h.Auth, user, userRoles, allRoles, permissions, history, auditLogs), nil)
}

// AdminUserRolesFragment re-renders just a user's role badges, used to
// refresh the roles panel after a grant/revoke over HTMX.
func (h *Handler) AdminUserRolesFragment(c *echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	user, err := h.Auth.Repo.UserGetByID(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	roles, err := h.Auth.UserRolesList(ctx, id)
	if err != nil {
		return err
	}
	return etr.Render(c, http.StatusOK, views.AdminUserRoleRows(user, roles), nil)
}

// AdminUserNotesUpdate saves an admin-only note on a user's AppMetadata —
// deliberately not exposed on the user's own /profile page, to illustrate
// the UserMetadata/AppMetadata split EzAuth documents.
func (h *Handler) AdminUserNotesUpdate(c *echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	user, err := h.Auth.Repo.UserGetByID(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	user.SetAppMeta("notes", c.FormValue("notes"))
	if _, err := h.Auth.Service.UserUpdate(ctx, user); err != nil {
		xlog.Error("failed to save admin notes", "user_id", id, "err", err)
		return err
	}
	return c.Redirect(http.StatusFound, "/admin/users/"+id)
}
