import { post, error, json } from '$lib/api.js';
import { backendApiPrefix } from '$lib/server.js';

export async function POST({ request, fetch }) {
	let user = await request.json();
	let response = await post(backendApiPrefix() + '/users', { body: user, fetch })

	let payload = await response.json();
	if (!response.ok) {
		return error(response.status, payload)
	} else {
		return json(payload)
	}
}
