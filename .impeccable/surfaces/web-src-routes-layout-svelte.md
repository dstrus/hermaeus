---
version: 1
slug: "web-src-routes-layout-svelte"
primary_target: "web/src/routes/+layout.svelte"
related_targets: ["web/src/routes/[save]/+layout.svelte","web/src/routes/[save]/npc/[uuid]/+layout.svelte"]
---

# App shell and character view

**Mode:** Operate. **Scope:** the shell (top bar, category column, and content area) and the character view (character table, Overview, Stats, Memories, and Diary). Every existing API, URL, and behavior stays.

**Audience and job:** the author at a desktop monitor between play sessions. First, tune before playing (quick toggles, then prompts), one glance and one click from any page. Second, fix what SkyrimNet stored: find the character, find the memory or entry, edit it. Whether Skyrim is running, and so whether saves are allowed, is always visible.

**Ranges:** about 70 characters (some unnamed); up to ~140 memories and ~40 long diary entries per character. States: Skyrim running (read-only, reason given), no folder, no playthroughs, empty search, save notices (copy updated, search vector pending), errors, retired identities, diary copies.

**Constraints:** dark and light themes from one token set, following the OS by default with a remembered switch in the top bar. Fonts bundled in the program (offline), with OFL notices. WCAG AA, visible focus, complete keyboard paths. Diary Markdown shown as written. Holds from 1280 to 1600 px; narrower stacks the categories as tabs.

## Direction contract

THESIS: Hermaeus is one more menu of the game you just closed: categories, a sortable item list, and a details pane, driven from the keyboard. It refuses the admin-dashboard default of gray sidebar, cards, and a blue button.

OWN-WORLD: Smoked-glass ground (near-black in dark, pale smoke in light), hairline rules, off-white or near-black text, one accent only for selection and action: verdigris in dark, frost blue in light. Tracked condensed caps (Barlow Condensed) for categories, headings and column heads; Jost for text; tabular figures. The ◆ marks the selected row.

STORY: The visitor sees their playthrough and its toggles first, flips what they need, then drops into a character's memories as an inventory-like table, edits in the details pane, and sees exactly what the save did.

FIRST VIEWPORT: Top bar: Hermaeus, quick toggles as one row, the playthrough and Skyrim's state at right. Left: the category column. Centre: the item table with sortable heads. Right: the details pane. Bottom right: the button-hint bar (E Edit, N New, Del Delete, / Find), backed by real shortcuts: the signature move.

FORM: The Game's Own Menus, first on the ordered list of grounded candidates (taken as the pick over the rolled assignment). seed:3f9dc9906b2e1a1d

FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, DESIGN.md, and every shipping raster carrying its provenance
