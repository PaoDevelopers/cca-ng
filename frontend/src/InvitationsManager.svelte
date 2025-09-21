<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { api, handleAPIError } from './api';
	import { getSSEClient } from './sse';
	import type { Invitation, Student, Course, SSEMessage } from './types';

	interface Props {
		students: Student[];
		courses: Course[];
		onUpdate: () => void;
	}

	const { students, courses, onUpdate }: Props = $props();

	let invitations = $state<Invitation[]>([]);
	let loading = $state<boolean>(false);
	let error = $state<string>('');
	let csvPreview = $state<any>(null);
	let csvFile = $state<File | null>(null);

	let editingInvitation = $state<Invitation | null>(null);
	let formStudentId = $state<number>(0);
	let formCourseId = $state<string>('');
	let formInvitationType = $state<'invite' | 'force'>('invite');
	let formError = $state<string>('');
	let formLoading = $state<boolean>(false);

	let studentSearchTerm = $state<string>('');
	let courseSearchTerm = $state<string>('');
	let showStudentList = $state<boolean>(false);
	let showCourseList = $state<boolean>(false);

	let filteredStudentsForm = $derived.by(() => {
		if (!studentSearchTerm) return students.slice(0, 10);
		const term = studentSearchTerm.toLowerCase();
		return students.filter(student => 
			student.name.toLowerCase().includes(term) ||
			student.id.toString().includes(term)
		).slice(0, 10);
	});

	let filteredCoursesForm = $derived.by(() => {
		if (!courseSearchTerm) return courses.slice(0, 10);
		const term = courseSearchTerm.toLowerCase();
		return courses.filter(course => 
			course.name.toLowerCase().includes(term) ||
			course.id.toLowerCase().includes(term) ||
			course.period_id.toLowerCase().includes(term)
		).slice(0, 10);
	});

	let searchTerm = $state('');
	let filteredInvitations = $derived.by(() => {
		if (!invitations || !Array.isArray(invitations)) return [];
		if (!searchTerm) return invitations;
		const term = searchTerm.toLowerCase();
		return invitations.filter(invitation => 
			invitation.student_name.toLowerCase().includes(term) ||
			invitation.course_name.toLowerCase().includes(term) ||
			invitation.course_id.toLowerCase().includes(term) ||
			invitation.student_id.toString().includes(term) ||
			invitation.course_period.toLowerCase().includes(term)
		);
	});

	let confirmingDelete = $state<Invitation | null>(null);

	let sseEventHandler: ((message: SSEMessage) => void) | null = null;

	onMount(() => {
		loadInvitations();
		
		const sseClient = getSSEClient();
		if (sseClient) {
			sseEventHandler = (message: SSEMessage) => {
				if (message.type === 'invalidate_invitations') {
					loadInvitations();
				}
			};
			sseClient.on('invalidate_invitations', sseEventHandler);
		}
	});

	onDestroy(() => {
		if (sseEventHandler) {
			const sseClient = getSSEClient();
			if (sseClient) {
				sseClient.off('invalidate_invitations', sseEventHandler);
			}
		}
	});

	async function loadInvitations(): Promise<void> {
		loading = true;
		error = '';
		try {
			invitations = await api.admin.getInvitations();
		} catch (err) {
			error = handleAPIError(err);
		} finally {
			loading = false;
		}
	}

	function openAddForm(): void {
		editingInvitation = null;
		formStudentId = 0;
		formCourseId = '';
		formInvitationType = 'invite';
		formError = '';
		studentSearchTerm = '';
		courseSearchTerm = '';
		showStudentList = false;
		showCourseList = false;
	}

	function editInvitation(invitation: Invitation): void {
		editingInvitation = invitation;
		formStudentId = invitation.student_id;
		formCourseId = invitation.course_id;
		formInvitationType = invitation.invitation_type as 'invite' | 'force';
		formError = '';
		
		const selectedStudent = students.find(s => s.id === invitation.student_id);
		const selectedCourse = courses.find(c => c.id === invitation.course_id);
		studentSearchTerm = selectedStudent ? `${selectedStudent.id} - ${selectedStudent.name}` : '';
		courseSearchTerm = selectedCourse ? `${selectedCourse.id} - ${selectedCourse.name}` : '';
		showStudentList = false;
		showCourseList = false;
	}

	function cancelForm(): void {
		editingInvitation = null;
		formStudentId = 0;
		formCourseId = '';
		formInvitationType = 'invite';
		formError = '';
		studentSearchTerm = '';
		courseSearchTerm = '';
		showStudentList = false;
		showCourseList = false;
	}

	async function saveInvitation(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		
		if (!formStudentId || !formCourseId) {
			formError = 'Student and Course are required';
			return;
		}

		formLoading = true;
		formError = '';

		try {
			const formData = new FormData();
			formData.append('student_id', formStudentId.toString());
			formData.append('course_id', formCourseId);
			formData.append('invitation_type', formInvitationType);

			if (editingInvitation) {
				await api.admin.updateInvitation(formData);
			} else {
				await api.admin.createInvitation(formData);
			}

			await loadInvitations();
			cancelForm();
		} catch (err) {
			formError = handleAPIError(err);
		} finally {
			formLoading = false;
		}
	}

	async function deleteInvitation(invitation: Invitation): Promise<void> {
		try {
			await api.admin.deleteInvitation(invitation.student_id, invitation.course_id);
			confirmingDelete = null;
			await loadInvitations();
		} catch (err) {
			error = handleAPIError(err);
		}
	}

	function handleFileSelect(event: Event): void {
		const target = event.target as HTMLInputElement;
		csvFile = target.files?.[0] || null;
		csvPreview = null;
	}

	async function previewCSV(): Promise<void> {
		if (!csvFile) return;

		loading = true;
		error = '';
		csvPreview = null;

		try {
			const formData = new FormData();
			formData.append('csv', csvFile);

			csvPreview = await api.admin.previewInvitationsCSV(formData);
		} catch (err) {
			error = handleAPIError(err);
		} finally {
			loading = false;
		}
	}

	async function uploadCSV(): Promise<void> {
		if (!csvFile || !csvPreview?.success) return;

		loading = true;
		error = '';

		try {
			const formData = new FormData();
			formData.append('csv', csvFile);

			await api.admin.uploadInvitationsCSV(formData);
			await loadInvitations();
			
			csvFile = null;
			csvPreview = null;
		} catch (err) {
			error = handleAPIError(err);
		} finally {
			loading = false;
		}
	}

	function downloadExample(): void {
		window.open('/api/admin/invitations/csv/example', '_blank');
	}

	function downloadInvitations(): void {
		window.open('/api/admin/invitations/csv/download', '_blank');
	}

	function getStudentName(studentId: number): string {
		return students.find(s => s.id === studentId)?.name || `Student ${studentId}`;
	}

	function getCourseName(courseId: string): string {
		return courses.find(c => c.id === courseId)?.name || courseId;
	}

	function selectStudent(student: Student): void {
		formStudentId = student.id;
		studentSearchTerm = `${student.id} - ${student.name}`;
		showStudentList = false;
	}

	function selectCourse(course: Course): void {
		formCourseId = course.id;
		courseSearchTerm = `${course.id} - ${course.name}`;
		showCourseList = false;
	}

	function clearStudentSelection(): void {
		formStudentId = 0;
		studentSearchTerm = '';
		showStudentList = false;
	}

	function clearCourseSelection(): void {
		formCourseId = '';
		courseSearchTerm = '';
		showCourseList = false;
	}
</script>

<div class="invitations-manager">
	{#if error}
		<div class="alert alert-error">
			{error}
		</div>
	{/if}

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
				{#if invitations && invitations.length > 0}
					<button onclick={downloadInvitations} class="btn btn-outline">
						Download Current CSV
					</button>
				{/if}
			</div>
			
			<!-- CSV Import Section - always shown when bulk operations expanded -->
			<div class="students-csv-upload-section">
				<h4 class="students-csv-upload-title">Import Invitations CSV</h4>
				<div class="form-group">
					<label class="form-label-with-description">
						Select CSV File
						<small class="students-csv-help">
							(student_id, course_id, invitation_type)
						</small>
					</label>
					<input
						type="file"
						accept=".csv"
						onchange={handleFileSelect}
						class="input"
					/>
				</div>

				{#if csvFile}
					<button
						onclick={previewCSV}
						disabled={loading}
						class="btn btn-secondary btn-sm"
					>
						{loading ? 'Processing...' : 'Preview CSV'}
					</button>
				{/if}

				{#if csvPreview}
					<div class="students-csv-preview-container">
						<div class="students-csv-preview-header">
							<div class="students-csv-preview-title">
								<h5 class="students-csv-preview-heading">
									Preview Results: {csvPreview.total_rows} rows
								</h5>
								<span class="badge {csvPreview.has_errors ? 'badge-danger' : 'badge-success'}">
									{csvPreview.has_errors ? 'Has Errors' : 'No Errors'}
								</span>
							</div>
						</div>
						<div class="students-csv-preview-body">
							{#if csvPreview.has_errors}
								<div class="alert alert-error">
									Found {csvPreview.preview.filter((r) => !r.is_valid).length} errors. Please fix them before uploading.
								</div>
							{:else}
								<div class="alert alert-success">
									Ready to import {csvPreview.total_rows} invitations
								</div>
							{/if}

							<div class="students-csv-preview-table-container">
								<table class="table">
									<thead class="table-header">
										<tr>
											<th class="table-header-cell">Row</th>
											<th class="table-header-cell">Student</th>
											<th class="table-header-cell">Course</th>
											<th class="table-header-cell">Type</th>
											<th class="table-header-cell">Status</th>
											<th class="table-header-cell">Errors</th>
										</tr>
									</thead>
									<tbody class="table-body">
										{#each csvPreview.preview as row}
											<tr class="table-row {row.is_valid ? '' : 'table-row--error'}">
												<td class="table-cell">{row.row_number}</td>
												<td class="table-cell">{row.data.student_id || '-'}</td>
												<td class="table-cell">{row.data.course_id || '-'}</td>
												<td class="table-cell">{row.data.invitation_type || '-'}</td>
												<td class="table-cell">
													{#if row.will_update}
														<span class="badge badge-warning">Update</span>
													{:else if row.is_valid}
														<span class="badge badge-success">Create</span>
													{:else}
														<span class="badge badge-danger">Error</span>
													{/if}
												</td>
												<td class="table-cell-secondary">
													{row.errors.join(', ')}
												</td>
											</tr>
										{/each}
									</tbody>
								</table>
							</div>

							{#if csvPreview.success}
								<div class="students-csv-preview-actions">
									<button
										onclick={uploadCSV}
										disabled={loading}
										class="btn btn-primary"
									>
										{loading ? 'Importing...' : `Import ${csvPreview.total_rows} Invitations`}
									</button>
								</div>
							{/if}
						</div>
					</div>
				{/if}
			</div>
		</div>
	</details>

	<!-- Add Invitation Section -->
	<details class="management-section">
		<summary class="management-section-header">
			{editingInvitation ? `Edit Invitation: ${editingInvitation.student_name} â†’ ${editingInvitation.course_name}` : 'Add New Invitation'}
		</summary>
		<div class="management-section-content">
			{#if formError}
				<div class="alert alert-error">
					{formError}
				</div>
			{/if}

			<form onsubmit={saveInvitation} class="invitations-form">
				<div class="invitations-form-fields">
					<div class="form-group">
						<label class="form-label">Student *</label>
						<div class="searchable-select">
							<input
								type="text"
								bind:value={studentSearchTerm}
								onfocus={() => showStudentList = true}
								oninput={() => showStudentList = true}
								placeholder="Search students by name or ID..."
								required={!formStudentId}
								disabled={formLoading}
								class="input"
							/>
							{#if formStudentId}
								<button
									type="button"
									onclick={clearStudentSelection}
									class="clear-selection-btn"
									disabled={formLoading}
								>
									Clear
								</button>
							{/if}
							{#if showStudentList}
								<div class="searchable-list">
									{#if filteredStudentsForm.length === 0}
										<div class="searchable-list-empty">
											No students found
										</div>
									{:else}
										{#each filteredStudentsForm as student}
											<div
												class="searchable-list-item"
												onclick={() => selectStudent(student)}
											>
												<div class="searchable-list-item-title">
													{student.name}
												</div>
												<div class="searchable-list-item-subtitle">
													ID: {student.id} â€ Grade: {student.grade}
												</div>
											</div>
										{/each}
									{/if}
									<button
										type="button"
										onclick={() => showStudentList = false}
										class="searchable-list-close"
									>
										Close
									</button>
								</div>
							{/if}
						</div>
					</div>

					<div class="form-group">
						<label class="form-label">Course *</label>
						<div class="searchable-select">
							<input
								type="text"
								bind:value={courseSearchTerm}
								onfocus={() => showCourseList = true}
								oninput={() => showCourseList = true}
								placeholder="Search courses by name, ID, or period..."
								required={!formCourseId}
								disabled={formLoading}
								class="input"
							/>
							{#if formCourseId}
								<button
									type="button"
									onclick={clearCourseSelection}
									class="clear-selection-btn"
									disabled={formLoading}
								>
									Clear
								</button>
							{/if}
							{#if showCourseList}
								<div class="searchable-list">
									{#if filteredCoursesForm.length === 0}
										<div class="searchable-list-empty">
											No courses found
										</div>
									{:else}
										{#each filteredCoursesForm as course}
											<div
												class="searchable-list-item"
												onclick={() => selectCourse(course)}
											>
												<div class="searchable-list-item-title">
													{course.name}
												</div>
												<div class="searchable-list-item-subtitle">
													{course.id} â€ {course.period_id}
												</div>
											</div>
										{/each}
									{/if}
									<button
										type="button"
										onclick={() => showCourseList = false}
										class="searchable-list-close"
									>
										Close
									</button>
								</div>
							{/if}
						</div>
					</div>

					<div class="form-group">
						<label class="form-label">Type *</label>
						<select
							bind:value={formInvitationType}
							required
							disabled={formLoading}
							class="select"
						>
							<option value="invite">Invite (can deselect with confirmation)</option>
							<option value="force">Force (cannot deselect)</option>
						</select>
					</div>
				</div>

				<div class="form-actions">
					<button
						type="button"
						onclick={cancelForm}
						disabled={formLoading}
						class="btn btn-secondary"
					>
						{editingInvitation ? 'Cancel Edit' : 'Clear'}
					</button>
					<button
						type="submit"
						disabled={formLoading}
						class="btn btn-primary"
					>
						{formLoading ? 'Saving...' : (editingInvitation ? 'Update' : 'Create')}
					</button>
				</div>
			</form>
		</div>
	</details>

	<!-- Invitations List -->
	<div class="invitations-list-section">
		<div class="invitations-list-header">
			{#if invitations && invitations.length > 0}
				<div class="invitations-search-section">
					<input
						type="text"
						bind:value={searchTerm}
						placeholder="Search invitations by student, course, ID..."
						class="invitations-search-input"
					/>
				</div>
			{/if}
		</div>
		{#if loading}
			<div class="invitations-loading">
				<div class="invitations-loading-text">Loading invitations...</div>
			</div>
		{:else if filteredInvitations.length === 0}
			{#if searchTerm}
				<div class="invitations-empty-state">
					<p class="invitations-empty-message">No invitations found matching "{searchTerm}"</p>
					<button onclick={() => searchTerm = ''} class="btn btn-secondary btn-sm">
						Clear search
					</button>
				</div>
			{:else}
				<div class="invitations-empty-state">
					<div class="invitations-empty-message">
						No invitations found. Add some invitations or import from CSV.
					</div>
				</div>
			{/if}
		{:else}
			<div class="invitations-table-container">
				<table class="table">
					<thead class="table-header">
						<tr>
							<th class="table-header-cell">Student</th>
							<th class="table-header-cell">Course</th>
							<th class="table-header-cell">Period</th>
							<th class="table-header-cell">Type</th>
							<th class="table-header-cell">Actions</th>
						</tr>
					</thead>
					<tbody class="table-body">
						{#each filteredInvitations as invitation}
							<tr class="table-row">
								<td class="table-cell">
									<div class="invitations-student-info">
										<div class="invitations-student-name">
											{invitation.student_name}
										</div>
										<div class="invitations-student-details">
											ID: {invitation.student_id} â€ Grade: {invitation.student_grade}
										</div>
									</div>
								</td>
								<td class="table-cell">
									<div class="invitations-course-info">
										<div class="invitations-course-name">
											{invitation.course_name}
										</div>
										<div class="invitations-course-id">
											{invitation.course_id}
										</div>
									</div>
								</td>
								<td class="table-cell">
									{invitation.course_period}
								</td>
								<td class="table-cell">
									<span class="badge {invitation.invitation_type === 'force' 
										? 'badge-danger' 
										: 'badge-primary'}">
										{invitation.invitation_type === 'force' ? 'Forced' : 'Invited'}
									</span>
								</td>
								<td class="table-cell invitations-table-actions">
									{#if confirmingDelete === invitation}
										<div class="invitations-delete-confirmation">
											<span class="invitations-delete-prompt">Delete invitation?</span>
											<button
												onclick={() => deleteInvitation(invitation)}
												class="invitations-delete-confirm-btn"
											>
												Confirm
											</button>
											<button
												onclick={() => confirmingDelete = null}
												class="invitations-delete-cancel-btn"
											>
												Cancel
											</button>
										</div>
									{:else}
										<button
											onclick={() => editInvitation(invitation)}
											class="invitations-edit-btn"
										>
											Edit
										</button>
										<button
											onclick={() => confirmingDelete = invitation}
											class="invitations-delete-btn"
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
