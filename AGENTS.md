# AI Agent Guide

## Layout Profile

This is the standard Sphere layout. It combines Protobuf/Buf, generated HTTP
handlers, Ent, Wire, Swagger, a dashboard API, local file storage, and a minimal
username/password application login. It intentionally has no Telegram or
WeChat dependency.

## Ownership Rules

Read `.sphere/layout.json` before changing files. Paths are classified as
`layout_owned`, `mixed`, or `generated`; every unmatched path is
`project_owned`. Never edit generated files by hand. Treat mixed files as
integration seams and preserve both layout wiring and project additions.

Application contracts belong in `proto/<domain>/v1`, business logic in
`internal/biz/<domain>`, HTTP implementations in `internal/service/<domain>`,
and Ent schemas in `internal/pkg/database/schema`. Do not put product-specific
logic into layout-owned helpers.

See `docs/LAYOUT_CONTRACT.md` for the complete authoring and synchronization
protocol, including legacy-project adoption and conflict handling.

## Workflow

Use the Makefile as the workflow contract:

- `make gen/all` regenerates Ent, Proto, Swagger, Wire, and mapping outputs.
- `make test` runs the Go tests.
- `make lint` runs non-mutating Go and Buf checks.
- `make check` verifies dependency, formatting, lint, and test state.
- `make build` builds the local binary.

After changing Proto, schemas, constructors, or provider sets, run
`make gen/all` before tests. A completed change must pass `make check` and
`make build`, and tracked generated files must have no unexplained drift.

## Authentication Extension

Password authentication resolves a local `User` and then issues a token in
`internal/service/api/auth.go`. A third-party provider should add its own Proto
contract and identity persistence, resolve or create a local user, and reuse
the token-response seam. Do not add provider SDKs to the standard layout.

### Session cookie

The dashboard session lives in two places that are not interchangeable: the
server-minted `auth_token` cookie (HttpOnly, `SameSite=Lax`, Secure behind TLS)
and whatever the client keeps for itself (the embedded dash page holds tokens in
memory only). Browser requests that cannot set headers — page loads, `<img>`,
EventSource — authenticate with the cookie through the auth middleware. Four
consequences worth knowing before touching auth code:

- **Only the server can delete the cookie.** `document.cookie` cannot overwrite
  or delete an HttpOnly cookie of the same name — the write is silently dropped.
  A client-side "logout" that only forgets the token leaves a live credential in
  the browser until it expires. Use `POST /api/auth/logout`.
- **Let the logout response land before navigating.** A navigation aborts
  in-flight requests, and the aborted response's `Set-Cookie` is lost — the
  session survives the logout that was supposed to end it.
- **The access token cannot be revoked early.** It is a stateless JWT, so
  revoking the session row only stops the next refresh; deleting the cookie is
  what makes logout take effect.
- **Do not derive the login state from a second source.** If a page gate trusts
  the cookie while the client trusts `localStorage`, the two disagree as soon as
  one of them is cleared, and the gate can bounce the login page straight back
  to the page that redirects to it: an infinite redirect that renders nothing.
  Keep the login page unconditionally reachable.
