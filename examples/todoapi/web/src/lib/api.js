// our analog of sveltekit's error (and json) func
export const json = (payload) => {
	return new Response(JSON.stringify(payload), {
		status: 200,
		headers: { "Content-Type": "application/json" }
	})
}
export const error = (status, payload) => {
	return new Response(JSON.stringify(payload), {
		status: status,
		headers: { "Content-Type": "application/json" }
	})
}

// http methods
export const get = async (url, { fetch, token }) => {
	let opts = { method: "GET", headers: {} }
	if (!!token) {
		opts.headers["Authorization"] = `Bearer ${token}`;
	}

	return await fetch(url, opts)
}

export const post = async (url, { fetch, token, body }) => {
	let opts = { method: 'POST', headers: {} }
	if (!!token) {
		opts.headers["Authorization"] = `Bearer ${token}`;
	}
	if (!!body) {
		opts.headers["Content-Type"] = "application/json";
		opts.body = JSON.stringify(body)
	}

	return await fetch(url, opts)
}

export const patch = async (url, { fetch, token, body }) => {
	let opts = { method: 'PATCH', headers: {} }
	if (!!token) {
		opts.headers["Authorization"] = `Bearer ${token}`;
	}
	if (!!body) {
		opts.headers["Content-Type"] = "application/json";
		opts.body = JSON.stringify(body)
	}

	return await fetch(url, opts)
}

export const delete_ = async (url, { fetch, token }) => {
	let opts = { method: 'DELETE', headers: {} }
	if (!!token) {
		opts.headers["Authorization"] = `Bearer ${token}`;
	}

	return await fetch(url, opts)
};
