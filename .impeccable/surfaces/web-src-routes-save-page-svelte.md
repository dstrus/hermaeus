---
version: 1
slug: "web-src-routes-save-page-svelte"
primary_target: "web/src/routes/[save]/+page.svelte"
related_targets: ["web/src/lib/summary/Overview.svelte","web/src/lib/summary/Preview.svelte","web/src/routes/+page.svelte"]
---

# Playthrough Overview

**Mode:** Operate. **Scope:** the playthrough's home (`/<save>`), which becomes an Overview; Characters moves to `/<save>/characters`; the Playthroughs list gains a details pane previewing the selected playthrough. Every existing API, deeper URL, and behavior stays.

**Audience and job:** the author between play sessions, deciding where to go: read what the story did last, see what SkyrimNet has stored, jump into Characters, History, Screenshots, or Bard songs, and glance at tuning (models, voice engine, toggles).

**Content:** the latest diary entries, the newest events grouped by game day, counts per section, companions and pinned characters, the dialogue model, the voice engine, and the quick toggles' state. **States:** a fresh playthrough with nothing stored; no diary yet; no player name; Skyrim running.

**Constraints:** both themes from one token set; keyboard paths with real hints; holds 1280 to 1600; stacks under 60rem.

## Direction contract

THESIS: The story leads as reading, like the game's Journal, and the numbers are the navigation. It refuses the dashboard home of stat tiles and a grid of cards.

OWN-WORLD: Smoked-glass ground, hairline rules, Jost prose at 65ch for the diary, Barlow Condensed caps for every label; one accent (verdigris in dark, frost blue in light) for links, group headings, bars, and the diamond.

STORY: The player reads what Gram wrote last, scans the newest events by game day, sees how much is stored, and steps into a section with Enter or a click.

FIRST VIEWPORT: Left column: the player's name in display caps, with the game date, days recorded, and when the last event was recorded on a label-caps line under it (stacked rather than beside, so the status never wraps against the name); "The latest diary" as three prose entries with datelines; "Lately" as an item table of the newest events grouped under accent day headings. Right: the details pane with "Stored" as a stat list whose rows are the section doorways (the diamond marks the one the arrow keys reach; Enter opens it), then Company as one run-in line of names, then Tuning with links to Models, Engines, and Prompts.

FORM: The journal spread, second on the ordered list of grounded structures; dealt and chosen over the roll's lead. seed:4ee0ec4a3a04ac63

FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance
