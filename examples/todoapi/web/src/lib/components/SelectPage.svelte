<script>
	export let page = 1;
	export let maxPage = 1;
	export let onChange = () => {};

	$: pages = Array.from({ length: maxPage }, (_, i) => i + 1);
</script>

<div class="select">
	<select bind:value={page} on:change={() => onChange(page)}>
		{#each pages as p}
			<option value={p}>{p}</option>
		{/each}
	</select>
</div>
