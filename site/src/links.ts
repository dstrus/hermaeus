// Where the site sends people outside itself.
export const repository = 'https://github.com/dstrus/hermaeus';
export const releases = `${repository}/releases/latest`;
export const issues = `${repository}/issues`;
export const license = `${repository}/blob/main/LICENSE`;

// One program per OS, attached to each GitHub release. Until the asset
// names are settled, every download opens the latest release's page; until the
// first release is published, GitHub sends that to the empty Releases list.
export const downloads = [
	{ os: 'windows', label: 'Windows', note: '64-bit', href: releases },
	{ os: 'macos', label: 'macOS', note: '', href: releases },
	{ os: 'linux', label: 'Linux', note: '', href: releases }
] as const;
