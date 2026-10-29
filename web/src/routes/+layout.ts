import { ApiError } from '#lib/api/client.ts';
import { getStatus } from '#lib/api/status.ts';
import { getToggles, togglesDependency } from '#lib/settings/api.ts';
import type { PluginSchema, Toggle } from '#lib/api/types/settings.ts';
import { getPluginSchemas } from '#lib/features/plugins.ts';
import type { LayoutLoad } from './$types.ts';

// Hermaeus runs entirely in the browser: the Go server only serves the built
// files and the API, so nothing is rendered or prerendered ahead of time.
export const ssr = false;
export const prerender = false;

export const load: LayoutLoad = async ({ fetch, depends }) => {
	depends(togglesDependency);
	const status = await getStatus(fetch);
	// The quick toggles come from SkyrimNet's settings files; a problem
	// reading those shouldn't keep the rest of Hermaeus from opening.
	let toggles: Toggle[] = [];
	let togglesError = '';
	// Plugins that describe their own settings, for the sidebar.
	let pluginSettings: PluginSchema[] = [];
	if (status.skyrimNetDir) {
		try {
			toggles = (await getToggles(fetch)).toggles;
		} catch (e) {
			togglesError = e instanceof ApiError ? e.message : String(e);
		}
		pluginSettings = await getPluginSchemas(fetch).catch(() => []);
	}
	return { status, toggles, togglesError, pluginSettings };
};
