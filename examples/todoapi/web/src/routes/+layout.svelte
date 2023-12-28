<script>
	import { onMount } from "svelte";

	import "../app.css";
	import { get } from "$lib/api.js";
	import { user } from "$lib/stores/user.js";
	import Navbar from "$lib/components/Navbar.svelte";

	// load user data, if we have session token from previous logins
	onMount(() => {
		if (!!$user?.username) {
			// user already loaded, or we don't have session token
			return;
		}

		// no async in onMount :/
		// if we're unable to load the user, it's most likely a session problem
		// so we redirect user to re-login
		get("/api/me", { fetch }).then(async (response) => {
			let payload = await response.json();
			if (!response.ok) {
				goto("/login?error=" + encodeURIComponent(payload?.message));
			} else {
				user.set(payload);
			}
		});
	});
</script>

<Navbar />
<slot />
