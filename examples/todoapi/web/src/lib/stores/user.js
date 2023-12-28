import { writable } from 'svelte/store';

export const user = writable({
	id: 0,
	username: "",
})
