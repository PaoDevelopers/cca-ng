<script lang="ts">
	interface Student {
		id: number;
		name: string;
		grade: number;
		legal_sex: 'F' | 'M' | 'X';
		created_at: string;
		updated_at: string;
	}

	interface Props {
		students: Student[];
		grades: number[];
		onUpdate: () => void;
	}

	const { students, grades, onUpdate }: Props = $props();

	let activeSection = $state('list');
	let loading = $state(false);
	let error = $state('');
	let successMessage = $state('');

	let editingStudent = $state<Student | null>(null);
	let studentForm = $state({
		id: '',
		name: '',
		grade: '',
		legal_sex: 'F' as 'F' | 'M' | 'X',
		password: ''
	});

	let csvFile = $state<File | null>(null);
	let csvPreview = $state<any>(null);
	let csvUploading = $state(false);

	let searchTerm = $state('');
	let filteredStudents = $derived.by(() => {
		if (!searchTerm) return students;
		const term = searchTerm.toLowerCase();
		return students.filter(student => 
			student.name.toLowerCase().includes(term) ||
			student.id.toString().includes(term) ||
			student.grade.toString().includes(term) ||
			student.legal_sex.toLowerCase().includes(term)
		);
	});

	let confirmingDelete = $state<Student | null>(null);
	let confirmingDeleteAll = $state(false);

	function resetForm() {
		studentForm = {
			id: '',
			name: '',
			grade: '',
			legal_sex: 'F',
			password: ''
		};
		editingStudent = null;
		error = '';
		successMessage = '';
	}

	function editStudent(student: Student) {
		editingStudent = student;
		studentForm = {
			id: student.id.toString(),
			name: student.name,
			grade: student.grade.toString(),
			legal_sex: student.legal_sex,
			password: ''
		};
	}

	function cancelEdit() {
		resetForm();
	}

	async function saveStudent() {
		loading = true;
		error = '';
		successMessage = '';

		try {
			const formData = new FormData();
			formData.append('id', studentForm.id);
			formData.append('name', studentForm.name);
			formData.append('grade', studentForm.grade);
			formData.append('legal_sex', studentForm.legal_sex);
			
			if (!editingStudent) {
				if (!studentForm.password) {
					throw new Error('Password is required for new students');
				}
				formData.append('password', studentForm.password);
			}

			const method = editingStudent ? 'PUT' : 'POST';
			const response = await fetch('/api/admin/students', {
				method,
				body: formData
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText);
			}

			successMessage = `Student ${editingStudent ? 'updated' : 'created'} successfully`;
			resetForm();
			onUpdate();
		} catch (err: any) {
			error = err.message || 'Failed to save student';
		} finally {
			loading = false;
		}
	}

	async function deleteStudent(student: Student) {
		loading = true;
		error = '';

		try {
			const response = await fetch(`/api/admin/students?id=${student.id}`, {
				method: 'DELETE'
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText);
			}

			confirmingDelete = null;
			successMessage = 'Student deleted successfully';
			onUpdate();
		} catch (err: any) {
			error = err.message || 'Failed to delete student';
		} finally {
			loading = false;
		}
	}

	async function deleteAllStudents() {
		loading = true;
		error = '';

		try {
			const response = await fetch('/api/admin/students/delete-all', {
				method: 'DELETE'
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText);
			}

			confirmingDeleteAll = false;
			successMessage = 'All students deleted successfully';
			onUpdate();
		} catch (err: any) {
			error = err.message || 'Failed to delete all students';
		} finally {
			loading = false;
		}
	}

	async function handleCSVUpload(event: Event) {
		const input = event.target as HTMLInputElement;
		csvFile = input.files?.[0] || null;
		csvPreview = null;
		
		if (csvFile) {
			await previewCSV();
		}
	}

	async function previewCSV() {
		if (!csvFile) return;

		loading = true;
		error = '';

		try {
			const formData = new FormData();
			formData.append('csv', csvFile);

			const response = await fetch('/api/admin/students/csv/preview', {
				method: 'POST',
				body: formData
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText);
			}

			csvPreview = await response.json();
		} catch (err: any) {
			error = err.message || 'Failed to preview CSV';
		} finally {
			loading = false;
		}
	}

	async function uploadCSV() {
		if (!csvFile || !csvPreview?.success) return;

		csvUploading = true;
		error = '';

		try {
			const formData = new FormData();
			formData.append('csv', csvFile);

			const response = await fetch('/api/admin/students/csv/upload', {
				method: 'POST',
				body: formData
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText);
			}

			const result = await response.json();
			successMessage = `Successfully imported ${result.imported} students`;
			csvFile = null;
			csvPreview = null;
			onUpdate();
		} catch (err: any) {
			error = err.message || 'Failed to upload CSV';
		} finally {
			csvUploading = false;
		}
	}

	function downloadExample() {
		window.location.href = '/api/admin/students/csv/example';
	}

	function downloadStudents() {
		window.location.href = '/api/admin/students/csv/download';
	}
</script>

<div class="students-manager">
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

	<!-- Bulk Operations Section -->
	<details class="management-section">
		<summary class="management-section-header">
			Bulk Operations
		</summary>
		<div class="management-section-content">
			<div class="management-actions">
				<button
					onclick={downloadExample}
					class="btn btn-outline"
					disabled={loading}
				>
					Download Example CSV
				</button>
				<button
					onclick={downloadStudents}
					class="btn btn-outline"
					disabled={loading || students.length === 0}
				>
					Download Current CSV
				</button>
			</div>
			
			<!-- CSV Upload Section - always shown when bulk operations expanded -->
			<div class="students-csv-upload-section">
				<h4 class="students-csv-upload-title">Upload Students CSV</h4>
				{#if error}
					<div class="alert alert-error">
						{error}
					</div>
				{/if}
				<div class="form-group">
					<label class="form-label-with-description">
						Select CSV File
						<small class="students-csv-help">
							(id, name, grade, legal_sex, password)
						</small>
					</label>
					<input
						type="file"
						accept=".csv"
						onchange={handleCSVUpload}
						disabled={loading || csvUploading}
						class="input"
					/>
					<p class="form-help-text">
						Expected columns: id, name, grade, legal_sex, password
					</p>
				</div>

				{#if csvPreview}
					<div class="students-csv-preview-container">
						<div class="students-csv-preview-header">
							<div class="students-csv-preview-title">
								<h5 class="students-csv-preview-heading">Preview Results</h5>
								<span class="badge {csvPreview.has_errors ? 'badge-danger' : 'badge-success'}">
									{csvPreview.has_errors ? 'Has Errors' : 'Ready to Import'}
								</span>
							</div>
						</div>
						<div class="students-csv-preview-body">
							{#if csvPreview.has_errors}
								<div class="alert alert-error">
									Found {csvPreview.preview.filter((r: any) => !r.is_valid).length} errors. Please fix them before uploading.
								</div>
							{:else}
								<div class="alert alert-success">
									Ready to import {csvPreview.total_rows} students
								</div>
							{/if}

							<div class="students-csv-preview-table-container">
								<table class="table">
									<thead class="table-header">
										<tr>
											<th class="table-header-cell">Row</th>
											<th class="table-header-cell">ID</th>
											<th class="table-header-cell">Name</th>
											<th class="table-header-cell">Grade</th>
											<th class="table-header-cell">Legal Sex</th>
											<th class="table-header-cell">Status</th>
											<th class="table-header-cell">Errors</th>
										</tr>
									</thead>
									<tbody class="table-body">
										{#each csvPreview.preview as row}
											<tr class="table-row {row.is_valid ? '' : 'table-row--error'}">
												<td class="table-cell">{row.row_number}</td>
												<td class="table-cell">{row.data.id || '-'}</td>
												<td class="table-cell">{row.data.name || '-'}</td>
												<td class="table-cell">{row.data.grade || '-'}</td>
												<td class="table-cell">{row.data.legal_sex || '-'}</td>
												<td class="table-cell">
													{#if row.is_valid}
														<span class="badge {row.will_update ? 'badge-warning' : 'badge-success'}">
															{row.will_update ? 'Update' : 'Create'}
														</span>
													{:else}
														<span class="badge badge-danger">Invalid</span>
													{/if}
												</td>
												<td class="table-cell-secondary">
													{#if row.errors.length > 0}
														<ul class="students-csv-error-list">
															{#each row.errors as error}
																<li>‚Ä {error}</li>
															{/each}
														</ul>
													{:else}
														<span class="students-csv-success-mark">‚úì</span>
													{/if}
												</td>
											</tr>
										{/each}
									</tbody>
								</table>
							</div>

							<div class="students-csv-preview-actions">
								<button
									onclick={uploadCSV}
									disabled={csvUploading || csvPreview.has_errors}
									class="btn btn-primary"
								>
									{csvUploading ? 'Importing...' : `Import ${csvPreview.total_rows} Students`}
								</button>
							</div>
						</div>
					</div>
				{/if}
			</div>
			
			<!-- Delete All Section -->
			<details class="management-section management-section--danger">
				<summary class="management-section-header">
					‚öÔ∏è Danger Zone
				</summary>
				<div class="management-section-content">
					{#if confirmingDeleteAll}
						<div class="students-delete-all-actions">
							<button
								onclick={deleteAllStudents}
								class="btn btn-danger btn-sm"
								disabled={loading}
							>
								{loading ? 'Deleting...' : 'Confirm Delete All'}
							</button>
							<button
								onclick={() => confirmingDeleteAll = false}
								class="btn btn-secondary btn-sm"
								disabled={loading}
							>
								Cancel
							</button>
						</div>
					{:else}
						<button
							onclick={() => confirmingDeleteAll = true}
							class="btn btn-danger btn-sm"
							disabled={loading || students.length === 0}
						>
							Delete All Students
						</button>
					{/if}
				</div>
			</details>
		</div>
	</details>

	<!-- Add Student Section -->
	<details class="management-section">
		<summary class="management-section-header">
			{editingStudent ? `Edit Student: ${editingStudent.name}` : 'Add New Student'}
		</summary>
		<div class="management-section-content">
			{#if error}
				<div class="alert alert-error">
					{error}
				</div>
			{/if}
			<form onsubmit={(e) => { e.preventDefault(); saveStudent(); }} class="students-form">
				<div class="students-form-row">
					<div class="form-group">
						<label class="form-label">Student ID *</label>
						<input
							bind:value={studentForm.id}
							type="number"
							required
							disabled={editingStudent !== null || loading}
							class="input"
						/>
					</div>

					<div class="form-group">
						<label class="form-label">Name *</label>
						<input
							bind:value={studentForm.name}
							type="text"
							required
							disabled={loading}
							class="input"
						/>
					</div>
				</div>

				<div class="students-form-row">
					<div class="form-group">
						<label class="form-label">Grade *</label>
						<select
							bind:value={studentForm.grade}
							required
							disabled={loading}
							class="select"
						>
							<option value="">Select grade</option>
							{#each grades as grade}
								<option value={grade}>{grade}</option>
							{/each}
						</select>
					</div>

					<div class="form-group">
						<label class="form-label">Legal Sex *</label>
						<select
							bind:value={studentForm.legal_sex}
							required
							disabled={loading}
							class="select"
						>
							<option value="F">F</option>
							<option value="M">M</option>
							<option value="X">X</option>
						</select>
					</div>
				</div>

				{#if !editingStudent}
					<div class="form-group">
						<label class="form-label">Password *</label>
						<input
							bind:value={studentForm.password}
							type="password"
							required
							disabled={loading}
							class="input"
						/>
					</div>
				{/if}

				<div class="form-actions">
					<button
						type="button"
						onclick={resetForm}
						class="btn btn-secondary"
						disabled={loading}
					>
						{editingStudent ? 'Cancel Edit' : 'Clear'}
					</button>
					<button
						type="submit"
						class="btn btn-primary"
						disabled={loading}
					>
						{loading ? 'Saving...' : editingStudent ? 'Update Student' : 'Create Student'}
					</button>
				</div>
			</form>
		</div>
	</details>

	<!-- Search and Controls -->
	{#if students.length > 0}
		<div class="students-search-section">
			<input
				type="text"
				bind:value={searchTerm}
				placeholder="Search students by name, ID, grade, legal sex..."
				class="students-search-input"
			/>
		</div>
	{/if}

	<div class="students-list-section">
		{#if filteredStudents.length === 0}
			{#if searchTerm}
				<div class="students-empty-state">
					<p class="students-empty-message">No students found matching "{searchTerm}"</p>
					<button onclick={() => searchTerm = ''} class="btn btn-secondary btn-sm">
						Clear search
					</button>
				</div>
			{:else if students.length === 0}
				<div class="students-empty-state">
					<p class="students-empty-message">No students found.</p>
				</div>
			{:else}
				<div class="students-empty-state">
					<p class="students-empty-message">No students match your search</p>
					<button onclick={() => searchTerm = ''} class="btn btn-secondary btn-sm">
						Clear search
					</button>
				</div>
			{/if}
		{:else}
			<div class="students-table-container">
				<table class="table">
					<thead class="table-header">
						<tr>
							<th class="table-header-cell">ID</th>
							<th class="table-header-cell">Name</th>
							<th class="table-header-cell">Grade</th>
							<th class="table-header-cell">Legal Sex</th>
							<th class="table-header-cell">Actions</th>
						</tr>
					</thead>
					<tbody class="table-body">
						{#each filteredStudents as student}
							<tr class="table-row">
								<td class="table-cell students-table-id">
									{student.id}
								</td>
								<td class="table-cell">
									{student.name}
								</td>
								<td class="table-cell">
									{student.grade}
								</td>
								<td class="table-cell">
									{student.legal_sex}
								</td>
								<td class="table-cell students-table-actions">
									{#if confirmingDelete === student}
										<div class="students-delete-confirmation">
											<span class="students-delete-prompt">Delete {student.name}?</span>
											<button
												onclick={() => deleteStudent(student)}
												class="students-delete-confirm-btn"
												disabled={loading}
											>
												Confirm
											</button>
											<button
												onclick={() => confirmingDelete = null}
												class="students-delete-cancel-btn"
												disabled={loading}
											>
												Cancel
											</button>
										</div>
									{:else}
										<button
											onclick={() => editStudent(student)}
											class="students-edit-btn"
											disabled={loading}
										>
											Edit
										</button>
										<button
											onclick={() => confirmingDelete = student}
											class="students-delete-btn"
											disabled={loading}
										>
											Delete
										</button>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>

</div>
