import { get, json, error } from '$lib/api.js';

// client-side
export async function load({ fetch }) {
	let response = await get("/api/todos", { fetch });
	let payload = await response.json();
	if (!response.ok) {
		console.error(payload?.message)
		return { todos: [] }
	} else {
		return { todos: payload }
	}
}
