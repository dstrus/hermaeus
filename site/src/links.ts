// Where the site sends people outside itself.
export const repository = 'https://github.com/dstrus/hermaeus';
export const releases = `${repository}/releases/latest`;
export const issues = `${repository}/issues`;
export const license = `${repository}/blob/main/LICENSE`;

// One download per OS, attached to each GitHub release under the same names
// (see `make release`), so these always fetch the latest release's.
export const downloads = [
	{ os: 'windows', label: 'Windows', note: '64-bit', href: `${releases}/download/Hermaeus-windows-x64.zip` },
	{ os: 'macos', label: 'macOS', note: 'Apple silicon and Intel', href: `${releases}/download/Hermaeus-macos.zip` },
	{ os: 'linux', label: 'Linux', note: '64-bit', href: `${releases}/download/Hermaeus-linux-x64.tar.gz` }
] as const;
