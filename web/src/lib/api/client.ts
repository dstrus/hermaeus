// The one place the page talks to Hermaeus's API. Every request carries the
// session's token, and every failure becomes an ApiError holding the
// message the server wrote for the user.
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
	method: 'GET' | 'PUT' | 'POST' | 'DELETE',
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
		throw new ApiError('Hermaeus isn’t responding. Check that it’s still running, then reload the page.', 0);
	}
	if (!res.ok) {
		const reply = (await res.json().catch(() => null)) as ErrorBody | null;
		throw new ApiError(reply?.error ?? `${method} ${path} failed (${res.status})`, res.status);
	}
	return (await res.json()) as T;
}

// playthroughPath is the API path for something in one playthrough.
export function playthroughPath(save: string, ...parts: string[]): string {
	return '/api/playthroughs/' + [save, ...parts].map(encodeURIComponent).join('/');
}
