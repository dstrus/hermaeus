# Hermaeus

*An unofficial dashboard for SkyrimNet that does not require the game to be running, named for the Daedric Prince who hoards knowledge and memories.*

Hermaeus reads and edits SkyrimNet's data and settings from your own installation while Skyrim is closed. It's a single program that runs a small web server on your machine and opens itself in your browser.

Hermaeus works with SkyrimNet but isn't part of it, and isn't affiliated with or endorsed by Fateless.Ai. It includes none of SkyrimNet's files. The Elder Scrolls V: Skyrim is a trademark of ZeniMax Media Inc.

## Running it

```
hermaeus [-dir <SkyrimNet folder>] [-mod-dir <SkyrimNet mod folder>] [-port 8765] [-backups <folder>] [-no-browser] [-version]
```

With Mod Organizer 2, SkyrimNet's own files stay in its mod folder while what it writes lands in `overwrite`; Hermaeus looks for the mod folder next to `overwrite`, and `-mod-dir` names it when that fails. Without `-dir`, Hermaeus uses the folder you last chose, or looks next to the program and in the current directory (including `SKSE/Plugins/SkyrimNet`, `Data/SKSE/Plugins/SkyrimNet` and `overwrite/SKSE/Plugins/SkyrimNet` below them). If it doesn't find one, it asks in the browser.

## Safety

- Saving is refused while Skyrim is running, because SkyrimNet keeps its database open. You can still browse.
- Before the first change to each database in a session, a copy is written to `hermaeus-backups` in the SkyrimNet folder (or to `-backups`).
- Every save happens in one transaction. If any part fails, nothing changes.
- The server only listens on `127.0.0.1`, and every API request has to carry a per-session token from the page, so other websites open in your browser can't reach your data.

## Building

You need Go 1.26 or later and Node 22.17 or later.

```sh
make build              # the frontend, then the program, as ./hermaeus
make windows            # the program for 64-bit Windows, with its icon, as ./Hermaeus.exe
make notices            # THIRD_PARTY_NOTICES.txt, from the licenses of everything built into the program
make test               # Go tests, then type-checking the frontend
make dev DIR=/path/to/SkyrimNet   # Hermaeus plus Vite's dev server at http://localhost:5173
make types              # regenerate the API's TypeScript types after changing them in Go
```

`go test ./internal/...` works without Node. Building the program, or `go vet ./...`, needs the frontend built first, since it's embedded from `web/build`.

## Layout

| Path | Purpose |
| --- | --- |
| `cmd/hermaeus/` | Flags, startup, and opening the browser |
| `cmd/notices/` | Writes THIRD_PARTY_NOTICES.txt, with `web/third-party.ts`, which records the npm packages the frontend's build includes |
| `internal/server/` | The web server: host and token checks, serving the frontend, mounting each domain's routes |
| `internal/httpapi/` | Turning typed Go functions into JSON handlers |
| `internal/apperr/` | Errors meant for the user, by kind; the server maps kinds to HTTP statuses |
| `internal/store/` | The only code that opens SkyrimNet's databases: playthroughs, backups, transactions |
| `internal/install/` | Finding the SkyrimNet folder |
| `internal/prefs/` | Hermaeus's own remembered settings |
| `internal/gamecheck/` | Whether Skyrim is running, per OS |
| `internal/identity/` | NPCs SkyrimNet re-identified: resolving retired UUIDs to current ones |
| `internal/calendar/` | The game clock as the game's own dates, following mods that move the date |
| `internal/characters/` | The player and NPCs: who's listed, pins, each one's overview and stats |
| `internal/events/` | What SkyrimNet records happening in the game |
| `internal/memories/` | Each character's memories and diary entries |
| `internal/screenshots/` | OmniSight's screenshots: their descriptions, images, and thumbnails |
| `internal/songs/` | Bard songs: their words, and their audio |
| `internal/bios/` | Each character's bios: the save's dynamic and baseline bios, the default, history, and restore |
| `internal/layers/` | SkyrimNet's content layers: dashboard edits, enabled plugins in order, shipped defaults; the MO2 mod folder |
| `internal/prompts/` | SkyrimNet's prompts from every content layer, edited as the player's own copies |
| `internal/settings/` | SkyrimNet's settings files compared with their defaults, changed in place, and the quick toggles |
| `internal/templates/` | Checking the markup of SkyrimNet's prompt and bio templates |
| `internal/textfile/` | How files SkyrimNet reads are changed: atomic writes, backups, line endings kept |
| `internal/testenv/` | SkyrimNet folders with a real schema, for tests |
| `assets/` | The drawings behind the app icon and the mark in the top bar; `make icons` renders the icon files from them |
| `web/src/routes/` | The app's pages; each folder is a URL |
| `web/src/lib/nav.ts` | The sidebar's groups and their pages, and every page's URL: `/<save>/<group>/<page>` |
| `web/src/lib/ui/` | The look and its shared parts: theme tokens (`theme.css`), the keyboard-shortcut hint strip, sortable heads, row selection |
| `web/src/lib/<domain>/` | Each domain's API calls and components, mirroring `internal/<domain>/` |
| `web/src/lib/api/` | The API client, and TypeScript types generated from Go (`types/`, don't edit) |
| `tygo.yaml` | Which Go types become TypeScript types |

Domain packages own their logic, request and response types, routes, and tests. Only `store` touches SQLite, and only `textfile` writes files SkyrimNet reads. The visual system is recorded in [DESIGN.md](DESIGN.md).

## License

Hermaeus is released under the [MIT License](LICENSE). Copyright (c) 2026 Davey Strus. It includes other people's software and fonts, each under its own license; [THIRD_PARTY_NOTICES.txt](THIRD_PARTY_NOTICES.txt) lists them with their license texts, and releases include it.
