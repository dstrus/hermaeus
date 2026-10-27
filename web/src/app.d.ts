// See https://svelte.dev/docs/kit/types#app.d.ts
declare global {
	namespace App {
		// The message shown when a page fails to load (see +error.svelte).
		interface Error {
			message: string;
		}
	}
}

export {};
