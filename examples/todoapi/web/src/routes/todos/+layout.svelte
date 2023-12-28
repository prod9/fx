<script>
	import { user } from "$lib/stores/user.js";

	export let data;

	$: isLoggedIn = !!$user?.username;
	$: hasSession = !!data?.session;
</script>

{#if isLoggedIn || hasSession}
	<slot />
{:else}
	<!-- not logged in -->
	<section class="hero is-danger is-halfheight">
		<div class="hero-body">
			<div class="container">
				<p class="title">You are not Logged In</p>
				<p class="subtitle">Please login or register to access this page.</p>
				<a href="/register" class="button is-dark is-large"> Register </a>
				<a href="/login" class="button is-dark is-large"> Login </a>
			</div>
		</div>
	</section>
{/if}
