// Links between a plugin and what it brings: the Plugins page names a
// plugin by its layer's id, and the Prompts, Actions, Triggers, and bios
// pages narrow to one plugin with ?from=<that id>.
import { m } from '#lib/copy/messages.ts';
import { href } from '#lib/nav.ts';

export function pluginHref(save: string, id: string): string {
	return href(save, 'plugins', `?item=${encodeURIComponent(id)}`);
}

// fromHref is a content page narrowed to what one plugin brings.
export function fromHref(save: string, page: 'prompts' | 'actions' | 'triggers' | 'bios', id: string): string {
	const from = `?from=${encodeURIComponent(id)}`;
	return page === 'bios' ? `/${encodeURIComponent(save)}/bios${from}` : href(save, page, from);
}

// contentHref is a record on the Content page, by its path.
export function contentHref(save: string, path: string): string {
	return href(save, 'content', `?item=${encodeURIComponent(path)}`);
}

const recordKinds = ['triggers/', 'actions/', 'filters/', 'voice_effects/', 'entities/', 'dialogue_actions/', 'translator/', 'items/', 'spells/', 'furniture/', 'identity/'];

// editHref is the page that edits a content file, by its path under a
// layer: a bio, a prompt, or a record. Other files have none.
export function editHref(save: string, path: string): { href: string; page: string } | undefined {
	const s = `/${encodeURIComponent(save)}`;
	const bio = /^prompts\/characters\/([^/]+)\.prompt$/i.exec(path);
	if (bio) return { href: `${s}/bios/${encodeURIComponent(bio[1]!)}`, page: m.plugins.editOn.bio };
	if (/^prompts\/.+\.prompt$/i.test(path))
		return {
			href: href(save, 'prompts', '/' + path.slice('prompts/'.length).split('/').map(encodeURIComponent).join('/')),
			page: m.plugins.editOn.prompt
		};
	if (recordKinds.some((k) => path.toLowerCase().startsWith(k))) return { href: contentHref(save, path), page: m.plugins.editOn.content };
	return undefined;
}
