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
			<p class="title">{title}</p>
			{#if !!success}
				<SuccessNotice message={success} />
			{/if}
			{#if !!error}
				<ErrorNotice message={error} />
			{/if}
		</div>
	</div>
</section>
