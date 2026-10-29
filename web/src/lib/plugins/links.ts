// Links between a plugin and what it brings: the Plugins page names a
// plugin by its layer's id, and the Prompts, Actions, and Triggers pages
// narrow to one plugin with ?from=<that id>.

export function pluginHref(save: string, id: string): string {
	return `/${encodeURIComponent(save)}/plugins?item=${encodeURIComponent(id)}`;
}

// fromHref is a content page narrowed to what one plugin brings.
export function fromHref(save: string, page: 'prompts' | 'actions' | 'triggers' | 'bios', id: string): string {
	return `/${encodeURIComponent(save)}/${page}?from=${encodeURIComponent(id)}`;
}
