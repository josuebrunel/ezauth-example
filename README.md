# EzAuth Example

A runnable showcase of [EzAuth](https://github.com/josuebrunel/ezauth) — authentication, RBAC,
API keys, sessions, and impersonation — wired into **Echo v5**, **Templ**, **HTMX**, and
**SQLite**, styled with [picodaisy](https://github.com/josuebrunel/picodaisy).

Most of this app is a thin UI layer over EzAuth's *existing* HTTP routes. EzAuth already ships
working endpoints for registration, login, password reset, API keys, sessions, impersonation,
and admin/RBAC management — this project mostly renders views around them and lets forms/HTMX
POST straight to `/auth/*`. Only two things needed app-owned business logic: **profile updates**
and **changing your password while logged in**, since EzAuth exposes those only as Go method
calls (`auth.Service.UserUpdate` / `UserUpdatePassword`), not HTTP routes.

## Stack deviation

Bob, Goose, and RiverQueue — normally part of this stack — are **intentionally omitted**. EzAuth
owns 100% of the persisted schema and runs its own migrations (`auth.Migrate()`); this app has no
domain data of its own to query, migrate, or queue jobs against.

## Setup

### Docker Compose (recommended)

```sh
cp .env.example .env
# edit .env: set EZAUTH_API_KEY and EZAUTH_JWT_SECRET to real random values
make docker-up
```

Visit `http://localhost:8080`. Password-reset emails land in [Mailpit](http://localhost:8025) —
`docker-compose.yml` runs it alongside the app, no SMTP account needed.

### Local (no Docker)

```sh
export EZAUTH_API_KEY=dev-api-key
export EZAUTH_JWT_SECRET=dev-jwt-secret
export APP_ADMIN_EMAIL=admin@example.com
make build
make migrate-up
make run
```

### First login

On first boot, the app automatically:

- Creates an `admin` RBAC role.
- Creates the user named by `APP_ADMIN_EMAIL` (if it doesn't already exist) with a **random
  password logged once** — look for `admin bootstrap: created admin user` in the startup logs —
  and grants it the `admin` role.
- Creates four sample users (`alice@example.com`, `bob@example.com`, `carol@example.com`,
  `dave@example.com`) the same way, logged as `demo bootstrap: created sample user`, so the admin
  pages (impersonation, suspend/reactivate, role grants) have real targets immediately without
  registering accounts by hand.

All of this is idempotent — safe to leave `APP_ADMIN_EMAIL` set permanently and restart the app
as often as you like; existing accounts and their passwords are never touched again.

## Env vars

### App (`APP_*`)

| Var | Default | Purpose |
|---|---|---|
| `APP_ADDR` | `:8080` | Address the Echo server listens on |
| `APP_ADMIN_EMAIL` | — | Email to bootstrap as admin (see above) |
| `APP_LOG_LEVEL` | `DEBUG` | `DEBUG`/`INFO`/`WARN`/`ERROR` — shared by the app and EzAuth's own internal logs |

### EzAuth (`EZAUTH_*`, read directly by the library)

| Var | Required | Purpose |
|---|---|---|
| `EZAUTH_API_KEY` | yes | Gates `/auth/api/*` JSON routes (unused by this app's UI, but required at startup) |
| `EZAUTH_JWT_SECRET` | yes | Signs session/access tokens |
| `EZAUTH_DB_DIALECT` / `EZAUTH_DB_DSN` | | `sqlite` / path to the `.db` file |
| `EZAUTH_LOGIN_PAGE_URL` / `EZAUTH_REGISTER_PAGE_URL` | | Set to `/login` / `/register` — avoids a redirect loop, see `go-stack` skill notes |
| `EZAUTH_REDIRECT_AFTER_LOGIN` / `_REGISTER` | | Set to `/dashboard` / `/login` |
| `EZAUTH_SMTP_HOST` / `_PORT` / `_FROM` | | Point at Mailpit for local password-reset emails |
| `EZAUTH_BASE_URL` | | Used to build the password-reset email link |

## Feature walkthrough

1. **Auth** — `/register`, `/login`, logout button in the nav. Forms POST to `/auth/register` /
   `/auth/login`; no hidden CSRF token needed (EzAuth checks `Sec-Fetch-Site`/`Origin` instead of
   tokens — see `filippo.io/csrf`).
2. **Password reset** — `/forgot-password` → `/auth/password-reset/request` → email (view it in
   Mailpit) → `/reset-password?token=...` → `/auth/password-reset/confirm`. Note:
   `cmd/server/main.go` registers a specific `GET /auth/password-reset/confirm` handler that
   redirects to the app's own `/reset-password` page, ahead of the `/auth/*` wildcard mount —
   EzAuth's mailed link points straight at that (POST-only) EzAuth route, which has no GET
   handler of its own.
3. **Profile** (`/profile`) — edit name/username/avatar/locale/timezone and a `UserMetadata`
   "theme" preference, via `auth.Service.UserUpdate` (app-owned handler — no EzAuth route for
   this). Password change on the same page via `auth.Service.UserUpdatePassword`.
4. **API keys** (`/api-keys`) — create a scoped key (HTMX POST to `/auth/api-keys`; the raw key
   is shown once in a dialog), list, revoke (`hx-delete`).
5. **Sessions** (`/sessions`) — list active refresh-token sessions, revoke one or all.
6. **Impersonation** — from `/admin/users/{id}`, click Impersonate (POSTs to `/auth/impersonate`).
   A sticky banner appears site-wide; "Stop impersonating" posts to `/auth/impersonate/stop`.
7. **Admin — users** (`/admin/users`) — search, suspend/reactivate, and a detail page showing
   full profile info (name, locale, timezone, created/last-login/last-active), roles, effective
   permissions (derived transitively through roles), an admin-only `AppMetadata` "notes" field,
   auth history, and audit log.
8. **Admin — RBAC** (`/admin/roles`, `/admin/permissions`) — create/delete roles and permissions,
   grant/revoke permissions on roles, grant/revoke roles on users.

All of the above (except profile update/password change) POST/DELETE straight to EzAuth's
existing `/auth/*` routes — this app owns none of that business logic.

### Architecture note: HTMX + JSON responses

Every EzAuth route for API keys, sessions, admin/users, and RBAC responds with **JSON**, never
HTML — including the cookie-session "Form" variants. That means:

- **List/detail pages** call the underlying Go method directly (`auth.APIKeysList`,
  `auth.Service.UsersList`, etc.) and render Templ — there's no HTML response to fetch from
  EzAuth's own routes.
- **Destructive actions** (revoke, delete) hit EzAuth's routes directly via `hx-delete` with
  `hx-swap="delete"` — the response body doesn't matter, so no app code is needed.
- **Creates/grants** POST via `hx-post hx-swap="none"`, then trigger an `hx-get` refresh of an
  app-owned fragment route. API key creation is a special case: the raw secret is in the JSON
  response, so an `htmx:after-request` handler pulls it into a dialog before refreshing the list.

## Also available in EzAuth, not demoed here

- **OAuth2/social login** (Google, GitHub, Discord, generic OIDC, ...) — needs real provider
  credentials to configure.
- **MFA (TOTP)** and **WebAuthn/passkeys** — need an authenticator app or platform passkey
  support to demo interactively.
- **SMS OTP** — needs a Twilio-compatible account.
- **Organizations / multi-tenancy** — this example intentionally models a single-tenant app.
- **Invitations** — overlaps with the open registration flow already demoed here.

## Verification

```sh
make test
```

Table-driven tests cover `internal/bootstrap` (admin/demo-user seeding: idempotency, that
re-seeding never touches an existing account's password or profile edits) against a real
temp-file SQLite `*ezauth.EzAuth` instance — everything else is EzAuth's own tested code.

Manual walkthrough: log in as `APP_ADMIN_EMAIL` (password from the startup logs), open
`/admin/users`, manage one of the seeded demo users — suspend/reactivate, grant a role, view
their auth history and audit log, impersonate them, then check `/api-keys` and `/sessions` while
impersonating.
