<script lang="ts">
	interface CatalogItem {
		id: string | number;
		name?: string;
	}

	interface Props {
		title: string;
		items: CatalogItem[];
		apiEndpoint: string;
		itemName: string;
		idType: 'string' | 'number';
		onUpdate: () => void;
	}

	const { title, items, apiEndpoint, itemName, idType, onUpdate }: Props = $props();

	let newItemValue = $state('');
	let error = $state('');
	let loading = $state(false);

	async function addItem() {
		const trimmedValue = idType === 'number' ? String(newItemValue) : String(newItemValue).trim();
		if (!trimmedValue) return;

		loading = true;
		error = '';

		try {
			const formData = new FormData();
			if (idType === 'number') {
				formData.append('grade', trimmedValue);
			} else {
				formData.append('id', trimmedValue);
			}

			const response = await fetch(apiEndpoint, {
				method: 'POST',
				body: formData
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText);
			}

			newItemValue = '';
			onUpdate();
		} catch (err: any) {
			error = err.message || 'Failed to add item';
		} finally {
			loading = false;
		}
	}

	let confirmingDelete = $state<string | number | null>(null);

	async function deleteItem(item: CatalogItem) {
		loading = true;
		error = '';

		try {
			const params = new URLSearchParams();
			if (idType === 'number') {
				params.append('grade', String(item.id));
			} else {
				params.append('id', String(item.id));
			}

			const response = await fetch(`${apiEndpoint}?${params}`, {
				method: 'DELETE'
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText);
			}

			confirmingDelete = null;
			onUpdate();
		} catch (err: any) {
			error = err.message || 'Failed to delete item';
		} finally {
			loading = false;
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			addItem();
		}
	}
</script>

<div class="catalog-manager">
	{#if error}
		<div class="catalog-error-message">
			{error}
		</div>
	{/if}

	<div class="catalog-header">
		<h3 class="catalog-title">{title}</h3>
		<div class="catalog-add-section">
			<input
				bind:value={newItemValue}
				onkeydown={handleKeydown}
				placeholder={`Enter ${itemName}`}
				type={idType === 'number' ? 'number' : 'text'}
				disabled={loading}
				class="catalog-input"
			/>
			<button
				onclick={addItem}
				disabled={loading || !newItemValue}
				class="catalog-add-button"
			>
				{loading ? 'Adding...' : 'Add'}
			</button>
		</div>
	</div>

	<!-- Inline list of items -->
	{#if items.length === 0}
		<div class="catalog-empty-state">
			No {itemName}s defined yet. Add one above to get started.
		</div>
	{:else}
		<div class="catalog-items-list">
			{#each items as item}
				<div class="catalog-item">
					<span class="catalog-item-id">{item.id}</span>
					{#if confirmingDelete === item.id}
						<div class="catalog-item-delete-confirmation">
							<span class="catalog-item-delete-prompt">Delete?</span>
							<button
								onclick={() => deleteItem(item)}
								disabled={loading}
								class="catalog-item-delete-confirm"
								title="Confirm delete"
							>
								✓
							</button>
							<button
								onclick={() => confirmingDelete = null}
								disabled={loading}
								class="catalog-item-delete-cancel"
								title="Cancel"
							>
								✗
							</button>
						</div>
					{:else}
						<button
							onclick={() => confirmingDelete = item.id}
							disabled={loading}
							class="catalog-item-delete-button"
							title="Delete {itemName}"
						>
							×
						</button>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>
