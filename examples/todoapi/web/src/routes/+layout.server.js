export function load({ cookies, url }) {
	let token = cookies.get('session');

	return {
		session: token,
		success: url.searchParams.get('success') || null,
	}
}
