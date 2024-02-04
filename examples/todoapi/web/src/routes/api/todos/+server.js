import { get, post, patch, error, json } from '$lib/api.js';
import { backendApiPrefix } from '$lib/server.js';

export async function GET({ url, cookies, fetch }) {
	let page = parseInt(url.searchParams.get('page'), 10);
	if (isNaN(page) || page < 1) {
		page = 1
	}

	let token = cookies.get('session');
	let path = backendApiPrefix() + '/todos?per_page=5&page=' + page;
	let response = await get(path, { fetch, token })

	let payload = await response.json();
	if (!response.ok) {
		return error(response.status, payload)
	} else {
		return json(payload)
	}
}

export async function POST({ cookies, fetch, request }) {
	let token = cookies.get('session');
	let response = await post(backendApiPrefix() + '/todos', {
		fetch,
		token,
		body: await request.json()
	})

	let payload = await response.json();
	if (!response.ok) {
		return error(response.status, payload)
	} else {
		return json(payload)
	}
}

export async function PATCH({ cookies, fetch, request }) {
	let token = cookies.get('session');
	let todo = await request.json();
	let response = await patch(backendApiPrefix() + '/todos/' + todo.id, {
		fetch,
		token,
		body: { completed: todo.completed }
	})

	let payload = await response.json();
	if (!response.ok) {
		return error(response.status, payload)
	} else {
		return json(payload)
	}
}
