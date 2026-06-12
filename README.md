# Sliver GUI

A [Wails](https://wails.io) (Go + Svelte) desktop GUI for the [Sliver](https://github.com/BishopFox/sliver) C2 framework.

**Currently this is a personal demo project. Things are broken, half-finished, and will change without warning.** No support, no issues, no PRs at this time

## Build it

You need [Wails v2](https://wails.io/docs/gettingstarted/installation) (which needs Go and Node).

```sh
wails dev      # run with hot-reload
wails build    # produce a binary
```

## Layout

- `*.go` - Go backend, `package main`, grouped by domain (`app`, `connection`, `console`, `agents`, `server`, …). Wails binds the `App` struct's methods to the frontend.
- `frontend/src/lib/` - Svelte 5 app: `components` (UI primitives), `features` (views/panels), `api` (wrappers over generated Wails bindings), `stores` (shared state).

## Quality checks (optional)

```sh
make analyze   # go vet + staticcheck + frontend lint/typecheck/dead-code
make build
```
