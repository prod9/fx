<script>
	import { goto } from "$app/navigation";
	import { get, post } from "$lib/api.js";
	import { user } from "$lib/stores/user.js";

	import PageTitle from "../../lib/components/PageTitle.svelte";
	import Input from "$lib/components/Input.svelte";
	import InputPassword from "$lib/components/InputPassword.svelte";

	export let data;

	let username = "";
	let password = "";

	let otherError = null;
	let fieldErrors = {};
	let isLoading = false;

	const handleSubmit = async () => {
		isLoading = true;

		let response = await post("/api/login", {
			fetch,
			body: { username, password },
		});

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

		// success, load user data
		response = await get("/api/me", { fetch });
		payload = await response.json();
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
		goto("/todos?success=" + encodeURIComponent("Logged in"));
	};
</script>

<PageTitle title="Login" error={otherError} success={data?.success} />

<form class="form" disabled on:submit|preventDefault={handleSubmit}>
	<fieldset disabled={isLoading}>
		<section class="section">
			<div class="container">
				<Input
					label="Username"
					bind:value={username}
					placeholder="Username"
					errors={fieldErrors.username}
				/>
				<InputPassword
					label="Password"
					bind:value={password}
					placeholder="****"
					errors={fieldErrors.password}
				/>
			</div>
		</section>

		<section class="section">
			<div class="container">
				<p class="buttons">
					<button
						disabled={isLoading}
						type="submit"
						class="button is-large is-primary"
					>
						<strong>Login</strong>
					</button>
					<a class="button is-white" href="/">Back to Home</a>
				</p>
			</div>
		</section>
	</fieldset>
</form>
