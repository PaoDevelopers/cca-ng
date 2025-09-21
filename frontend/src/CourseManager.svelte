<script lang="ts">
	interface Course {
		id: string;
		name: string;
		description: string;
		period_id: string;
		max_students: number;
		sex_restriction: string;
		membership: string;
		teacher: string;
		location: string;
		category_id: string;
		allowed_grades: number[] | null;
	}

	interface Props {
		courses: Course[];
		periods: string[];
		categories: string[];
		grades: number[];
		onUpdate: () => void;
	}

	const { courses, periods, categories, grades, onUpdate }: Props = $props();

	let showAddForm = $state(false);
	let editingCourse = $state<Course | null>(null);
	let error = $state('');
	let loading = $state(false);
	let showDeleteAllConfirm = $state(false);
	let confirmingDelete = $state<Course | null>(null);

	let searchTerm = $state('');
	let filteredCourses = $derived.by(() => {
		if (!searchTerm) return courses;
		const term = searchTerm.toLowerCase();
		return courses.filter(course => 
			course.name.toLowerCase().includes(term) ||
			course.id.toLowerCase().includes(term) ||
			course.teacher.toLowerCase().includes(term) ||
			course.location.toLowerCase().includes(term) ||
			course.category_id.toLowerCase().includes(term) ||
			course.period_id.toLowerCase().includes(term) ||
			(course.description && course.description.toLowerCase().includes(term))
		);
	});

	let csvFile = $state<File | null>(null);
	let csvError = $state('');
	let csvLoading = $state(false);
	let csvPreview = $state<any>(null);
	let csvPreviewLoading = $state(false);

	let formData = $state({
		id: '',
		name: '',
		description: '',
		period_id: '',
		max_students: 20,
		sex_restriction: 'ANY',
		membership: 'free',
		teacher: '',
		location: '',
		category_id: '',
		allowed_grades: [] as number[]
	});

	function resetForm() {
		formData = {
			id: '',
			name: '',
			description: '',
			period_id: '',
			max_students: 20,
			sex_restriction: 'ANY',
			membership: 'free',
			teacher: '',
			location: '',
			category_id: '',
			allowed_grades: []
		};
		editingCourse = null;
		error = '';
	}

	function startEdit(course: Course) {
		formData = {
			id: course.id,
			name: course.name,
			description: course.description,
			period_id: course.period_id,
			max_students: course.max_students,
			sex_restriction: course.sex_restriction,
			membership: course.membership,
			teacher: course.teacher,
			location: course.location,
			category_id: course.category_id,
			allowed_grades: [...(course.allowed_grades || [])]
		};
		editingCourse = course;
		error = '';
	}

	async function saveCourse() {
		if (!formData.id.trim() || !formData.name.trim() || !formData.period_id || !formData.category_id) {
			error = 'ID, name, period, and category are required';
			return;
		}

		loading = true;
		error = '';

		try {
			const data = new FormData();
			data.append('id', formData.id.trim());
			data.append('name', formData.name.trim());
			data.append('description', formData.description);
			data.append('period_id', formData.period_id);
			data.append('max_students', String(formData.max_students));
			data.append('sex_restriction', formData.sex_restriction);
			data.append('membership', formData.membership);
			data.append('teacher', formData.teacher);
			data.append('location', formData.location);
			data.append('category_id', formData.category_id);
			data.append('allowed_grades', formData.allowed_grades.join(','));

			const method = editingCourse ? 'PUT' : 'POST';
			const response = await fetch('/api/admin/courses', {
				method,
				body: data
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText);
			}

			resetForm();
			onUpdate();
		} catch (err: any) {
			error = err.message || 'Failed to save course';
		} finally {
			loading = false;
		}
	}

	async function deleteCourse(course: Course) {
		loading = true;
		error = '';

		try {
			const response = await fetch(`/api/admin/courses?id=${encodeURIComponent(course.id)}`, {
				method: 'DELETE'
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText);
			}

			confirmingDelete = null;
			onUpdate();
		} catch (err: any) {
			error = err.message || 'Failed to delete course';
		} finally {
			loading = false;
		}
	}

	async function deleteAllCourses() {
		loading = true;
		error = '';

		try {
			const response = await fetch('/api/admin/courses/delete-all', {
				method: 'DELETE'
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText);
			}

			showDeleteAllConfirm = false;
			onUpdate();
		} catch (err: any) {
			error = err.message || 'Failed to delete all courses';
		} finally {
			loading = false;
		}
	}

	function handleGradeToggle(grade: number) {
		if (formData.allowed_grades.includes(grade)) {
			formData.allowed_grades = formData.allowed_grades.filter(g => g !== grade);
		} else {
			formData.allowed_grades = [...formData.allowed_grades, grade].sort((a, b) => a - b);
		}
	}

	function handleCSVFileChange(event: Event) {
		const target = event.target as HTMLInputElement;
		csvFile = target.files?.[0] || null;
		csvError = '';
		csvPreview = null;
	}

	async function previewCSV() {
		if (!csvFile) {
			csvError = 'Please select a file';
			return;
		}

		csvPreviewLoading = true;
		csvError = '';
		csvPreview = null;

		try {
			const data = new FormData();
			data.append('csv', csvFile);

			const response = await fetch('/api/admin/courses/csv/preview', {
				method: 'POST',
				body: data
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText);
			}

			csvPreview = await response.json();
		} catch (err: any) {
			csvError = err.message || 'Failed to preview CSV';
		} finally {
			csvPreviewLoading = false;
		}
	}

	async function commitCSV() {
		if (!csvFile || !csvPreview || csvPreview.has_errors) {
			csvError = 'Cannot commit CSV with errors';
			return;
		}

		csvLoading = true;
		csvError = '';

		try {
			const data = new FormData();
			data.append('csv', csvFile);

			const response = await fetch('/api/admin/courses/csv/upload', {
				method: 'POST',
				body: data
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText);
			}

			const result = await response.json();
			csvFile = null;
			csvPreview = null;
			showCSVUpload = false;
			onUpdate();
			alert(`Successfully imported ${result.imported} courses`);
		} catch (err: any) {
			csvError = err.message || 'Failed to commit CSV';
		} finally {
			csvLoading = false;
		}
	}

	function downloadExample() {
		window.open('/api/admin/courses/csv/example', '_blank');
	}

	function downloadCSV() {
		window.open('/api/admin/courses/csv/download', '_blank');
	}
</script>
<div>
	<!-- Bulk Operations Section -->
	<details class="management-section">
		<summary class="management-section-header">
			Bulk Operations
		</summary>
		<div class="management-section-content">
			<div class="management-actions">
				<button onclick={downloadExample} class="btn btn-outline">
					Download Example CSV
				</button>
				<button onclick={downloadCSV} class="btn btn-outline" disabled={courses.length === 0}>
					Download Current CSV
				</button>
			</div>
			
			<!-- CSV Upload Section - always shown when bulk operations expanded -->
			<div class="course-csv-upload-section">
				<h4 class="course-csv-upload-title">Upload CSV File</h4>
				{#if csvError}
					<div class="course-csv-error">
						{csvError}
					</div>
				{/if}
				
				<div class="course-csv-upload-controls">
					<input
						type="file"
						accept=".csv"
						onchange={handleCSVFileChange}
						disabled={csvPreviewLoading || csvLoading}
						class="course-csv-file-input"
					/>
					<button
						onclick={previewCSV}
						disabled={csvPreviewLoading || !csvFile}
						class="course-csv-preview-button"
					>
						{csvPreviewLoading ? 'Parsing...' : 'Preview'}
					</button>
				</div>

				{#if csvPreview}
					<div class="course-csv-preview-container">
						<div class="course-csv-preview-header">
							<div class="course-csv-preview-title">
								<h5 class="course-csv-preview-heading">
									CSV Preview ({csvPreview.total_rows} rows)
								</h5>
								<div class="badge {csvPreview.has_errors ? 'badge-danger' : 'badge-success'}">
									{csvPreview.has_errors ? '‚ö Contains errors' : '‚úì Ready to import'}
								</div>
							</div>
						</div>
						<div class="course-csv-preview-body">
							<div class="course-csv-preview-table-container">
								<table class="table">
									<thead class="table-header">
										<tr>
											<th class="table-header-cell">Row</th>
											<th class="table-header-cell">ID</th>
											<th class="table-header-cell">Name</th>
											<th class="table-header-cell">Period</th>
											<th class="table-header-cell">Category</th>
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
												<td class="table-cell">{row.data.period_id || '-'}</td>
												<td class="table-cell">{row.data.category_id || '-'}</td>
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
														<ul class="course-csv-error-list">
															{#each row.errors as error}
																<li>‚Ä {error}</li>
															{/each}
														</ul>
													{:else}
														<span class="course-csv-success-mark">‚úì</span>
													{/if}
												</td>
											</tr>
										{/each}
									</tbody>
								</table>
							</div>
							
							<div class="course-csv-preview-actions">
								<button
									onclick={commitCSV}
									disabled={csvLoading || csvPreview.has_errors}
									class="btn btn-primary"
								>
									{csvLoading ? 'Importing...' : `Import ${csvPreview.total_rows} Courses`}
								</button>
							</div>
						</div>
					</div>
				{/if}
				
				<p class="course-csv-upload-help">
					Upload a CSV file with course data. Existing courses with matching IDs will be updated.
				</p>
			</div>
		</div>
	</details>

	<!-- Add Course Section -->
	<details class="management-section">
		<summary class="management-section-header">
			{editingCourse ? `Edit Course: ${editingCourse.name}` : 'Add New Course'}
		</summary>
		<div class="management-section-content">
			{#if error}
				<div class="alert alert-error">
					{error}
				</div>
			{/if}
			<div class="course-form-row">
				<div class="form-group">
					<label class="form-label">Course ID *</label>
					<input
						bind:value={formData.id}
						disabled={editingCourse !== null}
						class="input"
					/>
				</div>
				
				<div class="form-group">
					<label class="form-label">Name *</label>
					<input
						bind:value={formData.name}
						class="input"
					/>
				</div>
			</div>
			
			<div class="form-group">
				<label class="form-label">Description</label>
				<textarea
					bind:value={formData.description}
					rows="3"
					class="textarea"
				></textarea>
			</div>
			
			<div class="course-form-row">
				<div class="form-group">
					<label class="form-label">Period *</label>
					<select
						bind:value={formData.period_id}
						class="select"
					>
						<option value="">Select Period</option>
						{#each periods as period}
							<option value={period}>{period}</option>
						{/each}
					</select>
				</div>
				
				<div class="form-group">
					<label class="form-label">Category *</label>
					<select
						bind:value={formData.category_id}
						class="select"
					>
						<option value="">Select Category</option>
						{#each categories as category}
							<option value={category}>{category}</option>
						{/each}
					</select>
				</div>
			</div>
			
			<div class="course-form-row">
				<div class="form-group">
					<label class="form-label">Max Students</label>
					<input
						bind:value={formData.max_students}
						type="number"
						min="0"
						class="input"
					/>
				</div>
				
				<div class="form-group">
					<label class="form-label">Legal Sex Restriction</label>
					<select
						bind:value={formData.sex_restriction}
						class="select"
					>
						<option value="ANY">Any</option>
						<option value="F">Female</option>
						<option value="M">Male</option>
						<option value="X">Non-binary</option>
					</select>
				</div>
			</div>
			
			<div class="course-form-row">
				<div class="form-group">
					<label class="form-label">Teacher</label>
					<input
						bind:value={formData.teacher}
						class="input"
					/>
				</div>
				
				<div class="form-group">
					<label class="form-label">Location</label>
					<input
						bind:value={formData.location}
						class="input"
					/>
				</div>
			</div>
			
			<div class="course-form-row">
				<div class="form-group">
					<label class="form-label">Membership Type</label>
					<select
						bind:value={formData.membership}
						class="select"
					>
						<option value="free">Free Choice</option>
						<option value="invite_only">Invite Only</option>
					</select>
				</div>
				
				<div></div> <!-- Empty div for grid alignment -->
			</div>
			
			<div class="form-group">
				<label class="form-label-with-description">Allowed Grades</label>
				<div class="course-grade-checkboxes">
					{#each grades as grade}
						<label class="course-grade-checkbox">
							<input
								type="checkbox"
								checked={formData.allowed_grades.includes(grade)}
								onchange={() => handleGradeToggle(grade)}
								class="checkbox"
							/>
							<span class="course-grade-checkbox-label">Grade {grade}</span>
						</label>
					{/each}
				</div>
				<p class="form-help-text">Leave empty to allow all grades</p>
			</div>
			
			<div class="form-actions">
				<button
					onclick={resetForm}
					disabled={loading}
					class="btn btn-secondary"
				>
					{editingCourse ? 'Cancel Edit' : 'Clear'}
				</button>
				<button
					onclick={saveCourse}
					disabled={loading}
					class="btn btn-primary"
				>
					{loading ? 'Saving...' : (editingCourse ? 'Update Course' : 'Create Course')}
				</button>
			</div>
		</div>
	</details>

	<!-- Search and Controls -->
	<div class="course-search-controls">
		<div class="course-search-container">
			<input
				type="text"
				bind:value={searchTerm}
				placeholder="Search courses by name, ID, teacher, location..."
				class="input"
			/>
		</div>
	</div>

	<!-- Courses List -->
	{#if filteredCourses.length === 0}
		{#if searchTerm}
			<div class="course-empty-state">
				<div class="course-empty-message">No courses found matching "{searchTerm}"</div>
				<button onclick={() => searchTerm = ''} class="btn btn-secondary btn-sm">
					Clear search
				</button>
			</div>
		{:else if courses.length === 0}
			<div class="course-empty-state">
				<div class="course-empty-message">No courses defined yet</div>
				<p class="course-empty-description">Use "Add New Course" above to get started.</p>
			</div>
		{:else}
			<div class="course-empty-state">
				<div class="course-empty-message">No courses match your search</div>
				<button onclick={() => searchTerm = ''} class="btn btn-secondary btn-sm">
					Clear search
				</button>
			</div>
		{/if}
	{:else}
		<div class="course-list">
			{#each filteredCourses as course}
				<div class="course-item">
					<div class="course-item-header">
						<div class="course-item-info">
							<h4 class="course-item-name">{course.name}</h4>
							<p class="course-item-id">ID: {course.id}</p>
						</div>
						<div class="course-item-actions">
							{#if confirmingDelete === course}
								<div class="course-delete-confirmation">
									<span class="course-delete-prompt">Delete "{course.name}"?</span>
									<button
										onclick={() => deleteCourse(course)}
										disabled={loading}
										class="course-delete-confirm-btn"
									>
										Confirm
									</button>
									<button
										onclick={() => confirmingDelete = null}
										disabled={loading}
										class="course-delete-cancel-btn"
									>
										Cancel
									</button>
								</div>
							{:else}
								<button
									onclick={() => startEdit(course)}
									class="course-edit-btn"
								>
									Edit
								</button>
								<button
									onclick={() => confirmingDelete = course}
									class="course-delete-btn"
								>
									Delete
								</button>
							{/if}
						</div>
					</div>
					
					<div class="course-item-details">
						<div class="course-detail-item">
							<span class="course-detail-label">Period:</span>
							<span class="course-detail-value">{course.period_id}</span>
						</div>
						<div class="course-detail-item">
							<span class="course-detail-label">Category:</span>
							<span class="course-detail-value">{course.category_id}</span>
						</div>
						<div class="course-detail-item">
							<span class="course-detail-label">Max Students:</span>
							<span class="course-detail-value">{course.max_students}</span>
						</div>
						<div class="course-detail-item">
							<span class="course-detail-label">Type:</span>
							<span class="course-detail-value">{course.membership === 'free' ? 'Free Choice' : 'Invite Only'}</span>
						</div>
						<div class="course-detail-item">
							<span class="course-detail-label">Teacher:</span>
							<span class="course-detail-value">{course.teacher || 'N/A'}</span>
						</div>
						<div class="course-detail-item">
							<span class="course-detail-label">Location:</span>
							<span class="course-detail-value">{course.location || 'N/A'}</span>
						</div>
						<div class="course-detail-item">
							<span class="course-detail-label">Legal Sex:</span>
							<span class="course-detail-value">{course.sex_restriction === 'ANY' ? 'Any' : course.sex_restriction}</span>
						</div>
						<div class="course-detail-item">
							<span class="course-detail-label">Grades:</span>
							<span class="course-detail-value">
								{(course.allowed_grades && course.allowed_grades.length) ? course.allowed_grades.join(', ') : 'All'}
							</span>
						</div>
					</div>
					
					{#if course.description}
						<div class="course-item-description">
							<p class="course-description-text">{course.description}</p>
						</div>
					{/if}
				</div>
			{/each}
		</div>
		
		<div class="course-danger-zone-container">
			<details class="management-section management-section--danger">
				<summary class="management-section-header">
					‚öÔ∏è Danger Zone
				</summary>
				<div class="management-section-content">
					{#if showDeleteAllConfirm}
						<div class="course-delete-all-warning">
							<p class="course-delete-all-warning-text">
								This will permanently delete ALL courses and cannot be undone. All course data, enrollments, and selections will be lost.
							</p>
						</div>
						<div class="course-delete-all-actions">
							<button
								onclick={deleteAllCourses}
								disabled={loading}
								class="btn btn-danger btn-sm"
							>
								{loading ? 'Deleting...' : 'Permanently Delete All Courses'}
							</button>
							<button
								onclick={() => showDeleteAllConfirm = false}
								disabled={loading}
								class="btn btn-secondary btn-sm"
							>
								Cancel
							</button>
						</div>
					{:else}
						<button
							onclick={() => showDeleteAllConfirm = true}
							disabled={courses.length === 0}
							class="btn btn-danger btn-sm"
						>
							Delete All Courses
						</button>
					{/if}
				</div>
			</details>
		</div>
	{/if}
</div>
