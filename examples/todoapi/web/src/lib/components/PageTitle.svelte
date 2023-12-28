<script>
	import ErrorNotice from "$lib/components/ErrorNotice.svelte";
	import SuccessNotice from "$lib/components/SuccessNotice.svelte";

	export let title = "";
	export let success = "";
	export let error = "";
</script>

<section class="hero is-light">
	<div class="hero-body">
		<div class="container">
			<div class="level">
				<div class="level-left">
					<p class="title">{title}</p>
				</div>
				<div class="level-right">
					<slot />
				</div>
			</div>

			{#if !!success}
				<SuccessNotice message={success} />
			{/if}
			{#if !!error}
				<ErrorNotice message={error} />
			{/if}
		</div>
	</div>
</section>
