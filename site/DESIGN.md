---
name: Hermaeus (web site)
description: The Hermaeus web site, set as one more menu of the game you just closed, larger and bolder than the app.
colors:
  accent: "#7fb8a0"
  on-accent: "#0e1a15"
  ground: "#0b0b0b"
  smoke: "#1c1c1c"
  select: "rgb(242 239 230 / 0.07)"
  ink: "#f2efe6"
  muted: "#a8a293"
  rule: "#2a2a2a"
  rule-strong: "#3a3a3a"
  changed: "#d2996b"
  warn: "#d9a441"
  accent-light: "#2a5f86"
  on-accent-light: "#eef4f9"
  ground-light: "#e4e3df"
  smoke-light: "#f3f2ef"
  select-light: "rgb(27 26 23 / 0.06)"
  ink-light: "#1b1a17"
  muted-light: "#5d5a52"
  rule-light: "#cdcbc4"
  rule-strong-light: "#b3b0a7"
  changed-light: "#9a4a20"
  warn-light: "#8a5a00"
typography:
  display:
    fontFamily: "Barlow Condensed, Jost, sans-serif"
    fontSize: "clamp(3.5rem, 2rem + 5vw, 6rem)"
    fontWeight: 600
    lineHeight: 0.9
    letterSpacing: "0.06em"
  headline:
    fontFamily: "Barlow Condensed, Jost, sans-serif"
    fontSize: "clamp(2.25rem, 1.5rem + 3vw, 3.75rem)"
    fontWeight: 600
    lineHeight: 1
    letterSpacing: "0.12em"
  entry:
    fontFamily: "Jost, system-ui, sans-serif"
    fontSize: "clamp(1.375rem, 1rem + 1.5vw, 2.125rem)"
    fontWeight: 400
    lineHeight: 1.45
  title:
    fontFamily: "Barlow Condensed, Jost, sans-serif"
    fontSize: "1.375rem"
    fontWeight: 600
    letterSpacing: "0.12em"
  lead:
    fontFamily: "Jost, system-ui, sans-serif"
    fontSize: "1.375rem"
    fontWeight: 400
    lineHeight: 1.35
  body:
    fontFamily: "Jost, system-ui, sans-serif"
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.55
    fontFeature: "tnum"
  label:
    fontFamily: "Barlow Condensed, Jost, sans-serif"
    fontSize: "0.8125rem"
    fontWeight: 500
    letterSpacing: "0.12em"
  button:
    fontFamily: "Barlow Condensed, Jost, sans-serif"
    fontSize: "1.125rem"
    fontWeight: 500
    lineHeight: 1.2
    letterSpacing: "0.12em"
  hint:
    fontFamily: "Barlow Condensed, Jost, sans-serif"
    fontSize: "1rem"
    fontWeight: 500
    lineHeight: 1.5
    letterSpacing: "0.12em"
rounded:
  base: "3px"
spacing:
  "1": "0.25rem"
  "2": "0.5rem"
  "3": "0.75rem"
  "4": "1rem"
  "5": "1.5rem"
  "6": "2rem"
  "7": "2.5rem"
  "8": "4rem"
  "9": "clamp(5rem, 3rem + 6vw, 9rem)"
  gutter: "clamp(1rem, 0.25rem + 3.5vw, 3.5rem)"
components:
  button:
    backgroundColor: "transparent"
    textColor: "{colors.ink}"
    typography: "{typography.button}"
    rounded: "{rounded.base}"
    padding: "0.75rem 1.5rem"
  button-primary:
    backgroundColor: "{colors.accent}"
    textColor: "{colors.on-accent}"
    typography: "{typography.button}"
    rounded: "{rounded.base}"
    padding: "0.75rem 1.5rem"
  theme-option:
    backgroundColor: "transparent"
    textColor: "{colors.muted}"
    typography: "{typography.label}"
    padding: "0.5rem 0.75rem"
  theme-option-selected:
    backgroundColor: "{colors.accent}"
    textColor: "{colors.on-accent}"
  section-link:
    textColor: "{colors.muted}"
    typography: "{typography.button}"
  hint-strip:
    backgroundColor: "{colors.ground}"
    textColor: "{colors.muted}"
    typography: "{typography.hint}"
    padding: "0.5rem clamp(1rem, 0.25rem + 3.5vw, 3.5rem)"
  hint-key:
    textColor: "{colors.ink}"
    rounded: "{rounded.base}"
    padding: "0 0.4em"
  screen-frame:
    backgroundColor: "{colors.ground}"
    rounded: "{rounded.base}"
  download-row:
    textColor: "{colors.ink}"
    typography: "{typography.headline}"
    padding: "1rem 0.75rem 1rem 2rem"
  download-row-mine:
    backgroundColor: "{colors.select}"
  item-row:
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    padding: "0.75rem"
  item-group-head:
    textColor: "{colors.accent}"
    typography: "{typography.label}"
    padding: "2rem 0.75rem 0.5rem"
---

# Design System: Hermaeus (web site)

## Overview

**Creative North Star: "The Game's Own Menus, Set Larger"**

The site is the app's world, the one recorded in the repository's root `DESIGN.md`, rendered bolder for a page that has to persuade. It keeps the app's token set role for role (the same smoked-glass ground, the same ink, hairlines, verdigris and frost blue accent, copper, 3px corners, and the same two typefaces), so a visitor who downloads Hermaeus opens a program that looks like the page they just left. What changes is scale and pacing: the app is a dense menu read at 15px; the site is a sequence of sections read top to bottom, with type steps up to 6rem and gaps between sections of up to 9rem.

The page leads with content, not a pitch: a diary entry, one NPC's own words at reading scale, with the app's diary dateline above it and a locus row below that says where it's kept. The name and the offer stand beside it behind a strong hairline. Real screens of the app follow, each with a sticky caption, and the one nearest the middle of the window is lit while the others dim. The rest of the app is listed in the app's own item table, the safety rules sit under hairlines, and the downloads are one row per system, with the visitor's own system first. The game's button-hint strip runs along the bottom of the window, and every hint in it is a working shortcut.

Where the site shares the app: every color token and its light counterpart; Jost for reading and tracked Barlow Condensed caps for labeling; 1px hairlines in two strengths; one 3px radius; the diamond marker; the hint strip with real shortcuts and key caps; the outlined button with an accent-filled primary; the segmented theme switch; the item table with accent group headings; copper only on what the player changed. Where the site departs: a longer type scale (display, headline, and the diary entry step are site-only), a 1rem body instead of 15px, larger buttons, a section rhythm with no app equivalent, full-width sections instead of a shell with a sections column and details pane, the diary entry as the hero, framed screens with scroll dimming, downloads ordered by the visitor's operating system, and the diamond widened from "the selected row" to "the selected thing" (the lit screen, the visitor's system).

**Key Characteristics:**
- The app's smoked-glass ground and token set, unchanged, in both themes.
- Larger steps: a 6rem name, 3.75rem section titles, and a diary entry at up to 2.125rem.
- One section rhythm: a caps title, a muted lead, then the material, with large gaps between sections.
- Content first: the diary entry is the hero, and real screens are the proof.
- Hairlines, not cards: no section, screen, or row is a filled tile, and nothing has a shadow.
- The bottom hint strip, D, S, and N, each a real shortcut.

## Colors

The app's near-monochrome palette, unchanged: off-white on near-black in dark, near-black on pale smoke in light, with one accent that changes material with the theme, verdigris in dark and frost blue in light. Every value matches the app's role for role; `data-theme` on the root is "dark", "light", or absent to follow the OS.

### Primary
- **Verdigris** (`accent` / `accent-light`): selection and action. The primary download button's fill, the chosen theme option, the lit screen's title and its diamond, the diamond on the visitor's system, group headings in the item table, the "See it in Hermaeus" text link, link underlines (mixed 75% into the ground at rest, full on hover), outline buttons' hover border, focus rings, text selection, and the caret.
- **Accent Ink** (`on-accent` / `on-accent-light`): text and the diamond on an accent fill.

### Secondary
- **Copper Changed** (`changed` / `changed-light`): what the player changed from what SkyrimNet ships, as the app shows it. On the site it appears only as the words "what you changed" in the Settings row of the item table, because that is where the app uses it.

### Neutral
- **Menu Black** (`ground` / `ground-light`): the outer ground and edge of the radial gradient, the hint strip's opaque fill, and the backing of each screen frame.
- **Smoke Heart** (`smoke` / `smoke-light`): the lighter center of the fixed radial gradient (an ellipse 70% by 60% at 65% 25%, fading to the ground at 75%).
- **Pane Glass** (`panel` / `panel-light`): the fill of each "By the way…" aside.
- **Selected Row Haze** (`select` / `select-light`): the visitor's own system in the download rows, and the backing of inline file names.
- **Parchment Ink** (`ink` / `ink-light`): reading text, the diary entry, titles, key caps, the dateline's author, and the visitor's own system name.
- **Weathered Ink** (`muted` / `muted-light`): leads, captions, the dateline's day and place, locus terms, section links and hints at rest, the theme options at rest, the footer, and provenance lines.
- **Hairline** (`rule` / `rule-light`): the top bar's underline, the locus row's top edge, item rows, download rows, the footer's top edge, and the hint strip's top edge.
- **Strong Hairline** (`rule-strong` / `rule-strong-light`): the offer's left edge in the hero, the dateline's underline, the rule above each screen caption and each safety rule, column-head and group-head underlines, the top of the download list, screen frames, button and switch borders, and key caps.

### Status
- **Amber Warning** (`warn` / `warn-light`): only the quoted state "Skyrim running · read-only" in the safety rules, as the app shows it.

### Named Rules
**The Same Token Set Rule.** The site doesn't define colors of its own. A color the site needs is a role the app already has, with the app's dark and light values; a new role ships in both, as a pair, or not at all.

**The Changed Rule.** Copper means the player changed this from what SkyrimNet ships, and the site shows it only where the app does. It is never a highlight, a warning, or a way to make something stand out.

**The Scarce Accent Rule.** The accent marks what's selected, what's actionable, and group headings inside a list. The largest accent fill is the primary download button. One exception, by request: the hook's promise, "with Skyrim closed", is set in the accent, as text and never as a fill.

## Typography

**Display Font:** Barlow Condensed 500 and 600 (with Jost, sans-serif), self-hosted via @fontsource
**Body Font:** Jost 400, 500, and 600 (with system-ui, sans-serif), self-hosted via @fontsource

**Character:** The app's pairing, at the scale of a page: tracked condensed capitals are the game's menu voice, now large enough to carry the name and section titles; Jost carries everything that's read, from the diary entry at reading scale down to the fine print. The site uses no monospace face.

### Hierarchy
- **Display** (Barlow Condensed 600, up to 6rem, line-height 0.9, 0.06em tracking, uppercase): the name in the hero only. It fills the offer column and never passes it (capped at 24% of the column's inline size), with its tracking halved from the label tracking because it's that large.
- **Headline** (Barlow Condensed 600, up to 3.75rem, line-height 1, label tracking, uppercase): section titles, and each system's name in the download rows.
- **Entry** (Jost 400, up to 2.125rem, line-height 1.45, 32ch measure): the diary entry in the hero, and nothing else.
- **Title** (Barlow Condensed 600, 1.375rem, uppercase): screen captions, safety rule names, and the name in the top bar.
- **Lead** (Jost 400, 1.375rem, line-height 1.35): the hook under the name, whose promise ("with Skyrim closed") is set in Jost 600 in the accent and never breaks across lines.
- **Body** (Jost 400, 1rem, line-height 1.55, tabular figures): prose, captions, table cells. Section leads run at 1.125rem in muted ink with a 60ch measure; rule text holds 44ch. The fine print under the downloads runs at 0.8125rem: first the requirement ("Requires SkyrimNet Beta 25 or newer.") in full ink, then the disclaimer in muted ink, opening with "Unofficial." in Jost 600 and full ink.
- **Button** (Barlow Condensed 500, 1.125rem, label tracking, uppercase; 600 on the primary): buttons and section links.
- **Hint** (Barlow Condensed 500, 1rem, label tracking, uppercase): the hint strip.
- **Label** (Barlow Condensed, 0.8125rem, 0.12em tracking, uppercase): the dateline's day and place, locus terms, column heads, group heads, theme options, and provenance lines.

### Named Rules
**The Label Versus Read Rule.** As in the app: anything that labels is tracked condensed caps; anything that's read is Jost in sentence case. A note inside a button (such as "64-bit") switches back to Jost at 0.8125rem with no tracking.

**The One Entry Rule.** The entry step is for an NPC's own words at reading scale. It is not a pull-quote style for marketing copy.

## Layout

Sections are full width, inset by a fluid gutter (1rem to 3.5rem), on the one fixed gradient. There is no sections column or details pane; the top bar holds the mark and name at left, then the section links and the theme switch at right, over a hairline.

The hero is a two-column grid filling the first viewport above the hint strip: the diary entry at left (up to 46rem) and the offer in a 19 to 26rem column at right (28rem from 80rem up), separated by a column gap of up to 6rem and the offer's strong-hairline left edge. Under 60rem the columns stack, the entry first and the offer after it, without the left edge.

**The Section Rhythm Rule.** Every section after the hero is the same: the largest spacing step above it (5 to 9rem), a caps headline with a muted lead under it (up to 52rem wide), a 4rem gap, then the material. Sections don't get backgrounds or borders of their own to set them apart; the rhythm does that.

The screens list stacks each screen with a 5 to 9rem gap. Each is a caption column (14 to 20rem) beside the frame, with the caption sticky at 2rem from the top; under 60rem the caption sits above the frame and stops sticking. The safety rules flow in an auto-fit grid of columns at least 22rem wide. The download rows are a single 48rem list. The item table caps at 64rem; under 30rem its rows become stacked name-and-description pairs.

The hint strip is fixed along the bottom, right-aligned. It hides under 60rem and on coarse pointers, since its shortcuts are for keyboards.

## Elevation & Depth

The site is flat, like the app. Depth comes from the smoked-glass gradient, from hairlines, and from the screens themselves, which are opaque images in hairline frames. There are no shadows anywhere on the site, including on screens and buttons.

The one depth effect is focus by light: while the visitor scrolls the screens, the screen nearest the middle of the window is lit at full strength and the other screens dim and desaturate (saturation 0.4); their captions never dim, so every caption keeps AA contrast, easing over 600ms on `cubic-bezier(0.16, 1, 0.3, 1)`. The dimming turns on only when script can observe the scroll, switches off under 60rem, and drops its transition under reduced motion.

### Named Rules
**The No Lift Rule.** Nothing on the page floats, so nothing has a shadow. Emphasis comes from light (the lit screen), the accent, and the diamond.

## Shapes

Nearly square corners, as in the app: one 3px radius on buttons, the theme switch, key caps, and screen frames; tables, rows, and sections are square. The only drawn geometry is the app's diamond, a 10-unit SVG rendered at 9px in the accent, and the mark. The diamond marks the selected thing: the lit screen's title, the visitor's own system in the download rows (in a column of its own, as in the app's tables), and inside the primary download button (in accent ink). It is hidden everywhere else.

## Components

### Buttons
Outlined caps that light up in the accent, larger than the app's.
- **Shape:** gently squared (3px radius), 1px strong-hairline border, padding 0.75rem by 1.5rem.
- **Default:** transparent, ink text, button typography; hover turns the border the accent.
- **Primary:** accent fill and border, accent-ink text at weight 600, with the diamond. One per group: the download for the visitor's system.
- **Note:** a short Jost note ("64-bit") sits at the button's far end.
- **Focus:** a 2px accent outline offset by 2px, on every control.

### Downloads
**The Your System First Rule.** The download for the visitor's own operating system leads, both in the hero and in Download, with the diamond on it. In the hero it is the one primary button and the only one that says "Download for"; the others give the system's name alone. In Download it moves to the top row, gets the `select` haze and an ink name, and its button becomes primary. Without script, or on a system it doesn't recognize, Windows leads in the hero and no row in Download is marked.

### Theme Switch
The app's segmented choice: a 3px outlined group of label caps (Auto, Dark, and Light) in muted ink; the chosen one fills with the accent in accent ink at weight 600. The choice is remembered and applied before first paint.

### Navigation
Section links in the top bar use button caps in muted ink, turning ink on hover; they hide under 60rem. The footer's links (Source, Releases, Issues) are body-size caps.

### Diary Entry (signature)
The hero's content: the app's diary dateline (the author in ink at body size, then the day, place, and mood in muted label caps joined by middots, over a strong hairline), the entry in the entry step, then a locus row over a hairline, with muted caps terms beside Jost values. A provenance line in muted label caps says the playthrough is made up. Under it, at the foot of the first window, an outlined caps control in the accent ("Why does this exist?") with a drawn down-arrow that nudges three times after load, unless reduced motion is set, leads to the next section.

### Screens
A screen of the app in both themes, showing the one that matches the page, in a 1px strong-hairline frame with 3px corners and no shadow. Each has a caption with a title-step caps name over a strong hairline and a muted description; the screen dims until it's lit, its caption never does; when lit, its name turns the accent and shows the diamond. A provenance line follows the list.

### By the Way… Aside
Tangents, caveats, and jokes pulled out of the main flow, as the MGO tutorial site's asides are: the app's details-pane material (Pane Glass fill, a strong hairline, 3px corners), a short title in ink at the title step in label caps, then muted Jost. The Why section stacks two beside its story; the safety section ends with one. Load-bearing text (download steps, safety rules) never goes in an aside.

### Item Table
The app's item table at page scale: muted label-caps column heads over a strong hairline, group heads in the accent with a strong hairline, page names in button-size caps, descriptions in Jost, and a hairline under each row. No row hover or selection; the table is read, not used.

### Button-Hint Strip (signature)
As in the app: fixed along the bottom on opaque ground, a hairline above, hints right-aligned. Each hint is a button, a key cap (1px strong-hairline border, 3px radius, ink, at least 1.6em wide) then the action in muted caps, turning ink on hover. The site has three: D jumps to Download and focuses its first button, S jumps to Safety, and N moves to the next section. Every hint is a real shortcut, ignored while a modifier is held or focus is in a field, and clicking it does the same thing. Whichever way a hint runs, its key cap presses in: it dips 2px and lights in the accent for a moment, unless reduced motion is set.

**Loading-screen tips.** The strip's left end shows a tip, as Skyrim's loading screens do: muted Jost at 0.8125rem, in a random order per visit, the next on a click or whenever a new section reaches the middle of the window, as Skyrim shows a new one at each loading screen (fading in from a slight blur). None repeats until every one has been shown, and changes aren't announced to screen readers, since scrolling would set them off at every section. Every tip is a true fact about Hermaeus in Davey's voice, kept short enough to fit on one line at 75rem; below that width the tip is hidden, and the strip keeps only the hints.

## Do's and Don'ts

### Do:
- **Do** use the app's tokens role for role, in both themes, and add a new color only as a dark and light pair.
- **Do** give every section the same rhythm: the largest spacing step above it, a caps headline and muted lead, then the material.
- **Do** show the real app in hairline frames, in both themes, and label made-up data as made up.
- **Do** put the diamond on the selected thing, and only there.
- **Do** lead the downloads with the visitor's own system.
- **Do** back every hint in the strip with a working shortcut, and hide the strip where there's no keyboard.

### Don't:
- **Don't** use cards, filled tiles, or section backgrounds; separate with hairlines and the section rhythm.
- **Don't** put a shadow on anything, screens included.
- **Don't** use copper anywhere the app wouldn't, or as a highlight.
- **Don't** add a second accent, or use the accent for decoration.
- **Don't** set marketing copy in the entry step; it is for an NPC's own words.
