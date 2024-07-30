<script>
	import { goto } from "$app/navigation";
	import { delete_ } from "$lib/api.js";
	import { user } from "$lib/stores/user";

	$: username = $user?.username;
	$: hasUser = !!username;

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

<nav
	class="navbar is-dark is-spaced has-shadow"
	role="navigation"
	aria-label="main navigation"
>
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
			{#if hasUser}
				<div class="navbar-start">
					<a class="navbar-item" href="/todos">To-dos</a>
				</div>
			{:else}
				<div class="navbar-start">
					<a class="navbar-item" href="/register">Register</a>
					<a class="navbar-item" href="/login">Login</a>
				</div>
			{/if}

			{#if hasUser}
				<div class="navbar-end">
					<div class="navbar-item">
						<p>{$user.username}</p>
					</div>
					<div class="navbar-item">
						<div class="buttons">
							<button class="button is-small is-light" on:click={handleLogout}>
								Log out
							</button>
						</div>
					</div>
				</div>
			{/if}
		</div>
	</div>
</nav>
