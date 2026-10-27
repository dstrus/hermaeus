# Hermaeus

*An unofficial dashboard for SkyrimNet that does not require the game to be running, named for the Daedric Prince who hoards knowledge and memories.*

Hermaeus reads and edits SkyrimNet's data and settings from your own installation while Skyrim is closed. It's a single program that runs a small web server on your machine and opens itself in your browser.

Hermaeus works with SkyrimNet but isn't part of it, and isn't affiliated with or endorsed by Fateless.Ai. It includes none of SkyrimNet's files. The Elder Scrolls V: Skyrim is a trademark of ZeniMax Media Inc.

## Running it

```
hermaeus [-dir <SkyrimNet folder>] [-port 8765] [-backups <folder>] [-no-browser] [-version]
```

Without `-dir`, Hermaeus uses the folder you last chose, or looks next to the program and in the current directory (including `SKSE/Plugins/SkyrimNet`, `Data/SKSE/Plugins/SkyrimNet` and `overwrite/SKSE/Plugins/SkyrimNet` below them). If it doesn't find one, it asks in the browser.

## Safety

- Saving is refused while Skyrim is running, because SkyrimNet keeps its database open. You can still browse.
- Before the first change to each database in a session, a copy is written to `hermaeus-backups` in the SkyrimNet folder (or to `-backups`).
- Every save happens in one transaction. If any part fails, nothing changes.
- The server only listens on `127.0.0.1`, and every API request has to carry a per-session token from the page, so other websites open in your browser can't reach your data.

## Building

You need Go 1.26 or later and Node 22.17 or later.

```sh
make build              # the frontend, then the program, as ./hermaeus
make test               # Go tests, then type-checking the frontend
make dev DIR=/path/to/SkyrimNet   # Hermaeus plus Vite's dev server at http://localhost:5173
make types              # regenerate the API's TypeScript types after changing them in Go
```

`go test ./internal/...` works without Node. Building the program, or `go vet ./...`, needs the frontend built first, since it's embedded from `web/build`.

## Layout

| Path | Purpose |
| --- | --- |
| `cmd/hermaeus/` | Flags, startup, and opening the browser |
| `internal/server/` | The web server: host and token checks, serving the frontend, mounting each domain's routes |
| `internal/httpapi/` | Turning typed Go functions into JSON handlers |
| `internal/apperr/` | Errors meant for the user, by kind; the server maps kinds to HTTP statuses |
| `internal/store/` | The only code that opens SkyrimNet's databases: playthroughs, backups, transactions |
| `internal/install/` | Finding the SkyrimNet folder |
| `internal/prefs/` | Hermaeus's own remembered settings |
| `internal/gamecheck/` | Whether Skyrim is running, per OS |
| `internal/characters/` | NPCs and the player: identities and pins |
| `internal/testenv/` | SkyrimNet folders with a real schema, for tests |
| `web/src/routes/` | The app's pages; each folder is a URL |
| `web/src/lib/<domain>/` | Each domain's API calls and components, mirroring `internal/<domain>/` |
| `web/src/lib/api/` | The API client, and TypeScript types generated from Go (`types/`, don't edit) |
| `tygo.yaml` | Which Go types become TypeScript types |

Domain packages own their logic, request and response types, routes, and tests. Only `store` touches SQLite.

## License

Copyright (c) 2026 Davey Strus.
