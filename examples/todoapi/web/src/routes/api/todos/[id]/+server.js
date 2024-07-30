import { delete_, error, json } from '$lib/api.js';
import { backendApiPrefix } from '$lib/server.js';

export async function DELETE({ cookies, fetch, params }) {
	let token = cookies.get('session');
	let response = await delete_(backendApiPrefix() + '/todos/' + params.id, {
		fetch,
		token
	})

	let payload = await response.json();
	if (!response.ok) {
		return error(response.status, payload)
	} else {
		return json(payload)
	}
}
