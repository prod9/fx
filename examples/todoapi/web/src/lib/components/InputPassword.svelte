<script>
	export let label = "";
	export let value = "";
	export let errors = null;
	export let placeholder = "";

	$: hasError = !!errors;
</script>

<div class="field">
	<label class="label">{label}</label>
	<div class="control">
		<input
			class="input is-large"
			class:is-danger={hasError}
			bind:value
			type="password"
			{placeholder}
		/>
	</div>
	{#if hasError}
		{#each errors as error}
			<p class="help is-danger">{error.field} {error.message}</p>
		{/each}
	{/if}
</div>
