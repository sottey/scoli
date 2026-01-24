# Multi-Tenancy Plan (User Accounts + Isolation)

## Current state (baseline)
- The server is single-tenant and file-based. All notes live under a single `notesDir`.
- API is unauthenticated; the UI assumes direct access.
- Per-user settings are stored as `settings.json` under the notes root.
- Key entry points:
  - `internal/server/server.go` mounts the API at `/api/v1` with a single `notesDir`.
  - `internal/api/*` reads/writes directly under `Server.notesDir`.
  - `docs/API.md` explicitly states "Authentication: None."

## Goals
- Support multiple users with username/password authentication.
- Enforce hard data isolation between users.
- Allow phased rollout (manual user creation first; signup/reset later).

## Mandates (must-have to be “multi-tenant”)

### 1) Authentication + session handling
Must-have so each request is tied to a user.
- Add auth middleware to `/api/v1` (except `/health` and auth endpoints).
- Introduce a session mechanism (signed cookies or server-side session store).
- Add login endpoint (e.g., `POST /api/v1/auth/login`) that issues a session.
- Logout endpoint that clears the session.
- Password hashing (bcrypt/argon2id) and constant-time comparisons.

Likely touchpoints:
- `internal/server/server.go` for global middleware.
- `internal/api/router.go` to add auth routes + middleware.
- New `internal/api/auth.go` (login/logout handlers).

### 2) User data isolation (per-user storage root)
Must-have to prevent cross-user data access.
- Convert `Server.notesDir` into a *base* directory.
- Resolve a user-specific root for every request:
  - Example: `notesBaseDir/<user-id>/`
- All filesystem operations must use the user root:
  - notes, folders, files, settings, journal, tasks, tags, search.

Likely touchpoints:
- `internal/api/server.go` (resolvePath + all filesystem walks)
- `internal/api/settings.go` (settings file per user)
- `internal/api/journal.go`, `internal/api/tasks.go`, `internal/api/icons.go`

### 3) User identity storage
Must-have for authentication to work.
- Minimal user store with:
  - `id`, `username`, `password_hash`, `created_at`, `disabled`
- For a quick MVP, store users in a local file or SQLite DB.
  - A small SQLite DB is safer than JSON for concurrent writes.

Likely touchpoints:
- New `internal/auth` or `internal/store` package.
- `cmd/scoli` for admin CLI commands (create user, disable user).

### 4) UI gating for unauthenticated users
Must-have to avoid API failures and data leaks.
- Add a login screen in the UI.
- Block app initialization until authenticated.
- Store session via cookie; do not store raw credentials.

Likely touchpoints:
- `internal/ui/web/index.html` (login modal/screen)
- `internal/ui/web/app.js` (auth flow and API error handling)

## Strongly recommended (should-have)

### A) Admin tooling (manual user creation)
You asked to allow manual adds early.
- CLI commands:
  - `scoli users add --username --password`
  - `scoli users disable --username`
- Optional: `users list`
- Seed notes per user at creation (re-using existing seed logic).

Likely touchpoints:
- `cmd/scoli` (new subcommands)
- `internal/server` or `internal/auth/store` for user CRUD

### B) Rate limiting + brute-force protection
- Lockout or exponential backoff on failed login attempts.
- Per-IP request limits on auth endpoints.

### C) Logging and auditability
- Log auth events (login success/fail, user disabled).
- Include user id in request logs (not raw usernames in every line).

### D) Per-user settings scope in UI
- Keep browser localStorage keys namespaced by user
  - Example: `scoli:<user-id>:<key>`

## Nice-to-have (could-have / later)

### Signup + password reset
You mentioned these can be delayed.
- Self-serve signup flow (email or invite code).
- Password reset tokens and email delivery.
- UI for password change.

### Two-factor auth
- TOTP, WebAuthn, or email codes.

### Per-tenant limits & quotas
- Storage quotas, note count limits, retention policies.

## Proposed MVP scope (manual user adds)

### Backend
- Add a user store (SQLite preferred).
- Add login/logout endpoints.
- Add auth middleware.
- Map users to per-user filesystem root: `notesBaseDir/<user-id>/`.
- Update all filesystem access to use user-scoped root.

### Frontend
- Simple login screen (username + password).
- Session persistence via HTTP-only cookie.
- On 401, show login screen.

### CLI
- `scoli users add` for manual user provisioning.
- `scoli users disable` for access revocation.

## Design choices and tradeoffs

### Storage model
Option 1: Per-user filesystem roots (recommended for MVP)
- Minimal change to note handling (still filesystem).
- Need to ensure every path resolves against the user root.

Option 2: Database-backed notes
- Larger refactor, but enables granular permissions and efficient queries.

### Sessions
Option A: Signed cookie sessions
- Simpler, but requires secret rotation strategy.

Option B: Server-side sessions (SQLite/Redis)
- Better for invalidation and multi-instance deployments.

## Estimated effort (relative)
- MVP with manual user provisioning: medium-to-large (auth + refactor + UI).
- Full user lifecycle (signup/reset/2FA): large.
- DB-backed notes: very large.

## Concrete file touchpoints (likely)
- `internal/server/server.go` (middleware, base dir config)
- `internal/api/router.go` (auth routes, auth middleware)
- `internal/api/server.go` (path resolution, root isolation)
- `internal/api/settings.go` (user-scoped settings)
- `internal/api/journal.go` / `internal/api/tasks.go` / `internal/api/icons.go`
- `internal/ui/web/app.js` / `internal/ui/web/index.html`
- `docs/API.md` (update Authentication section when implemented)

## Open questions (to finalize design)
- Do you want to allow multiple users to share a browser session?
- Should admins be a separate role with elevated permissions?
- How should user data be stored (SQLite vs JSON)?
- Should auth be handled in-app or by a reverse proxy (SSO/basic auth)?
