import { get, error, json } from '$lib/api.js';
import { backendApiPrefix } from '$lib/server.js';

export async function GET({ cookies, fetch }) {
	let token = cookies.get('session');;
	let response = await get(backendApiPrefix() + '/users/current', { fetch, token })

	let payload = await response.json();
	if (!response.ok) {
		return error(response.status, payload)
	} else {
		return json(payload)
	}
}
