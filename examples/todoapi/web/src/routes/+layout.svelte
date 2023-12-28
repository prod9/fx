<script>
	import { onMount } from "svelte";

	import "../app.css";
	import { get } from "$lib/api.js";
	import { user } from "$lib/stores/user.js";
	import Navbar from "$lib/components/Navbar.svelte";

	export let data;

	// load user data, if we have session token from previous logins
	onMount(() => {
		if (!!$user?.username || !data?.session) {
			// user already loaded, or we don't have session token
			return;
		}

		// no async in onMount :/
		get("/api/me", { fetch }).then(async (response) => {
			let payload = await response.json();
			if (!response.ok) {
				isLoading = false;
				if (!!payload?.data) {
					otherError = null;
					fieldErrors = payload?.data;
				} else {
					otherError = payload?.message;
					fieldErrors = {};
				}
				return;
			}

			user.set(payload);
		});
	});
</script>

<Navbar />
<slot />
