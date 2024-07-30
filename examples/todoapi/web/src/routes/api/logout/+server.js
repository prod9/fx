import { delete_, error, json } from '$lib/api.js';
import { backendApiPrefix } from '$lib/server.js';

export async function DELETE({ cookies, request, fetch }) {
	let token = cookies.get('session');
	let response = await delete_(backendApiPrefix() + '/sessions/current', { fetch, token })

	let session = await response.json();
	if (!response.ok) {
		return error(response.status, session)
	}

	cookies.set('session', '', {
		path: '/',
		expires: new Date(0),
		maxAge: 0,
	})
	return json(session)
}
