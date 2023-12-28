<script>
	import { post } from "$lib/api.js";
	import { goto } from "$app/navigation";

	import PageTitle from "$lib/components/PageTitle.svelte";
	import Input from "$lib/components/Input.svelte";
	import InputPassword from "$lib/components/InputPassword.svelte";

	let username = "";
	let password = "";
	let passwordConfirm = "";

	let otherError = null;
	let fieldErrors = {};
	let isLoading = false;

	const handleSubmit = async () => {
		isLoading = true;

		let response = await post("/api/register", {
			fetch,
			body: {
				username,
				password,
				password_confirmation: passwordConfirm,
			},
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

		goto("/login?success=" + encodeURIComponent("Registration successful"));
	};
</script>

<PageTitle title="Register" error={otherError} />

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
				<InputPassword
					label="Password Confirmation"
					bind:value={passwordConfirm}
					placeholder="**** again"
					errors={fieldErrors.password_confirmation}
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
						<strong>Register</strong>
					</button>
					<a class="button is-white" href="/">Back to Home</a>
				</p>
			</div>
		</section>
	</fieldset>
</form>
