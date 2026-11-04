# Changelog

What changed in each release of Hermaeus, newest first.

## 0.1.0-beta.3 (2026-11-04)

### Added

- "Config", at the top of the sidebar. Lets you point Hermaeus at a different SkyrimNet folder, and remembers it, even when it auto-detected one. "Change folder" on the Playthroughs page does the same.

### Changed

- "Playthrough" in the top bar links to the list of playthroughs, which was previously reachable only through the Hermaeus logo/wordmark in the upper-left.

### Fixed

- Quick Find closed when a text selection in its search field ended with the mouse button/trackpad being released with the cursor outside the modal.

## 0.1.0-beta.2 (2026-11-03)

### Fixed

- A new entry on the Knowledge page was saved as an ordinary memory (type `EXPERIENCE`) instead of as world knowledge. New entries now start as `KNOWLEDGE`, with an importance of 0.7. Entries already added this way keep their type. To fix one, edit it and choose `KNOWLEDGE`.

### Changed

- The Characters page explains which characters each view lists. "This playthrough" shows everyone SkyrimNet has stored something for, and "Every bio" shows only characters who already have a bio, since Hermaeus can't create an initial bio.
- "Create a save bio from this" now has its own line, just above the bio it copies, so it's more obvious why you can't edit the built-in bios.

## 0.1.0-beta.1 (2026-11-01)

The first public build: an unofficial dashboard for SkyrimNet that reads and edits its data and settings with Skyrim closed, tested against SkyrimNet Beta 29.
