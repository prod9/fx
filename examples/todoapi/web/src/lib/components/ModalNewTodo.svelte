<script>
	import { post } from "$lib/api.js";

	import Input from "$lib/components/Input.svelte";
	import TextArea from "$lib/components/TextArea.svelte";

	export let isOpen = false;
	export let onCreated = (_todo) => {};

	let title = "";
	let description = "";

	let otherError = null;
	let fieldErrors = {};
	let isLoading = false;

	const handleClose = () => {
		isOpen = false;
		isLoading = false;
	};

	const handleSubmit = async () => {
		isLoading = true;

		let response = await post("/api/todos", {
			fetch,
			body: { title, description },
		});

		let payload = await response.json();
		isLoading = false;
		if (!response.ok) {
			if (!!payload?.data) {
				otherError = null;
				fieldErrors = payload?.data;
			} else {
				otherError = payload?.message;
				fieldErrors = {};
			}
		} else {
			onCreated?.(payload);
		}
	};
</script>

<div class="modal" class:is-active={isOpen}>
	<div class="modal-background" on:click|preventDefault={handleClose} />
	<div class="modal-card">
		<header class="modal-card-head">
			<p class="modal-card-title">New Todo</p>
			<button
				on:click|preventDefault={handleClose}
				class="delete"
				aria-label="close"
			/>
		</header>

		<form on:submit|preventDefault={handleSubmit}>
			<section class="modal-card-body">
				<fieldset disabled={isLoading}>
					<Input
						label="Title"
						bind:value={title}
						placeholder="Title"
						errors={fieldErrors.title}
					/>
					<TextArea
						label="Description"
						bind:value={description}
						placeholder="Description"
						errors={fieldErrors.description}
					/>
				</fieldset>
			</section>

			<footer class="modal-card-foot">
				<button
					disabled={isLoading}
					type="submit"
					class="is-large button is-success">Add Todo</button
				>
			</footer>
		</form>
	</div>
</div>
