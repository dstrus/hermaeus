import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

// During development, `make dev` runs Hermaeus and this dev server in front
// of it. API requests are passed through to Hermaeus's port with the token
// both were started with, since the page Vite serves can't carry it.
const devToken = process.env.HERMAEUS_DEV_TOKEN ?? '';
const devPort = process.env.HERMAEUS_DEV_PORT ?? '8775';

export default defineConfig({
	plugins: [
		sveltekit({
			// A single-page app: Hermaeus serves index.html for every view's
			// address, and the page routes itself.
			adapter: adapter({ pages: 'build', fallback: 'index.html' })
		})
	],
	server: {
		proxy: {
			'/api': {
				target: `http://127.0.0.1:${devPort}`,
				changeOrigin: true,
				headers: { 'X-Hermaeus-Token': devToken }
			}
		}
	}
});
