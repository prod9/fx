<script>
	export let message = "";
	$: visible = !!message;

	const handleClose = () => {
		visible = false;
	};
</script>

{#if visible}
	<article class="message is-success is-small">
		<div class="message-header">
			<p>Success</p>
			<button class="delete" on:click={handleClose} aria-label="delete" />
		</div>
		<div class="message-body">
			{message}
		</div>
	</article>
{/if}
