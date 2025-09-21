<script lang="ts">
	import { onMount } from 'svelte';
	import { api, handleAPIError } from './api.ts';
	import type { Selection, Course, Student } from './types.ts';

	interface Props {
		students: Student[];
		courses: Course[];
		onUpdate: () => Promise<void>;
	}

	const { students, courses, onUpdate }: Props = $props();

	console.log(' SelectionsManager: Component initialized with props', {
		studentsLength: Array.isArray(students) ? students.length : 'not array',
		coursesLength: Array.isArray(courses) ? courses.length : 'not array',
		onUpdate: typeof onUpdate
	});

	console.log(' SelectionsManager: Props data debug', {
		students: students,
		courses: courses
	});
	
	console.log(' SelectionsManager: EMERGENCY - Direct courses access:', courses);
	if (courses && Array.isArray(courses) && courses.length > 0) {
		console.log(' SelectionsManager: EMERGENCY - First course:', courses[0]);
		console.log(' SelectionsManager: EMERGENCY - First course period_id:', courses[0]?.period_id);
	}
	
	$effect(() => {
		console.log(' SelectionsManager: EFFECT - courses changed:', courses);
		console.log(' SelectionsManager: EFFECT - courses length:', Array.isArray(courses) ? courses.length : 'not array');
	});

	let selections = $state<Selection[]>([]);
	let loading = $state<boolean>(true);
	let error = $state<string>('');

	let searchTerm = $state<string>('');
	let filterPeriod = $state<string>('');
	let filterInvitationType = $state<string>('');
	let filterGrade = $state<string>('');

	let editingSelection = $state<Selection | null>(null);
	let formData = $state({
		student_id: '',
		course_id: '',
		invitation_type: 'no' as const
	});
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

	let periods = $derived.by(() => {
		if (!Array.isArray(courses) || courses.length === 0) {
			return [];
		}
		return [...new Set(courses.map(c => c?.period_id).filter(Boolean))].sort();
	});
	
	let grades = $derived.by(() => {
		if (!Array.isArray(students) || students.length === 0) {
			return [];
		}
		return [...new Set(students.map(s => s?.grade).filter(grade => grade != null))].sort();
	});
	
	let filteredSelections = $derived.by(() => {
		if (!Array.isArray(selections) || selections.length === 0) {
			return [];
		}
		
		return selections.filter(selection => {
			if (!selection) return false;
			
			if (searchTerm) {
				const searchLower = searchTerm.toLowerCase();
				const studentName = selection.student_name?.toLowerCase() || '';
				const courseName = selection.course_name?.toLowerCase() || '';
				const courseId = selection.course_id?.toLowerCase() || '';
				
				if (!studentName.includes(searchLower) && 
					!courseName.includes(searchLower) && 
					!courseId.includes(searchLower)) {
					return false;
				}
			}

			if (filterPeriod && selection.period_id !== filterPeriod) {
				return false;
			}

			if (filterInvitationType && selection.invitation_type !== filterInvitationType) {
				return false;
			}

			if (filterGrade && selection.student_grade?.toString() !== filterGrade) {
				return false;
			}

			return true;
		});
	});


	async function loadSelections(): Promise<void> {
		console.log(' SelectionsManager: loadSelections() called');
		loading = true;
		error = '';
		try {
			console.log(' SelectionsManager: Making API call to get selections');
			const result = await api.admin.getSelections();
			console.log(' SelectionsManager: API call successful', { result, length: Array.isArray(result) ? result.length : 'not array' });
			selections = result;
			console.log(' SelectionsManager: selections state updated', { selections });
		} catch (err) {
			console.error(' SelectionsManager: loadSelections() failed:', err);
			error = handleAPIError(err);
			selections = [];
		} finally {
			loading = false;
			console.log(' SelectionsManager: loadSelections() finished, loading =', loading);
		}
	}

	onMount(() => {
		console.log(' SelectionsManager: onMount() called');
		console.log(' SelectionsManager: Starting initial data load');
		loadSelections().catch(err => {
			console.error(' SelectionsManager: Failed to load selections in onMount:', err);
		});
	});

	function resetForm(): void {
		console.log(' SelectionsManager: resetForm() called');
		formData = {
			student_id: '',
			course_id: '',
			invitation_type: 'no'
		};
		formError = '';
		editingSelection = null;
		studentSearchTerm = '';
		courseSearchTerm = '';
		showStudentList = false;
		showCourseList = false;
		console.log(' SelectionsManager: Form reset completed');
	}

	function startEdit(selection: Selection): void {
		console.log(' SelectionsManager: startEdit() called', { selection });
		editingSelection = selection;
		formData = {
			student_id: selection.student_id.toString(),
			course_id: selection.course_id,
			invitation_type: selection.invitation_type
		};
		formError = '';
		
		const selectedStudent = students.find(s => s.id === selection.student_id);
		const selectedCourse = courses.find(c => c.id === selection.course_id);
		studentSearchTerm = selectedStudent ? `${selectedStudent.id} - ${selectedStudent.name}` : '';
		courseSearchTerm = selectedCourse ? `${selectedCourse.id} - ${selectedCourse.name}` : '';
		showStudentList = false;
		showCourseList = false;
		console.log(' SelectionsManager: Edit form populated');
	}

	async function handleSubmit(): Promise<void> {
		console.log(' SelectionsManager: handleSubmit() called', { formData, editingSelection: !!editingSelection });
		
		if (!formData.student_id || !formData.course_id) {
			console.log(' SelectionsManager: Form validation failed - missing required fields');
			formError = 'Student and course are required';
			return;
		}

		formLoading = true;
		formError = '';

		try {
			const submitData = new FormData();
			console.log(' SelectionsManager: Preparing form data for submission');

			if (editingSelection) {
				console.log(' SelectionsManager: Updating existing selection');
				submitData.append('student_id', formData.student_id);
				submitData.append('old_course_id', editingSelection.course_id);
				submitData.append('new_course_id', formData.course_id);
				
				const course = Array.isArray(courses) ? courses.find(c => c && c.id === formData.course_id) : null;
				console.log(' SelectionsManager: Looking for course', { courseId: formData.course_id, found: !!course });
				if (!course) {
					formError = 'Invalid course selected or courses not loaded';
					console.error(' SelectionsManager: Course not found for update');
					return;
				}
				submitData.append('new_period_id', course.period_id);
				submitData.append('new_invitation_type', formData.invitation_type);

				console.log(' SelectionsManager: Calling API updateSelection');
				await api.admin.updateSelection(submitData);
				console.log(' SelectionsManager: Update selection API call completed');
			} else {
				console.log(' SelectionsManager: Adding new selection');
				submitData.append('student_id', formData.student_id);
				submitData.append('course_id', formData.course_id);
				
				const course = Array.isArray(courses) ? courses.find(c => c && c.id === formData.course_id) : null;
				console.log(' SelectionsManager: Looking for course', { courseId: formData.course_id, found: !!course });
				if (!course) {
					formError = 'Invalid course selected or courses not loaded';
					console.error(' SelectionsManager: Course not found for add');
					return;
				}
				submitData.append('period_id', course.period_id);
				submitData.append('invitation_type', formData.invitation_type);

				console.log(' SelectionsManager: Calling API addSelection');
				await api.admin.addSelection(submitData);
				console.log(' SelectionsManager: Add selection API call completed');
			}

			console.log(' SelectionsManager: Resetting form and reloading data');
			resetForm();
			await loadSelections();
			await onUpdate();
			console.log(' SelectionsManager: Form submission completed successfully');
		} catch (err) {
			console.error(' SelectionsManager: Form submission failed:', err);
			formError = handleAPIError(err);
		} finally {
			formLoading = false;
			console.log(' SelectionsManager: Form submission finished, formLoading =', formLoading);
		}
	}

	let confirmingDelete = $state<Selection | null>(null);

	async function handleDelete(selection: Selection): Promise<void> {
		console.log(' SelectionsManager: handleDelete() called', { selection });

		try {
			console.log(' SelectionsManager: Calling API deleteSelection');
			await api.admin.deleteSelection(selection.student_id, selection.course_id);
			console.log(' SelectionsManager: Delete API call completed');
			confirmingDelete = null;
			await loadSelections();
			await onUpdate();
			console.log(' SelectionsManager: Delete completed and data reloaded');
		} catch (err) {
			console.error(' SelectionsManager: Delete failed:', err);
			error = handleAPIError(err);
		}
	}

	function clearFilters(): void {
		console.log(' SelectionsManager: clearFilters() called');
		searchTerm = '';
		filterPeriod = '';
		filterInvitationType = '';
		filterGrade = '';
		console.log(' SelectionsManager: Filters cleared');
	}

	function selectStudent(student: Student): void {
		formData.student_id = student.id.toString();
		studentSearchTerm = `${student.id} - ${student.name}`;
		showStudentList = false;
	}

	function selectCourse(course: Course): void {
		formData.course_id = course.id;
		courseSearchTerm = `${course.id} - ${course.name}`;
		showCourseList = false;
	}

	function clearStudentSelection(): void {
		formData.student_id = '';
		studentSearchTerm = '';
		showStudentList = false;
	}

	function clearCourseSelection(): void {
		formData.course_id = '';
		courseSearchTerm = '';
		showCourseList = false;
	}
</script>

<div class="selections-manager">
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
				<button onclick={() => {/* TODO: Add bulk export functionality */}} class="btn btn-outline" disabled>
					Export Selections CSV
				</button>
				<button onclick={() => {/* TODO: Add bulk import functionality */}} class="btn btn-outline" disabled>
					Import Selections CSV
				</button>
			</div>
		</div>
	</details>

	<!-- Add Selection Section -->
	<details class="management-section">
		<summary class="management-section-header">
			{editingSelection ? `Edit Selection: ${editingSelection.student_name} â†’ ${editingSelection.course_name}` : 'Add New Selection'}
		</summary>
		<div class="management-section-content">
			{#if formError}
				<div class="alert alert-error">
					{formError}
				</div>
			{/if}

			<form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="selections-form">
				<div class="selections-form-fields">
					<div class="form-group">
						<label class="form-label">Student *</label>
						<div class="searchable-select">
							<input
								type="text"
								bind:value={studentSearchTerm}
								onfocus={() => showStudentList = true}
								oninput={() => showStudentList = true}
								placeholder="Search students by name or ID..."
								required={!formData.student_id}
								disabled={editingSelection !== null || formLoading}
								class="input"
							/>
							{#if formData.student_id}
								<button
									type="button"
									onclick={clearStudentSelection}
									class="clear-selection-btn"
									disabled={formLoading}
								>
									Clear
								</button>
							{/if}
							{#if showStudentList && !editingSelection}
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
								required={!formData.course_id}
								disabled={formLoading}
								class="input"
							/>
							{#if formData.course_id}
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
						<label class="form-label">Invitation Type</label>
						<select bind:value={formData.invitation_type} disabled={formLoading} class="select">
							<option value="no">Normal</option>
							<option value="invite">Invited</option>
							<option value="force">Forced</option>
						</select>
					</div>
				</div>

				<div class="form-actions">
					<button type="button" onclick={resetForm} disabled={formLoading} class="btn btn-secondary">
						{editingSelection ? 'Cancel Edit' : 'Clear'}
					</button>
					<button type="submit" disabled={formLoading} class="btn btn-primary">
						{formLoading ? 'Saving...' : (editingSelection ? 'Update' : 'Add')} Selection
					</button>
				</div>
			</form>
		</div>
	</details>

	<!-- Search and Controls -->
	<div class="selections-controls">
		<!-- Search and filters -->
		<div class="selections-filters">
			<div class="form-group">
				<label class="form-label">Search</label>
				<input
					type="text"
					bind:value={searchTerm}
					placeholder="Student or course name..."
					class="input"
				/>
			</div>

			<div class="form-group">
				<label class="form-label">Period</label>
				<select bind:value={filterPeriod} class="select">
					<option value="">All periods</option>
					{#each (periods || []) as period}
						<option value={period}>{period}</option>
					{/each}
				</select>
			</div>

			<div class="form-group">
				<label class="form-label">Type</label>
				<select bind:value={filterInvitationType} class="select">
					<option value="">All types</option>
					<option value="no">Normal</option>
					<option value="invite">Invited</option>
					<option value="force">Forced</option>
				</select>
			</div>

			<div class="form-group">
				<label class="form-label">Grade</label>
				<select bind:value={filterGrade} class="select">
					<option value="">All grades</option>
					{#each (grades || []) as grade}
						<option value={grade.toString()}>{grade}</option>
					{/each}
				</select>
			</div>

			<div class="selections-clear-filters">
				<button onclick={clearFilters} class="btn btn-secondary btn-sm">
					Clear Filters
				</button>
			</div>
		</div>

		<!-- Results summary -->
		<div class="selections-results-summary">
			Showing {(filteredSelections || []).length} of {(selections || []).length} selections
		</div>
	</div>

	<!-- Selections List -->
	{#if loading}
		<div class="selections-loading">
			<div class="selections-loading-text">Loading selections...</div>
		</div>
	{:else if (filteredSelections || []).length === 0}
		<div class="selections-empty-state">
			<div class="selections-empty-message">
				{(selections || []).length === 0 ? 'No selections found' : 'No selections match the current filters'}
			</div>
		</div>
	{:else}
		<div class="selections-table-container">
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
					{#each (filteredSelections || []) as selection}
						<tr class="table-row">
							<td class="table-cell">
								<div class="selections-student-info">
									<div class="selections-student-name">{selection.student_name}</div>
									<div class="selections-student-details">ID: {selection.student_id}, Grade {selection.student_grade}</div>
								</div>
							</td>
							<td class="table-cell">
								<div class="selections-course-info">
									<div class="selections-course-name">{selection.course_name}</div>
									<div class="selections-course-id">{selection.course_id}</div>
								</div>
							</td>
							<td class="table-cell">
								<div class="selections-period">{selection.period_id}</div>
							</td>
							<td class="table-cell">
								<span class="badge {
									selection.invitation_type === 'force' ? 'badge-danger' :
									selection.invitation_type === 'invite' ? 'badge-warning' :
									'badge-success'
								}">
									{selection.invitation_type === 'no' ? 'Normal' :
									 selection.invitation_type === 'invite' ? 'Invited' : 'Forced'}
								</span>
							</td>
							<td class="table-cell selections-table-actions">
								{#if confirmingDelete === selection}
									<div class="selections-delete-confirmation">
										<span class="selections-delete-prompt">Delete selection?</span>
										<button onclick={() => handleDelete(selection)} class="selections-delete-confirm-btn">
											Confirm
										</button>
										<button onclick={() => confirmingDelete = null} class="selections-delete-cancel-btn">
											Cancel
										</button>
									</div>
								{:else}
									<button onclick={() => startEdit(selection)} class="selections-edit-btn">
										Edit
									</button>
									<button onclick={() => confirmingDelete = selection} class="selections-delete-btn">
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
