<script>
	import { onMount } from "svelte";
	import { get, patch, delete_ } from "$lib/api.js";
	import { user } from "$lib/stores/user.js";

	import ModalNewTodo from "../../lib/components/ModalNewTodo.svelte";
	import PageTitle from "../../lib/components/PageTitle.svelte";
	import SelectPage from "../../lib/components/SelectPage.svelte";

	export let data;

	let newTodoOpen = false;
	let updatedTodo = null;
	let todos = data?.todos ?? [];

	let page = todos.page;
	$: maxPage = todos.total_pages;

	let error = "";
	let isLoading = false;

	const loadTodos = async () => {
		newTodoOpen = false;
		isLoading = true;

		let response = await get("/api/todos?page=" + page.toString(), { fetch });
		let payload = await response.json();
		isLoading = false;
		if (!response.ok) {
			error = payload?.message;
			return todos;
		} else {
			error = "";
			return payload ?? [];
		}
	};

	onMount(() => {
		user.subscribe(async (_user) => {
			todos = await loadTodos();
		});
	});

	const handleRefresh = async () => {
		updatedTodo = null;
		todos = await loadTodos();
	};
	const handleNewTodo = () => {
		newTodoOpen = true;
	};
	const handleTodoCreated = async (todo) => {
		newTodoOpen = false;
		updatedTodo = todo;

		if (todos.data.length == todos.rows_per_page) {
			// if current page is full, the new todo would show up on a new page
			maxPage = todos.total_pages + 1;
		}

		page = maxPage;
		todos = await loadTodos();
	};

	const handleCompleteTodo = async (todo, completed) => {
		isLoading = true;

		let response = await patch("/api/todos", {
			fetch,
			body: { id: todo.id, completed: completed },
		});

		let payload = await response.json();
		isLoading = false;
		if (!response.ok) {
			error = payload?.message;
			return;
		}

		error = "";
		for (let i = 0; i < todos.data.length; i++) {
			if (todos.data[i].id === payload.id) {
				todos.data[i] = payload;
				updatedTodo = payload;
				return;
			}
		}
	};
	const handleDeleteTodo = async (todo) => {
		isLoading = true;

		let response = await delete_("/api/todos/" + todo.id, { fetch });
		let payload = await response.json();
		isLoading = false;
		if (!response.ok) {
			error = payload?.message;
			return;
		}

		todos = await loadTodos();
	};
</script>

<PageTitle title="Todos" {error} success={data?.success}>
	<div class="field is-grouped">
		<div class="control">
			<SelectPage bind:page bind:maxPage onChange={handleRefresh} />
		</div>
		<div class="control">
			<button disabled={isLoading} on:click={handleRefresh} class="button">
				Refresh
			</button>
		</div>
		<div class="control">
			<button
				disabled={isLoading}
				on:click={handleNewTodo}
				class="button is-success"
			>
				Add Todo
			</button>
		</div>
	</div>
</PageTitle>

<ModalNewTodo onCreated={handleTodoCreated} bind:isOpen={newTodoOpen} />

<section class="section">
	<div class="container">
		<table class="table is-hoverable is-fullwidth is-striped">
			<thead>
				<tr><th>Task</th><th>Completed</th><th>Actions</th></tr>
			</thead>
			<tbody>
				{#each todos.data as todo}
					<tr class:is-selected={todo.id === updatedTodo?.id}>
						<td>
							<p><strong>{todo.title}</strong></p>
							<p>{todo.description}</p>
						</td>
						<td>
							{#if todo.completed}
								<button
									on:click={() => handleCompleteTodo(todo, false)}
									class="button is-small is-success">Completed</button
								>
							{:else}
								<button
									on:click={() => handleCompleteTodo(todo, true)}
									class="button is-small is-light">Not Completed</button
								>
							{/if}
						</td>
						<td>
							<div class="buttons">
								<button
									on:click={() => handleDeleteTodo(todo)}
									class="button is-small is-danger">Delete</button
								>
							</div>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</section>
