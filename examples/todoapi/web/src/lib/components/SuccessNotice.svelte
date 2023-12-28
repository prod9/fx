<script>
	export let message = "";
	$: visible = !!message;
</script>

{#if visible}
	<article class="message is-success">
		<div class="message-body">
			{message}
		</div>
	</article>
{/if}
