# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Stack

Astro, building static HTML for GitHub Pages. User documentation is written as Markdown; Svelte components may be used for interactive parts. The site lives in `site/` of the Hermaeus repository, with its own `package.json`, apart from the app's build: the Go program embeds `web/build` only, and no Go files go under `site/`.

Served at `https://daveystrus.com/hermaeus`. Astro's `site` is `https://daveystrus.com` and its `base` is `/hermaeus`, so every link and asset path carries the base.

## Users

Primary: Skyrim players who use SkyrimNet (on Skyrim SE, AE, or VR) and are deciding whether to install Hermaeus. Many run large mod lists under Mod Organizer 2 or Vortex, where starting the game takes minutes, and want to see and change what SkyrimNet stores without launching it.

Secondary: people who already run Hermaeus and need the manual: setup with Mod Organizer 2 or Vortex, what each page does, and what it changes on disk. Possibly later: developers reading developer documentation.

## Product Purpose

The public web site for Hermaeus: a landing page that lets a SkyrimNet player understand what Hermaeus does and decide to download it, user documentation, and possibly developer documentation. Success means a player knows within a minute what Hermaeus is, that it's unofficial and safe with their data, and where to get it; and an installed user finds how to do what they came for.

Hermaeus itself is an unofficial, offline dashboard for SkyrimNet, named for the Daedric Prince who hoards knowledge and memories. It reads and edits SkyrimNet's databases, bios, settings, and content from the player's own installation while Skyrim is closed. The app's own product record is the repository's root `PRODUCT.md`.

## Positioning

SkyrimNet's own dashboard needs the game running. Hermaeus works with the game closed, from the player's own files, and treats every write conservatively: a backup before the first change, single-transaction saves, minimal changes to files in SkyrimNet's own formats, and refusal to save while Skyrim runs.

## Operating Context

- Visitors arrive from wherever SkyrimNet players talk and find mods; where Hermaeus is announced is not decided.
- Downloads are builds attached to GitHub Releases in the public repository: one program each for Windows, macOS, and Linux. A Nexus Mods page may follow after more testing; until it exists, the site doesn't mention or link one.
- The repository is public at `https://github.com/dstrus/hermaeus`. It has no releases yet, so the download links, which open the latest release, land on an empty Releases page until the first one is published.

## Capabilities and Constraints

- Static hosting only: no server code, no accounts, no forms that need a back end.
- Hermaeus requires SkyrimNet Beta 25 or newer. The site says so under the hero's downloads and at the top of the Download section.
- What Hermaeus can do is in the root `PRODUCT.md` and `README.md`; the site describes only what the app actually does, in its terms (SkyrimNet, playthrough, NPC, memory, diary entry, event, bio, knowledge pack, plugin, trigger, action, filter, voice effect, quick toggle).
- The site never describes Hermaeus as a rewrite, port, or successor of anything, and never mentions an earlier version.
- Hermaeus is released under the MIT License (`LICENSE` at the repository root).
- Undecided: developer documentation's scope, and where Hermaeus is announced.

## Brand Commitments

- Name: Hermaeus, after Hermaeus Mora. The mark (the eye of Hermaeus Mora among his tentacles) is in `assets/`: `hermaeus-mark.svg`, `hermaeus-icon.svg`, and `hermaeus-icon-small.svg`.
- Voice: Davey's own, as in the README and the MGO tutorial site: first person, conversational, and honest about Hermaeus's limits (including that most people should use SkyrimNet's own dashboard). Dry, self-aware humor goes in asides, parentheticals, and leads, never in download steps or safety rules, which stay plain and exact about what is saved, where, and when. No hype.
- The unofficial-status and trademark statements stay, as the README words them: Hermaeus works with SkyrimNet but isn't part of it, isn't affiliated with or endorsed by Fateless.Ai, and includes none of SkyrimNet's files; The Elder Scrolls V: Skyrim is a trademark of ZeniMax Media Inc.
- American English spelling and the serial comma.

## Evidence on Hand

- The app runs, so real screenshots can be taken. Taking them needs a SkyrimNet folder whose content is fit to publish: playthrough data shows character names and dialogue, so pick what appears deliberately.
- No testimonials, users, download counts, reviews, or press exist yet; don't invent any.

## Product Principles

1. **Show the real app.** Real screens and real wording from Hermaeus beat any claim about it.
2. **Earn trust on safety.** Say plainly what Hermaeus touches, what it backs up, and what it refuses to do, before asking anyone to point it at their saves.
3. **Unofficial, and says so.** Never look or sound like part of SkyrimNet or Fateless.
4. **Docs answer the real setup.** Mod Organizer 2, Vortex, and manual installs, on Windows first, with macOS and Linux covered.

## Accessibility & Inclusion

WCAG AA, as the app holds itself to: sufficient contrast in both themes, visible focus, complete keyboard paths, and documentation that reads well with a screen reader.
