// The one place the page talks to Hermaeus's API. Every request carries the
// session's token, and every failure becomes an ApiError holding the
// message the server wrote for the user.
import { m } from '#lib/copy/messages.ts';
import type { Error as ErrorBody } from './types/httpapi.ts';
import { TokenHeader } from './types/api.ts';

type Fetch = typeof fetch;

export class ApiError extends Error {
	constructor(
		message: string,
		readonly status: number
	) {
		super(message);
	}
}

function token(): string {
	return document.querySelector<HTMLMetaElement>('meta[name="hermaeus-token"]')?.content ?? '';
}

// request sends a request to the API. Pass the fetch a load function
// receives when there is one, so SvelteKit can track the request.
export async function request<T>(
	method: 'GET' | 'PUT' | 'POST' | 'PATCH' | 'DELETE',
	path: string,
	body?: unknown,
	fetchFn: Fetch = fetch
): Promise<T> {
	const headers: Record<string, string> = { [TokenHeader]: token() };
	if (body !== undefined) headers['Content-Type'] = 'application/json';
	let res: Response;
	try {
		res = await fetchFn(path, {
			method,
			headers,
			body: body === undefined ? undefined : JSON.stringify(body)
		});
	} catch {
		throw new ApiError(m.shell.notResponding, 0);
	}
	if (!res.ok) {
		const reply = (await res.json().catch(() => null)) as ErrorBody | null;
		throw new ApiError(reply?.error ?? m.shell.requestFailed({ method, path, status: res.status }), res.status);
	}
	return (await res.json()) as T;
}

// playthroughPath is the API path for something in one playthrough.
export function playthroughPath(save: string, ...parts: string[]): string {
	return '/api/playthroughs/' + [save, ...parts].map(encodeURIComponent).join('/');
}

// Media (images, audio) needs the token too, which an <img> or <audio> can't
// send, so it's fetched here and shown from a blob URL. The most recent are
// kept, so going back to one doesn't fetch it again.
const kept = new Map<string, Promise<string>>();
const keep = 80;

// mediaURL is a blob URL for the API path's file. A failed fetch isn't kept,
// so it can be tried again.
export function mediaURL(path: string): Promise<string> {
	const had = kept.get(path);
	if (had) {
		kept.delete(path); // to the newest end
		kept.set(path, had);
		return had;
	}
	const p = fetch(path, { headers: { [TokenHeader]: token() } }).then(async (res) => {
		if (!res.ok) {
			const reply = (await res.json().catch(() => null)) as ErrorBody | null;
			throw new ApiError(reply?.error ?? `GET ${path} failed (${res.status})`, res.status);
		}
		return URL.createObjectURL(await res.blob());
	});
	p.catch(() => kept.delete(path));
	kept.set(path, p);
	while (kept.size > keep) {
		const [oldest, url] = kept.entries().next().value!;
		kept.delete(oldest);
		url.then(URL.revokeObjectURL, () => {});
	}
	return p;
}
