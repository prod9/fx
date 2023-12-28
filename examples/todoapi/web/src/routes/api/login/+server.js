import { post, error, json } from '$lib/api.js';
import { backendApiPrefix } from '$lib/server.js';

export async function POST({ cookies, request, fetch }) {
	let login = await request.json();
	let response = await post(backendApiPrefix() + '/sessions', { fetch, body: login })

	let session = await response.json();
	if (!response.ok) {
		return error(response.status, session)
	}

	cookies.set('session', session.token, {
		path: '/',
		maxAge: 7 * 24 * 60 * 60, // 7 days, see DefaultSessionAge on the backend
	})
	return json(session)
}
