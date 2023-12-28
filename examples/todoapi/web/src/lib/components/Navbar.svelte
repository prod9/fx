<script>
	import { goto } from "$app/navigation";
	import { delete_ } from "$lib/api.js";
	import { user } from "$lib/stores/user";

	$: hasUser = !!$user?.username;

	const handleLogout = async () => {
		let response = await delete_("/api/logout", { fetch });
		let payload = await response.json();
		if (!response.ok) {
			// todo app's session are not that important that we need to ensure correct logout
			// mishandling this will also lead to bad UX in general, so we'll fire-and-forget
			console.error("logout failed: ", payload);
		}

		user.set(null);
		goto("/login?success=" + encodeURIComponent("Logged out"));
	};
</script>

<nav class="navbar is-dark" role="navigation" aria-label="main navigation">
	<div class="container">
		<div class="navbar-brand">
			<a class="navbar-item" href="/">
				<p>Examples TODO App</p>
			</a>

			<a
				role="button"
				class="navbar-burger"
				aria-label="menu"
				aria-expanded="false"
				data-target="navbarBasicExample"
			>
				<span aria-hidden="true" />
				<span aria-hidden="true" />
				<span aria-hidden="true" />
			</a>
		</div>

		<div class="navbar-menu">
			<div class="navbar-end">
				{#if hasUser}
					<div class="navbar-item">
						<p>{$user.username}</p>
					</div>
					<div class="navbar-item">
						<div class="buttons">
							<button class="button is-small" on:click={handleLogout}>
								Log out
							</button>
						</div>
					</div>
				{/if}
			</div>
		</div>
	</div>
</nav>
