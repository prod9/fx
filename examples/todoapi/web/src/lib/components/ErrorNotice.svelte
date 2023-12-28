<script>
	export let message = "";
	$: visible = !!message;
</script>

{#if visible}
	<article class="message is-danger">
		<div class="message-body">
			{message}
		</div>
	</article>
{/if}
