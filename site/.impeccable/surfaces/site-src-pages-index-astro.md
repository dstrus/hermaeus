---
version: 1
slug: "site-src-pages-index-astro"
primary_target: "site/src/pages/index.astro"
related_targets: []
---

# Landing page (site/src/pages/index.astro)

Mode: Persuade. Audience: SkyrimNet players deciding whether to install Hermaeus. Job: in under a minute, know what it is, that it's unofficial and careful with their files, and where to download. Action: per-OS download (Windows first) linking to the latest GitHub release of github.com/dstrus/hermaeus. Proof: real screens of the app running on a made-up playthrough, labeled as made up; the safety rules as the README states them. Constraints: static Astro, base /hermaeus, both themes (dark default from the app's world, light via prefers-color-scheme), WCAG AA, no testimonials or counts, no mention of any earlier version or a Nexus page.

## Direction contract

THESIS: The page opens on what SkyrimNet remembers, one NPC's own diary entry at reading scale, not on a product pitch; the category's headline-over-screenshot hero and feature-card grid are refused.

OWN-WORLD: The app's Game's Own Menus world, rendered bolder: smoked-glass radial ground on menu black (pale smoke in light), verdigris accent (frost blue in light), copper only for what the player changed, Barlow Condensed tracked caps for labels and display at hero scale, Jost for reading, 1px hairlines, 3px corners, the diamond marker, the bottom button-hint strip with real shortcuts. No cards, no shadows on sections.

STORY: Visitor reads Hjordis's entry, sees it is diary entry 143 in the playthrough's database, edited by the player; scrolls through the real app showing that entry's memory table, its edit pane, and the character; reads exactly what Hermaeus touches, backs up, and refuses; picks a download.

FIRST VIEWPORT: Top bar (mark, HERMAEUS, nav, Download). Left two-thirds: dateline in label caps over the entry in large Jost (~2.1rem, 30ch), a locus row beneath (entry number, where stored, importance, "Edited by you" in copper). Right column behind a strong hairline: HERMAEUS display (~6.5rem caps), the hook, three download buttons (Windows primary), the unofficial line. Hint strip fixed at the bottom: D Download, S Safety, arrow Next. "From a made-up playthrough" labels the entry.

FORM: What SkyrimNet Remembers, position 3 of 7 on my ordered list; seed:6251d4bb266ddc4a. Signature interaction: as the visitor scrolls the screens, the one in focus brightens and the others dim (raise from the declined streaming-wall challenger); every hint in the strip is a working keyboard shortcut.

FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance

## Recorded adaptations
- The locus says "Diary entry 9 of 12", the entry's place in Hjordis's diary in the made-up playthrough, not a row number.
- "Edited by you" in copper is dropped: the app doesn't mark an edited diary entry, so the page would claim what the app doesn't show. Copper appears instead where the app uses it, on what the player changed in settings, in the table of the rest of SkyrimNet.
- The hint strip's Next is N, since the arrow keys already scroll the page.
- On narrow screens the entry still leads; the name sits in the top bar above it and the offer follows.
- The download for the visitor's own system leads, in the first viewport and in Download, marked with the diamond.

## Unresolved
- The repository is public, but it has no releases yet, so the download links land on an empty Releases page until the first one is published.
- Each download opens the latest release's page; linking each system's file directly waits on the release file names.
