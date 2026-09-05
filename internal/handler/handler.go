// Package handler holds the app-owned Echo v5 handlers for ezauth-example.
//
// Most pages here just call an EzAuth Go method directly and render a Templ
// view — EzAuth's own /auth/* HTTP routes only ever respond with JSON, never
// HTML (see the handful of "Form*" handlers in EzAuth's pkg/handler), so
// list/detail pages can't be built by fetching those routes. Only two
// handlers in this package own real business logic (ProfileUpdate and
// ProfilePasswordUpdate) — everywhere else, forms and HTMX actions POST or
// DELETE straight to EzAuth's existing /auth/* routes.
package handler

import "github.com/josuebrunel/ezauth"

// Handler holds the shared EzAuth instance used by every app-owned route.
type Handler struct {
	Auth *ezauth.EzAuth
}

// New returns a Handler wired to the given EzAuth instance.
func New(auth *ezauth.EzAuth) *Handler {
	return &Handler{Auth: auth}
}
