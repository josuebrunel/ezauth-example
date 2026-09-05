package handler

import (
	"net/http"

	"github.com/josuebrunel/gopkg/etr"
	"github.com/labstack/echo/v5"

	"github.com/josuebrunel/ezauth-example/views"
)

// AdminRolesPage lists roles and permissions, via auth.Service.RolesList /
// PermissionsList — EzAuth's own /auth/admin/roles route responds with JSON
// only.
func (h *Handler) AdminRolesPage(c *echo.Context) error {
	ctx := c.Request().Context()
	roles, err := h.Auth.Service.RolesList(ctx)
	if err != nil {
		return err
	}
	permissions, err := h.Auth.Service.PermissionsList(ctx)
	if err != nil {
		return err
	}
	return etr.Render(c, http.StatusOK, views.AdminRoles(h.Auth, roles, permissions), nil)
}

// AdminRolesFragment re-renders just the role list, used to refresh after a
// create/delete/grant over HTMX.
func (h *Handler) AdminRolesFragment(c *echo.Context) error {
	ctx := c.Request().Context()
	roles, err := h.Auth.Service.RolesList(ctx)
	if err != nil {
		return err
	}
	permissions, err := h.Auth.Service.PermissionsList(ctx)
	if err != nil {
		return err
	}
	return etr.Render(c, http.StatusOK, views.AdminRoleRows(h.Auth, roles, permissions), nil)
}

// AdminPermissionsPage lists permissions, via auth.Service.PermissionsList.
func (h *Handler) AdminPermissionsPage(c *echo.Context) error {
	permissions, err := h.Auth.Service.PermissionsList(c.Request().Context())
	if err != nil {
		return err
	}
	return etr.Render(c, http.StatusOK, views.AdminPermissions(h.Auth, permissions), nil)
}

// AdminPermissionsFragment re-renders just the permission rows, used to
// refresh after a create/delete over HTMX.
func (h *Handler) AdminPermissionsFragment(c *echo.Context) error {
	permissions, err := h.Auth.Service.PermissionsList(c.Request().Context())
	if err != nil {
		return err
	}
	return etr.Render(c, http.StatusOK, views.AdminPermissionRows(permissions), nil)
}
