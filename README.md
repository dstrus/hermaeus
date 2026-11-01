# Hermaeus

*An unofficial dashboard for SkyrimNet that does not require the game to be running*

Yes, it's vibe-coded, but at least it's vibe-coded by a professional software engineer with lots of experience. 🤷🏻‍♂️

## Who?

**Hermaeus Mora** is the Daedric Prince of knowledge, memory, fate, etc.

If you mean the mean the weirdo behind this project, it's me, Davey! I've been a software engineer for somewhere in the neighborhood of a quarter-century, and I play way too much Skyrim VR. I spend _even more_ time fiddling with it.

## What?

Hermaeus reads and edits SkyrimNet's data and settings from your own installation while Skyrim is closed. (It's read-only if the game is running.) It's a single program that runs a small web server on your machine and opens itself in your browser.

Hermaeus is not a 1:1 replacement for the SkyrimNet Dashboard. A number of functions really do require the game to be running, and my goal here is not to offer a comprehensive alternative, but simply to provide access to the settings I am (and perhaps you are) most likely to want to adjust offline.

It works with SkyrimNet but isn't part of it, and isn't affiliated with or endorsed by Fateless.Ai. It includes none of SkyrimNet's files.

## Why?

I made this for myself. Frankly you should use the actual SkyrimNet Dashboard unless you have a compelling reason not to (which I think I do).

SkyrimNet's dashboard is excellent, but I play Skyrim VR ([with ~1700 mods](https://synergyvr.org/mgo)). Launching the game is not exactly instantaneous. It can take several minutes to launch the game and load a save, and I need to wear my headset the whole time. (Yes, I know that there are headless options for running Skyrim VR. No, that doesn't make it _not_ a hassle.)

The things I tend to change are in files (SQLite databases, structured text files) that can be edited manually without an interface that knows anything about SkyrimNet, but finding the right file (and the right _copy_ of that file) to edit isn't always straightforward. So an interface that knows a thing or two is pretty helpful.

# How?

## Installing it (Windows)

**MO2 users:** Put `Hermaeus.exe` in your MO2 installation folder (the same folder as `ModOrganizer.exe`).

**Everyone else (including Vortex users)**: Put `Hermaeus.exe` in your actual Skyrim installation folder (the same folder as `Skyrim.exe` or `SkyrimVR.exe`).

## Running it

For the Windows version, just run `Hermaeus.exe`.

If you want to run it on a different platform or from a terminal, there are fancy command-line switches.

```
hermaeus [-dir <SkyrimNet folder>] [-mod-dir <SkyrimNet mod folder>] [-port 8765] [-backups <folder>] [-no-browser] [-version]
```

With Mod Organizer 2, SkyrimNet's files stay in its mod folder while it writes to `overwrite`. Hermaeus should find all of that stuff automatically if you run it from the suggested location. Failing that, and without `-dir`, Hermaeus uses the folder you last chose, or looks next to the program and in the current directory (including `SKSE/Plugins/SkyrimNet`, `Data/SKSE/Plugins/SkyrimNet` and `overwrite/SKSE/Plugins/SkyrimNet` below them). Failing _that_, it asks in the browser.

## Safety

- Hermaeus will not save any changes while Skyrim is running, because SkyrimNet keeps its database open. You can still browse.
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
| `web/src/lib/copy/` | The words the app shows, in `en.yaml`, grouped by page; `web/copy.ts` turns them into a typed module (`messages.ts`, generated) as the frontend builds |
| `web/src/lib/ui/` | The look and its shared parts: theme tokens (`theme.css`), the keyboard-shortcut hint strip, sortable heads, row selection |
| `web/src/lib/<domain>/` | Each domain's API calls and components, mirroring `internal/<domain>/` |
| `web/src/lib/api/` | The API client, and TypeScript types generated from Go (`types/`, don't edit) |
| `tygo.yaml` | Which Go types become TypeScript types |

Domain packages own their logic, request and response types, routes, and tests. Only `store` touches SQLite, and only `textfile` writes files SkyrimNet reads. The visual system is recorded in [DESIGN.md](DESIGN.md).

## License

Hermaeus is released under the [MIT License](LICENSE). Copyright (c) 2026 Davey Strus. It includes other people's software and fonts, each under its own license; [THIRD_PARTY_NOTICES.txt](THIRD_PARTY_NOTICES.txt) lists them with their license texts, and releases include it.

The Elder Scrolls V: Skyrim is a trademark of ZeniMax Media Inc.