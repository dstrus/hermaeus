# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Stack

Organized so that each concern has its own place, with explicit contracts between them, to keep human review and maintenance easy.

- **Backend:** Go, organized by domain under `internal/` (for example `characters`, `memories`, `bios`, `settings`, `prompts`, `plugins`, `records`). Shared infrastructure gets its own packages (`server`, `install`, `store`, `textfile`, `gamecheck`), and `cmd/hermaeus` holds startup. Each domain owns its logic, typed request and response structs, routes, and tests. Only `store` touches SQLite, and only `textfile` rewrites files on disk.
- **Frontend:** Svelte 5 and TypeScript, built with Vite; SvelteKit in static/SPA mode for folder-based routing. Feature folders mirror the Go domains. TypeScript types are generated from the Go structs so contract drift breaks the build.
- **Delivery:** the built frontend is embedded in the Go binary, which ships as one program per OS (Windows, macOS, Linux). The SQLite driver is pure Go, so cross-compiling needs no C toolchain. Node is needed only to build.

## Users

Primary user: the author, a SkyrimNet player on Skyrim VR with a heavy mod list (around 1,700 mods), where launching the game and loading a save takes minutes and means wearing a headset the whole time. The author wants to view and adjust SkyrimNet's data and settings without launching the game.

Possible later audience: other SkyrimNet players (Skyrim SE, AE, or VR), if Hermaeus is released. Release is an option to keep open, not a commitment. Decisions should not shut it off: avoid assumptions tied to the author's machine, paths, mod manager, or setup, and keep the app understandable to someone who didn't build it.

## Product Purpose

Hermaeus is an unofficial, offline dashboard for SkyrimNet, named for the Daedric Prince who hoards knowledge and memories. It reads and edits SkyrimNet's databases, bios, settings, and content directly from the player's installation while the game is closed. It is not a 1:1 replacement for SkyrimNet's in-game dashboard; it covers what players most want to inspect or change offline. Success means a player can find and safely change what they came for without starting Skyrim.

## Positioning

SkyrimNet's own dashboard requires the game to be running. Hermaeus works with the game closed, reading the player's own files, and treats every write conservatively: backups before changes, single-transaction saves, minimal-diff file rewrites in SkyrimNet's own formats, and refusal to save while Skyrim is running.

## Operating Context

- Runs on the player's machine next to their mod manager setup: Mod Organizer 2 (beside `ModOrganizer.exe`, with files often in `overwrite/SKSE/Plugins/SkyrimNet`) or Vortex/manual installs (beside `Skyrim.exe`/`SkyrimVR.exe`).
- Works against SkyrimNet's folder (`data`, `config`, `logs`, `saves`, `screenshots`), its mod folder, installed packs and plugins, and MO2 profile mods.
- Data is organized per playthrough (save ID); players may have several tabs open on different playthroughs.
- Players move between Hermaeus and SkyrimNet's in-game dashboard; shared state (pins, dashboard edits, `content-registry.json`) must stay compatible with SkyrimNet's own formats.
- Vortex reports Hermaeus's file changes as external changes on the next deploy, as it does for dashboard edits.

## Capabilities and Constraints

Capabilities: characters (overview, stats, memories, diaries, events, re-identified NPCs, pins, Universal Translator speech patterns), save-specific bios with history and restore, events, knowledge and knowledge packs, OmniSight screenshot descriptions, bard songs with playback, quick toggles, settings compared with defaults, models and an OpenRouter model picker, AI provider keys, prompts, plugins (enable and reorder), YAML content records with forms, and overrides.

Constraints:

- **Local web UI in a browser:** one program per OS that serves only `127.0.0.1` and opens the browser; every view has its own bookmarkable address that starts with the playthrough's save ID, and Back and Forward work.
- **Unofficial:** not part of SkyrimNet and not affiliated with or endorsed by Fateless.Ai. It ships none of SkyrimNet's files and reads only the player's own installation.
- **Safety:** writes are refused while Skyrim is running (browsing still works); backups precede changes; secrets (API keys, tokens, passwords) are never sent to the browser, only whether they are set; requests carry a per-session token.

Scope: all of these. The author has used each of them offline in a real playthrough. Characters and bios are used broadly; settings and content (prompts, plugins, records) only in a small subset. The full feature set stays, but the interface should lead with the commonly used parts and keep the rest reachable without cluttering everyday use. Which settings and content make up that common subset isn't recorded yet; settle it before designing those areas.

Terminology: SkyrimNet, playthrough/save ID, NPC, memory, diary entry, event, save bio vs default bio, dynamic vs baseline bio, knowledge pack, dashboard edits, layer, override, plugin, trigger, action, filter, voice effect, quick toggle.

## Brand Commitments

- Name: Hermaeus, after Hermaeus Mora.
- Voice: plain and precise, with a dry, slightly wry edge. It says exactly what is saved, where, and when, and what isn't possible and why.
- The unofficial status and SkyrimNet/ZeniMax trademark disclaimers stay.
- No visual identity (themes, palette, icon) is committed yet.

## Evidence on Hand

- No user testimonials, download counts, or reviews are on hand; don't invent any.

## Product Principles

1. **The player's data comes first.** Every write is reversible, minimal, and in SkyrimNet's own format; when in doubt, refuse and explain.
2. **Offline is the point.** Anything that only works with the game running doesn't belong, and anything that doesn't need it shouldn't make the player launch Skyrim.
3. **Say what actually happens.** Name the file, the backup, and the consequence; don't soften or guess.
4. **Fit alongside SkyrimNet, not over it.** Shared state stays compatible with SkyrimNet's dashboard, and Hermaeus never presents itself as official.
5. **Built for one, releasable to many.** Serve the author's own setup first, but nothing should hard-code that setup or need the author to explain it.
