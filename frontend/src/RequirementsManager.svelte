<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { getSSEClient, type SSEMessage } from './sse';

	interface GradeRequirement {
		grade: number;
		min_total: number;
	}

	interface RequirementGroup {
		id: number;
		label: string;
		min_count: number;
		categories: string[];
	}

	interface Props {
		grades: number[];
		categories: string[];
		onUpdate: () => void;
	}

	const { grades, categories, onUpdate }: Props = $props();

	let loading = $state(false);
	let error = $state('');
	let successMessage = $state('');

	let gradeRequirements = $state<GradeRequirement[]>([]);
	let selectedGrade = $state<number | null>(null);
	let selectedGradeData = $state<any>(null);

	let gradeReqForm = $state({
		grade: '',
		min_total: ''
	});

	let groupForm = $state({
		id: '',
		label: '',
		min_count: '',
		categories: [] as string[]
	});
	let editingGroup = $state<RequirementGroup | null>(null);
	let showGroupForm = $state(false);

	$effect(() => {
		loadAllGradeRequirements();
	});

	function handleRequirementsInvalidation(message: SSEMessage) {
		console.log('Requirements invalidated, refreshing data...');
		loadAllGradeRequirements();
		if (selectedGrade !== null) {
			loadGradeRequirements(selectedGrade);
		}
	}

	onMount(() => {
		const client = getSSEClient();
		if (client) {
			client.on('invalidate_requirements', handleRequirementsInvalidation);
		}
	});

	onDestroy(() => {
		const client = getSSEClient();
		if (client) {
			client.off('invalidate_requirements', handleRequirementsInvalidation);
		}
	});

	async function loadAllGradeRequirements() {
		loading = true;
		error = '';

		try {
			const response = await fetch('/api/admin/requirements');
			if (response.ok) {
				gradeRequirements = await response.json() || [];
			}
		} catch (err: any) {
			error = err.message || 'Failed to load grade requirements';
		} finally {
			loading = false;
		}
	}

	async function loadGradeRequirements(grade: number) {
		loading = true;
		error = '';

		try {
			const response = await fetch(`/api/admin/requirements?grade=${grade}`);
			if (response.ok) {
				selectedGradeData = await response.json();
			}
		} catch (err: any) {
			error = err.message || 'Failed to load grade requirements';
		} finally {
			loading = false;
		}
	}

	function selectGrade(grade: number) {
		selectedGrade = grade;
		loadGradeRequirements(grade);
		gradeReqForm.grade = grade.toString();
		gradeReqForm.min_total = '';
	}

	async function saveGradeRequirement() {
		if (!gradeReqForm.grade || !gradeReqForm.min_total) {
			error = 'Grade and minimum total are required';
			return;
		}

		loading = true;
		error = '';
		successMessage = '';

		try {
			const formData = new FormData();
			formData.append('grade', gradeReqForm.grade);
			formData.append('min_total', gradeReqForm.min_total);

			const response = await fetch('/api/admin/requirements', {
				method: 'POST',
				body: formData
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText);
			}

			successMessage = 'Grade requirement saved successfully';
			loadAllGradeRequirements();
			if (selectedGrade) {
				loadGradeRequirements(selectedGrade);
			}
			onUpdate();
		} catch (err: any) {
			error = err.message || 'Failed to save grade requirement';
		} finally {
			loading = false;
		}
	}

	async function deleteGradeRequirement(grade: number) {
		if (!confirm(`Are you sure you want to delete the requirement for grade ${grade}?`)) {
			return;
		}

		loading = true;
		error = '';

		try {
			const response = await fetch(`/api/admin/requirements?grade=${grade}`, {
				method: 'DELETE'
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText);
			}

			successMessage = 'Grade requirement deleted successfully';
			loadAllGradeRequirements();
			if (selectedGrade === grade) {
				loadGradeRequirements(grade);
			}
			onUpdate();
		} catch (err: any) {
			error = err.message || 'Failed to delete grade requirement';
		} finally {
			loading = false;
		}
	}

	function openGroupForm(group?: RequirementGroup) {
		if (group) {
			editingGroup = group;
			groupForm = {
				id: group.id.toString(),
				label: group.label,
				min_count: group.min_count.toString(),
				categories: [...group.categories]
			};
		} else {
			editingGroup = null;
			groupForm = {
				id: '',
				label: '',
				min_count: '',
				categories: []
			};
		}
		showGroupForm = true;
		error = '';
		successMessage = '';
	}

	function closeGroupForm() {
		showGroupForm = false;
		editingGroup = null;
		groupForm = {
			id: '',
			label: '',
			min_count: '',
			categories: []
		};
	}

	function toggleCategory(category: string) {
		if (groupForm.categories.includes(category)) {
			groupForm.categories = groupForm.categories.filter(c => c !== category);
		} else {
			groupForm.categories = [...groupForm.categories, category];
		}
	}

	async function saveGroup() {
		if (!selectedGrade || !groupForm.label || !groupForm.min_count) {
			error = 'Label and minimum count are required';
			return;
		}

		loading = true;
		error = '';
		successMessage = '';

		try {
			const formData = new FormData();
			if (editingGroup) {
				formData.append('id', groupForm.id);
			}
			formData.append('grade', selectedGrade.toString());
			formData.append('label', groupForm.label);
			formData.append('min_count', groupForm.min_count);
			formData.append('categories', groupForm.categories.join(','));

			const method = editingGroup ? 'PUT' : 'POST';
			const response = await fetch('/api/admin/requirement-groups', {
				method,
				body: formData
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText);
			}

			successMessage = `Requirement group ${editingGroup ? 'updated' : 'created'} successfully`;
			closeGroupForm();
			loadGradeRequirements(selectedGrade);
			onUpdate();
		} catch (err: any) {
			error = err.message || 'Failed to save requirement group';
		} finally {
			loading = false;
		}
	}

	async function deleteGroup(group: RequirementGroup) {
		if (!confirm(`Are you sure you want to delete the requirement group "${group.label}"?`)) {
			return;
		}

		loading = true;
		error = '';

		try {
			const response = await fetch(`/api/admin/requirement-groups?id=${group.id}`, {
				method: 'DELETE'
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText);
			}

			successMessage = 'Requirement group deleted successfully';
			if (selectedGrade) {
				loadGradeRequirements(selectedGrade);
			}
			onUpdate();
		} catch (err: any) {
			error = err.message || 'Failed to delete requirement group';
		} finally {
			loading = false;
		}
	}
</script>

<div class="requirements-manager">
	{#if error}
		<div class="alert alert-error">
			{error}
		</div>
	{/if}

	{#if successMessage}
		<div class="alert alert-success">
			{successMessage}
		</div>
	{/if}

	<!-- Grade Selection and Management -->
		<div class="requirements-grade-section">
			<div class="requirements-grade-buttons">
				{#each grades as grade}
					<button
						onclick={() => selectGrade(grade)}
						class="requirements-grade-button {selectedGrade === grade ? 'requirements-grade-button--active' : 'requirements-grade-button--inactive'}"
						disabled={loading}
					>
						Grade {grade}
					</button>
				{/each}
			</div>

			{#if selectedGrade !== null}
				<div class="requirements-selected-grade-section">
					<!-- Minimum Total Courses -->
					<div class="card">
						<div class="card-header">
							<h4 class="requirements-section-title">
								Minimum Total Courses for Grade {selectedGrade}
							</h4>
						</div>
						<div class="card-body">
							<form onsubmit={(e) => { e.preventDefault(); saveGradeRequirement(); }} class="requirements-minimum-form">
								<div class="form-group">
									<label class="form-label">
										Minimum Total Courses
									</label>
									<input
										bind:value={gradeReqForm.min_total}
										type="number"
										min="0"
										placeholder={selectedGradeData?.min_total?.toString() || '0'}
										disabled={loading}
										class="requirements-minimum-input"
									/>
								</div>
								<button
									type="submit"
									class="btn btn-primary btn-sm"
									disabled={loading}
								>
									{loading ? 'Saving...' : 'Save'}
								</button>
							</form>
							
							{#if selectedGradeData?.min_total > 0}
								<p class="requirements-current-status">
									Current requirement: {selectedGradeData.min_total} courses minimum
								</p>
							{/if}
						</div>
					</div>

					<!-- Category Groups -->
					<div class="card">
						<div class="card-header">
							<h4 class="requirements-section-title">
								Category Requirements for Grade {selectedGrade}
							</h4>
						</div>
						<div class="card-body">
							<!-- Add Group Section -->
							<details class="expander">
								<summary class="expander-header">
									Add New Requirement Group
								</summary>
								<div class="expander-content">
									<button
										onclick={() => openGroupForm()}
										class="btn btn-primary btn-sm"
										disabled={loading}
									>
										Add Group
									</button>
								</div>
							</details>

							{#if selectedGradeData?.groups?.length === 0}
								<p class="requirements-empty-message">
									No category requirements defined. Use "Add New Requirement Group" above to create one.
								</p>
							{:else}
								<div class="requirements-groups-list">
									{#each (selectedGradeData?.groups || []) as group}
										<div class="requirements-group-item">
											<div class="requirements-group-content">
												<div class="requirements-group-header">
													<h5 class="requirements-group-title">{group.label}</h5>
													<div class="requirements-group-actions">
														<button
															onclick={() => openGroupForm(group)}
															class="requirements-group-edit-btn"
															disabled={loading}
														>
															Edit
														</button>
														<button
															onclick={() => deleteGroup(group)}
															class="requirements-group-delete-btn"
															disabled={loading}
														>
															Delete
														</button>
													</div>
												</div>
												<p class="requirements-group-info">
													Minimum: {group.min_count} courses
												</p>
												<p class="requirements-group-categories">
													Categories: {group.categories.join(', ') || 'None'}
												</p>
											</div>
										</div>
									{/each}
								</div>
							{/if}

							<!-- Group Form (Expandable Section) -->
							{#if showGroupForm}
								<div class="requirements-group-form-container">
									<div class="card-header">
										<div class="requirements-group-form-header">
											<h4 class="card-title">
												{editingGroup ? 'Edit' : 'Add'} Requirement Group
											</h4>
											<button
												onclick={closeGroupForm}
												class="requirements-group-form-close-btn"
											>
												Cancel
											</button>
										</div>
									</div>
									<div class="card-body">
										<form onsubmit={(e) => { e.preventDefault(); saveGroup(); }} class="requirements-group-form">
											<div class="form-group">
												<label class="form-label">
													Group Label *
												</label>
												<input
													bind:value={groupForm.label}
													type="text"
													placeholder="e.g., Sport & Enrichment"
													required
													disabled={loading}
													class="input"
												/>
											</div>

											<div class="form-group">
												<label class="form-label">
													Minimum Count *
												</label>
												<input
													bind:value={groupForm.min_count}
													type="number"
													min="0"
													placeholder="1"
													required
													disabled={loading}
													class="input"
												/>
											</div>

											<div class="form-group">
												<label class="form-label-with-description">
													Categories
												</label>
												<div class="requirements-categories-list">
													{#each categories as category}
														<label class="requirements-category-checkbox">
															<input
																type="checkbox"
																checked={groupForm.categories.includes(category)}
																onchange={() => toggleCategory(category)}
																disabled={loading}
																class="checkbox"
															/>
															<span class="requirements-category-label">{category}</span>
														</label>
													{/each}
												</div>
												{#if categories.length === 0}
													<p class="requirements-no-categories-message">No categories available. Create categories first.</p>
												{/if}
											</div>

											<div class="form-actions">
												<button
													type="button"
													onclick={closeGroupForm}
													class="btn btn-secondary"
													disabled={loading}
												>
													Cancel
												</button>
												<button
													type="submit"
													class="btn btn-primary"
													disabled={loading}
												>
													{loading ? 'Saving...' : editingGroup ? 'Update' : 'Create'}
												</button>
											</div>
										</form>
									</div>
								</div>
							{/if}
						</div>
					</div>
				</div>
			{:else}
				<p class="requirements-select-grade-message">
					Select a grade above to manage its requirements.
				</p>
			{/if}
	</div>
</div>
