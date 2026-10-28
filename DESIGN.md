---
name: Hermaeus
description: An offline dashboard for SkyrimNet that looks like one more menu of the game you just closed.
colors:
  gold: "#c9a24b"
  on-gold: "#141109"
  ground: "#0b0b0b"
  smoke: "#1c1c1c"
  panel: "#121212"
  glass-strong: "rgb(0 0 0 / 0.38)"
  select: "rgb(242 239 230 / 0.07)"
  ink: "#f2efe6"
  muted: "#a8a293"
  rule: "#2a2a2a"
  rule-strong: "#3a3a3a"
  warn: "#d9a441"
  danger: "#e57368"
  gold-light: "#7f5d12"
  on-gold-light: "#fdf8ec"
  ground-light: "#e4e3df"
  smoke-light: "#f3f2ef"
  panel-light: "#eceae6"
  glass-strong-light: "rgb(255 255 255 / 0.55)"
  select-light: "rgb(27 26 23 / 0.06)"
  ink-light: "#1b1a17"
  muted-light: "#5d5a52"
  rule-light: "#cdcbc4"
  rule-strong-light: "#b3b0a7"
  warn-light: "#8a5a00"
  danger-light: "#b3261e"
typography:
  display:
    fontFamily: "Barlow Condensed, Jost, sans-serif"
    fontSize: "2.5rem"
    fontWeight: 600
    lineHeight: 1
    letterSpacing: "0.12em"
  headline:
    fontFamily: "Barlow Condensed, Jost, sans-serif"
    fontSize: "1.375rem"
    fontWeight: 600
    letterSpacing: "0.12em"
  title:
    fontFamily: "Barlow Condensed, Jost, sans-serif"
    fontSize: "1.125rem"
    fontWeight: 500
    letterSpacing: "0.12em"
  body:
    fontFamily: "Jost, system-ui, sans-serif"
    fontSize: "0.9375rem"
    fontWeight: 400
    lineHeight: 1.5
    fontFeature: "tnum"
  label:
    fontFamily: "Barlow Condensed, Jost, sans-serif"
    fontSize: "0.8125rem"
    fontWeight: 500
    letterSpacing: "0.12em"
  button:
    fontFamily: "Barlow Condensed, Jost, sans-serif"
    fontSize: "0.9375rem"
    fontWeight: 500
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
components:
  button:
    backgroundColor: "transparent"
    textColor: "{colors.ink}"
    typography: "{typography.button}"
    rounded: "{rounded.base}"
    padding: "0.25rem 0.75rem"
  button-primary:
    backgroundColor: "{colors.gold}"
    textColor: "{colors.on-gold}"
    typography: "{typography.button}"
    rounded: "{rounded.base}"
    padding: "0.25rem 0.75rem"
  input:
    backgroundColor: "{colors.glass-strong}"
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    rounded: "{rounded.base}"
    padding: "0.5rem 0.75rem"
  table-row:
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    padding: "0.5rem 0.75rem"
  table-row-selected:
    backgroundColor: "{colors.select}"
    textColor: "{colors.ink}"
  column-head:
    textColor: "{colors.muted}"
    typography: "{typography.label}"
    padding: "0.5rem 0.75rem"
  column-head-sorted:
    textColor: "{colors.gold}"
  details-pane:
    backgroundColor: "{colors.panel}"
    textColor: "{colors.ink}"
    padding: "1.5rem"
    width: "clamp(18rem, 26vw, 24rem)"
  hint-bar:
    backgroundColor: "{colors.ground}"
    textColor: "{colors.muted}"
    typography: "{typography.button}"
    padding: "0.5rem 2rem"
  hint-key:
    textColor: "{colors.ink}"
    rounded: "{rounded.base}"
    padding: "0 0.4em"
  tab:
    textColor: "{colors.muted}"
    typography: "{typography.title}"
    padding: "0.5rem 0"
  tab-active:
    textColor: "{colors.ink}"
---

# Design System: Hermaeus

## Overview

**Creative North Star: "The Game's Own Menus"**

Hermaeus is one more menu of the game you just closed. Every screen is a categories column, a sortable item table, and a full-height details pane, set on smoked glass with hairline rules and driven from the keyboard. It deliberately refuses the admin-dashboard default of gray sidebar, cards, and a blue button: there are no cards, no filled surfaces beyond the details pane, and no accent color besides gold.

The world is dense and quiet. Text is Jost at a compact 15px with tabular figures so counts, times, and importance scores line up in columns; anything that labels rather than reads (categories, headings, column heads, keys, and buttons) is set in tracked, uppercase Barlow Condensed. Gold is scarce and means "this is selected" or "this is what you can do". Dark mode is the game's near-black menus; light mode is the same world in pale smoke, built from one token set with the same roles.

The signature is the bottom button-hint strip: every hint is a real keyboard shortcut, and clicking the hint does the same thing. The SVG diamond marks the selected row, as it does in the game.

**Key Characteristics:**
- Smoked-glass ground: a fixed radial gradient from a lighter smoke heart to the ground color.
- Hairline 1px rules (two strengths) instead of boxes, cards, or shadows.
- Tracked condensed caps for labels; Jost for reading text; tabular figures everywhere.
- Gold for selection and action, and for the group headings that sort lists into sections.
- Item table plus a sticky, full-height details pane that sizes itself to end at the hint strip.
- A fixed bottom hint strip of live keyboard shortcuts, right-aligned like the game's button hints.

## Colors

A near-monochrome warm palette, off-white on near-black (or near-black on pale smoke), with one gold accent. Each role has a dark value and a `-light` counterpart; the light set applies under `data-theme="light"` or when the theme is "system" and the OS prefers light.

### Primary
- **Septim Gold** (`gold` / `gold-light`): the selection and action color. The diamond marker, the primary button fill, the sorted column head and its hover, the active tab's underline, focus rings, text selection, the caret, checkbox accents, inline text buttons, the stats bars (at 55% opacity), and the group headings in tables and stat lists (a day in the memory table, a plugin group, a stat section's title). Light mode uses a much darker bronze so it holds contrast on pale smoke.
- **Gold Ink** (`on-gold` / `on-gold-light`): text on a gold fill, used for the primary button and selected text.

### Neutral
- **Menu Black** (`ground` / `ground-light`): the outer ground, the edge of the radial gradient, and the hint strip's opaque background.
- **Smoke Heart** (`smoke` / `smoke-light`): the lighter center of the smoked-glass gradient (an ellipse at 65% 25%), and the quick-find dialog's surface.
- **Pane Glass** (`panel` / `panel-light`): the opaque fill of the details pane, so table content doesn't show through while the pane sticks.
- **Smoked Field** (`glass-strong` / `glass-strong-light`): the translucent fill of inputs, selects, and text areas.
- **Selected Row Haze** (`select` / `select-light`): the faint ink wash on the selected table row, the selected quick-find result, and the checked theme option.
- **Parchment Ink** (`ink` / `ink-light`): body text, current navigation, key caps.
- **Weathered Ink** (`muted` / `muted-light`): secondary text, column heads, definition terms, unbuilt sections, hints at rest, placeholders.
- **Hairline** (`rule` / `rule-light`): row dividers and structural borders between shell zones.
- **Strong Hairline** (`rule-strong` / `rule-strong-light`): column-head underlines, group-heading underlines, control borders, the pane's left edge.

### Status
- **Amber Warning** (`warn` / `warn-light`): "Skyrim running · read-only" and warning notices or flags.
- **Ember Error** (`danger` / `danger-light`): error notices only.

### Named Rules
**The Scarce Gold Rule.** Gold marks selection, action, and group headings that sort a list into sections. It never fills a surface larger than a button and never decorates. If something is gold, it is selected, actionable, or a heading inside a list.

**The One Token Set Rule.** Light mode is the same roles with different values, never a different design. A new color ships as a dark/light pair or not at all.

## Typography

**Display Font:** Barlow Condensed 500 and 600 (with Jost, sans-serif), self-hosted via @fontsource
**Body Font:** Jost 400 and 500 (with system-ui, sans-serif), self-hosted via @fontsource

**Character:** Barlow Condensed in tracked capitals is the game's menu voice, crisp and narrow; Jost is a calm geometric text face that stays legible at small sizes. The scale is short (five steps) because the interface is a menu, not a document.

### Hierarchy
- **Display** (Barlow Condensed 600, 2.5rem, line-height 1, uppercase): the page title only (Characters, Playthroughs, a character's name).
- **Headline** (Barlow Condensed 600, 1.375rem, uppercase): the selected item's name in the details pane, and the app name in the top bar.
- **Title** (Barlow Condensed 500, 1.125rem, uppercase): category links in the sections column and the tabs under a character.
- **Body** (Jost 400, 0.9375rem, line-height 1.5, tabular figures): table cells, values, prose. Long text in the pane runs at line-height 1.6 with a 65ch measure.
- **Label** (Barlow Condensed 500, 0.8125rem, 0.12em tracking, uppercase): column heads, group headings, definition terms in caps, the playthrough and game-state labels, and the theme switch.
- **Button** (Barlow Condensed 500, 0.9375rem, 0.12em tracking, uppercase): buttons and the hint strip.

### Named Rules
**The Label Versus Read Rule.** Anything that labels is tracked condensed caps; anything that is read is Jost in sentence case. Counts beside a label switch back to Jost with no tracking.

**The Lined-Up Figures Rule.** Tabular figures are on at the body level. Numeric columns align right.

## Layout

The shell is a CSS grid: a full-width top bar, then a 13rem sections column beside the main area, with the hint strip fixed along the bottom. The top bar holds the name at left and the playthrough select, Skyrim's state, and the theme switch at right. Main content has 1.5rem top padding, 2rem side padding and 5rem bottom padding to clear the hint strip.

List screens split into the item table and a details pane of `clamp(18rem, 26vw, 24rem)` (the memory and diary tables use `minmax(20rem, 26rem)`, narrowing to `minmax(18rem, 22rem)` under 70rem). The pane sticks to the top and is sized by script to end exactly at the hint strip, so a form's Save stays in view. A filter bar above the table stops at the pane's edge.

Spacing follows a seven-step scale from 0.25rem to 2.5rem; table cells use 0.5rem by 0.75rem, gaps between zones use 1.5rem, and section gaps on overview grids use 2.5rem.

Responsive behavior: under 70rem, wide columns drop and two-column overview grids stack. Under 60rem, the sections column becomes a horizontal scrolling row under the top bar (unbuilt sections and group labels hide), the pane stacks below the table with a top rule instead of a left one, and less important columns hide.

## Elevation & Depth

The world is flat. Depth comes from the smoked-glass gradient behind everything, from hairline rules, and from the one opaque surface (the details pane) against the translucent ground. Rows, sections, and the shell have no shadows.

### Shadow Vocabulary
- **Modal lift** (`box-shadow: 0 1.5rem 3rem rgb(0 0 0 / 0.45)`, backdrop `rgb(0 0 0 / 0.45)`): only for the quick-find dialog, which floats over the page.

### Named Rules
**The Hairline Rule.** Separate with a 1px rule, never a box or a shadow. Use `rule` between rows and shell zones, `rule-strong` under column heads and group headings and around controls.

## Shapes

Corners are nearly square: one radius (3px) on buttons, inputs, key caps, the theme switch, and the dialog. Tables, the pane, and the shell have square edges. The only drawn geometry is the 10-unit SVG diamond (rendered at 9px, filled gold) that marks the selected row, and an 8px triangle on the sorted column head that flips for ascending order. Marked memories show a 9px gold square in the same marker column.

## Components

### Buttons
Quiet outlined caps that light up gold on hover.
- **Shape:** gently squared (3px radius), 1px `rule-strong` border, padding 0.25rem by 0.75rem.
- **Default:** transparent fill, ink text, button typography. Hover turns the border gold.
- **Primary:** gold fill and border, gold-ink text. Used once per action group (Save, confirm).
- **Disabled:** 50% opacity, not-allowed cursor.
- **Text button:** gold, underlined, no border, for actions inside a sentence.
- **Focus:** a 2px gold outline offset by 2px, across all controls.

### Inputs / Fields
- **Style:** smoked-field fill, 1px `rule-strong` border, 3px radius, 0.5rem by 0.75rem padding, muted placeholder, gold caret.
- **Labels:** the field's label sits above it in muted text at the label size.
- **Checkboxes and radios:** native, with a gold accent color.

### Item Table
The core of every list screen.
- Collapsed borders, 1px `rule` under every row, cells padded 0.5rem by 0.75rem.
- A narrow first column (1.5rem) holds the diamond for the selected row.
- The selected row gets the `select` haze; inactive rows drop to muted text.
- Column heads are sort buttons in label caps, muted, gold when sorted or hovered, with a `rule-strong` underline. Counts and dates sort largest or newest first.
- Group heading rows (a day, a plugin) are gold label caps with generous top padding (1.5rem) and a `rule-strong` underline.
- Memory and diary rows enter with a short fly-in (6px rise, 180ms).

### Details Pane
- Opaque `panel` fill, a `rule-strong` left edge, 1.5rem padding, sticky at the top, sized to end at the hint strip.
- The item's name in headline caps, then a definition grid (7 to 7.5rem term column, muted terms at the label size), then actions.
- Forms inside the pane keep their buttons in a sticky footer with a `rule` top edge.

### Navigation
- **Sections column:** title-size caps in muted ink; the current section and hover turn ink. Unbuilt sections are listed in muted caps under a small "Not built yet" label with a hairline under it; the Configuration group label is muted label caps, never as bright as the current section.
- **Tabs:** title-size caps in muted ink, spaced 2rem apart over a `rule` baseline; the current tab is ink with a 1px gold underline. A tab's count follows in muted Jost.
- **Theme switch:** a 3px-radius outlined group of three caps options (Auto, Dark, Light); the checked one gets the `select` haze and ink text.

### Button-Hint Strip (signature)
Fixed along the bottom of the viewport, opaque ground, `rule` top edge, hints right-aligned with 1.5rem gaps. Each hint is a button: a key cap (1px `rule-strong` border, 3px radius, ink, min 1.6em wide) followed by the action in muted caps; hover turns it ink, unavailable hints fade to 45%. Every hint is backed by a real shortcut registered in the shortcut registry, and the strip shows only what the current screen offers.

### Stat List
A name-and-count list under a gold label-caps heading. Each row carries a 2px gold bar at 55% opacity under the name, scaled to the count; counts are muted and right-aligned.

### Quick Find
A dialog (28rem max, 18vh from the top) on the smoke surface with a strong hairline, 3px radius and the modal lift; results are full-width rows with the diamond on the selected one.

## Do's and Don'ts

### Do:
- **Do** put every list in the item table with a diamond column and a details pane beside it.
- **Do** register a real keyboard shortcut for every primary action and let the hint strip show it.
- **Do** set labels in tracked Barlow Condensed caps and reading text in Jost, with tabular figures.
- **Do** separate with 1px hairlines (`rule`, `rule-strong`) and keep corners at 3px.
- **Do** define every new color as a dark and light pair in the same token set.
- **Do** keep the details pane opaque and sized to end at the hint strip.

### Don't:
- **Don't** use cards, gray sidebars, filled tiles, or a blue (or any second) accent color.
- **Don't** use gold for decoration or large fills; it means selection, action, or a list's group heading.
- **Don't** add shadows to rows, panes, or the shell; the only shadow is the floating quick-find dialog.
- **Don't** show a hint in the strip that isn't backed by a working shortcut.
- **Don't** make a group label as bright as the current section.
