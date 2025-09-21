<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import ChangePasswordForm from './ChangePasswordForm.svelte';
	import { getSSEClient } from './sse.ts';

	interface AuthState {
		authenticated: boolean;
		role?: string;
		username?: string;
		user_id?: number;
	}

	interface Props {
		authState: AuthState;
		onLogout: () => void;
	}

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
		current_enrollment: number;
		student_invitation_type: 'no' | 'invite' | 'force';
		availability_status:
			| 'selected'
			| 'available'
			| 'at_capacity'
			| 'sex_restriction'
			| 'grade_restriction'
			| 'selection_closed'
			| 'invite_only';
	}

	interface Selection {
		course_id: string;
		course_name: string;
		description: string;
		period_id: string;
		teacher: string;
		location: string;
		category_id: string;
		invitation_type: 'no' | 'invite' | 'force';
	}

	interface SelectionByPeriod {
		period_id: string;
		course_id?: string;
		course_name?: string;
		teacher?: string;
		location?: string;
		category_id?: string;
		invitation_type?: 'no' | 'invite' | 'force';
	}

	interface RequirementsStatus {
		total_requirements: {
			grade: number;
			required_total: number;
			current_total: number;
		};
		group_requirements: {
			id: number;
			label: string;
			required_count: number;
			current_count: number;
			required_categories: string[];
		}[];
	}

	const { authState, onLogout }: Props = $props();

	let showChangePassword = $state(false);
	let activeTab = $state<'select' | 'review'>('select');
	let viewMode = $state<'cards' | 'table'>('cards');

	let courses: Course[] = $state([]);
	let filteredCourses: Course[] = $state([]);
	let selectionsByPeriod: SelectionByPeriod[] = $state([]);
	let requirementsStatus: RequirementsStatus | null = $state(null);
	let loading = $state(true);
	let error = $state('');
	let submitting = $state(false);
	let searchQuery = $state('');
	let selectionEnabled = $state(true);
	let pendingDeselection = $state<{ course: Course; message: string } | null>(null);

	let coursesByPeriod: Record<string, Course[]> = $state({});

	let tableGroups: { period: string; courses: Course[] }[] = $state([]);

	let initialLoaded = $state(false);

	function naturalCompare(a: string, b: string): number {
		return a.localeCompare(b, undefined, { numeric: true, sensitivity: 'base' });
	}

	$effect(() => {
		const q = searchQuery.toLowerCase();
		const filtered = courses.filter((course) => {
			if (!q) return true;
			return (
				course.name.toLowerCase().includes(q) ||
				course.teacher.toLowerCase().includes(q) ||
				course.location.toLowerCase().includes(q) ||
				course.category_id.toLowerCase().includes(q) ||
				(course.description && course.description.toLowerCase().includes(q))
			);
		});

		filteredCourses = filtered;

		const periodMap: Record<string, Course[]> = {};
		for (const course of filtered) {
			(periodMap[course.period_id] ??= []).push(course);
		}

		coursesByPeriod = periodMap;

		const periods = Object.keys(periodMap).sort(naturalCompare);
		tableGroups = periods.map((period) => {
			const list = [...(periodMap[period] ?? [])];
			list.sort((a, b) => naturalCompare(a.name, b.name));
			return { period, courses: list };
		});
	});

	function handleLogout(): void {
		onLogout();
	}

	async function loadCourses(): Promise<void> {
		try {
			const response = await fetch('/api/student/courses');
			if (!response.ok) throw new Error('Failed to load courses');
			courses = await response.json();
		} catch (err) {
			console.error('Failed to load courses:', err);
			error = 'Failed to load courses';
		}
	}

	async function loadSelectionsByPeriod(): Promise<void> {
		try {
			const response = await fetch('/api/student/selections/by-period');
			if (!response.ok) throw new Error('Failed to load selections');
			selectionsByPeriod = await response.json();
		} catch (err) {
			console.error('Failed to load selections:', err);
			error = 'Failed to load selections';
		}
	}

	async function loadRequirements(): Promise<void> {
		try {
			const response = await fetch('/api/student/requirements');
			if (!response.ok) throw new Error('Failed to load requirements');
			requirementsStatus = await response.json();
		} catch (err) {
			console.error('Failed to load requirements:', err);
			error = 'Failed to load requirements';
		}
	}

	async function loadSelectionStatus(): Promise<void> {
		try {
			const response = await fetch('/api/student/selection-status');
			if (!response.ok) throw new Error('Failed to load selection status');
			const data = await response.json();
			selectionEnabled = data.enabled;
		} catch (err) {
			console.error('Failed to load selection status:', err);
			selectionEnabled = true;
		}
	}

	async function loadData(opts: { showSpinner?: boolean } = {}): Promise<void> {
		const { showSpinner = false } = opts;
		if (showSpinner) loading = true;
		error = '';
		try {
			await Promise.all([loadCourses(), loadSelectionsByPeriod(), loadRequirements(), loadSelectionStatus()]);
		} finally {
			if (showSpinner) loading = false;
			initialLoaded = true;
		}
	}

	function getAvailabilityText(course: Course): string {
		switch (course.availability_status) {
			case 'selected':
				return course.student_invitation_type === 'force'
					? 'Forced'
					: course.student_invitation_type === 'invite'
						? 'Invited & Selected'
						: 'Selected';
			case 'available':
				return course.student_invitation_type === 'invite' ? 'Invited' : 'Available';
			case 'at_capacity':
				return 'At Capacity';
			case 'sex_restriction':
				return 'Sex Restriction';
			case 'grade_restriction':
				return 'Grade Restriction';
			case 'invite_only':
				return 'Invite Only';
			default:
				return 'Unavailable';
		}
	}

	function getAvailabilityColor(course: Course): string {
		switch (course.availability_status) {
			case 'selected':
				return course.student_invitation_type === 'force'
					? 'bg-red-100 text-red-800'
					: 'bg-green-100 text-green-800';
			case 'available':
				return course.student_invitation_type === 'invite'
					? 'bg-blue-100 text-blue-800'
					: 'bg-gray-100 text-gray-800';
			case 'at_capacity':
			case 'sex_restriction':
			case 'grade_restriction':
			case 'invite_only':
				return 'bg-red-100 text-red-800';
			default:
				return 'bg-gray-100 text-gray-600';
		}
	}

	function canSelect(course: Course): boolean {
		return (
			course.availability_status === 'available' ||
			(course.availability_status === 'selected' && course.student_invitation_type === 'invite')
		);
	}

	function canDeselect(course: Course): boolean {
		return course.availability_status === 'selected' && course.student_invitation_type !== 'force';
	}

	async function selectCourse(course: Course): Promise<void> {
		if (submitting) return;

		submitting = true;
		try {
			const formData = new FormData();
			formData.append('course_id', course.id);

			const response = await fetch('/api/student/select-course', {
				method: 'POST',
				body: formData
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText || 'Failed to select course');
			}

			await refreshQuietly();
		} catch (err) {
			console.error('Failed to select course:', err);
			error = err instanceof Error ? err.message : 'Failed to select course';
		} finally {
			submitting = false;
		}
	}

	async function deselectCourse(course: Course): Promise<void> {
		if (submitting) return;

		const confirmMessage =
			course.student_invitation_type === 'invite'
				? 'You are about to decline an invitation. You will irrevocably lose your spot if this course is invite-only. Are you sure?'
				: 'You will irrevocably lose your spot if this course becomes full. Are you sure?';

		pendingDeselection = { course, message: confirmMessage };
	}

	async function confirmDeselectCourse(): Promise<void> {
		if (!pendingDeselection || submitting) return;

		submitting = true;
		try {
			const response = await fetch(
				`/api/student/select-course?course_id=${encodeURIComponent(pendingDeselection.course.id)}`,
				{ method: 'DELETE' }
			);

			if (!response.ok) throw new Error('Failed to deselect course');

			await refreshQuietly();
		} catch (err) {
			console.error('Failed to deselect course:', err);
			error = 'Failed to deselect course';
		} finally {
			submitting = false;
			pendingDeselection = null;
		}
	}

	function checkRequirementsMet(): { met: boolean; issues: string[] } {
		if (!requirementsStatus) return { met: true, issues: [] };

		const issues: string[] = [];

		const totalReq = requirementsStatus.total_requirements;
		if (totalReq && totalReq.current_total < totalReq.required_total) {
			issues.push(`Need ${totalReq.required_total - totalReq.current_total} more courses total`);
		}

		if (requirementsStatus.group_requirements) {
			requirementsStatus.group_requirements.forEach((group) => {
				if (group.current_count < group.required_count) {
					const needed = group.required_count - group.current_count;
					issues.push(`Need ${needed} more from ${group.label} (${group.required_categories.join(', ')})`);
				}
			});
		}

		return { met: issues.length === 0, issues };
	}

	async function refreshQuietly(): Promise<void> {
		const scroller = document.scrollingElement || document.documentElement;
		const { scrollTop, scrollLeft } = scroller;
		await loadData();
		scroller.scrollTo(scrollLeft, scrollTop);
	}

	let sseCleanup: (() => void) | null = null;

	let sseRefreshQueued = false;
	const scheduleCoursesRefresh = () => {
		if (sseRefreshQueued) return;
		sseRefreshQueued = true;
		requestAnimationFrame(async () => {
			try {
				await loadCourses();
			} finally {
				sseRefreshQueued = false;
			}
		});
	};

	onMount(async () => {
		await loadData({ showSpinner: true });

		const sseClient = getSSEClient();
		if (sseClient) {
			const handleCourseUpdate = () => {
				scheduleCoursesRefresh();
			};

			const handleDataInvalidation = () => {
				refreshQuietly();
			};

			const handleGradeSelectionStatusChange = (_message: any) => {
				if (authState.user_id) {
					loadSelectionStatus();
				}
			};

			sseClient.on('course_enrollment_update', handleCourseUpdate);
			sseClient.on('invalidate_courses', handleDataInvalidation);
			sseClient.on('invalidate_requirements', handleDataInvalidation);
			sseClient.on('grade_selection_status_change', handleGradeSelectionStatusChange);

			sseCleanup = () => {
				sseClient.off('course_enrollment_update', handleCourseUpdate);
				sseClient.off('invalidate_courses', handleDataInvalidation);
				sseClient.off('invalidate_requirements', handleDataInvalidation);
				sseClient.off('grade_selection_status_change', handleGradeSelectionStatusChange);
			};
		}
	});

	onDestroy(() => {
		if (sseCleanup) sseCleanup();
	});
</script>

<div class="student-dashboard">
	<header class="student-header">
		<div class="student-header-content">
			<div class="student-header-layout">
				<h1 class="student-header-title">Student Portal</h1>
				<div class="student-header-actions">
					<span class="student-header-user-info">Student ID: {authState.user_id}</span>
					<button
						onclick={() => showChangePassword = !showChangePassword}
						class="student-header-button student-header-button--secondary"
					>
						Change Password
					</button>
					<button onclick={() => handleLogout()} class="student-header-button student-header-button--primary">
						Logout
					</button>
				</div>
			</div>
		</div>
	</header>

	<main class="student-main-content">
		{#if showChangePassword}
			<div class="student-change-password-section">
				<ChangePasswordForm onClose={() => (showChangePassword = false)} />
			</div>
		{/if}

		{#if error}
			<div class="student-error-message">
				{error}
			</div>
		{/if}

		<div class="student-tabs-container">
			<div class="student-tabs-header">
				<nav class="student-tab-navigation">
					<button
						onclick={() => (activeTab = 'select')}
						class="student-tab-button {activeTab === 'select'
							? 'student-tab-button--active'
							: 'student-tab-button--inactive'}"
					>
						Select
					</button>
					<button
						onclick={() => (activeTab = 'review')}
						class="student-tab-button {activeTab === 'review'
							? 'student-tab-button--active'
							: 'student-tab-button--inactive'}"
					>
						Review
					</button>
				</nav>
			</div>

			{#if !initialLoaded && loading}
				<div class="student-loading-container">Loading...</div>
			{:else if activeTab === 'select'}
				<div class="student-select-tab-content">
					<div class="student-select-controls">
						<div class="student-search-container">
							<input
								type="text"
								placeholder="Search courses, teachers, locations..."
								bind:value={searchQuery}
								class="student-search-input"
							/>
						</div>

						<div class="student-view-toggle">
							<button
								onclick={() => (viewMode = 'cards')}
								class="student-view-toggle-button {viewMode === 'cards'
									? 'student-view-toggle-button--active'
									: 'student-view-toggle-button--inactive'}"
								aria-pressed={viewMode === 'cards'}
							>
								Cards
							</button>
							<button
								onclick={() => (viewMode = 'table')}
								class="student-view-toggle-button {viewMode === 'table'
									? 'student-view-toggle-button--active'
									: 'student-view-toggle-button--inactive'}"
								aria-pressed={viewMode === 'table'}
							>
								Table
							</button>
						</div>
					</div>

					{#if !selectionEnabled}
						<div class="student-selection-closed-notice">
							<div class="student-selection-closed-icon">
								<svg class="student-selection-closed-icon-svg" viewBox="0 0 20 20" fill="currentColor">
									<path
										fill-rule="evenodd"
										d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
										clip-rule="evenodd"
									/>
								</svg>
							</div>
							<div class="student-selection-closed-content">
								<h3 class="student-selection-closed-title">Course Selection Closed</h3>
								<div class="student-selection-closed-message">
									Course selection is currently closed for your grade. You can view available courses but cannot make changes to your selections.
								</div>
							</div>
						</div>
					{/if}

					{#if viewMode === 'table'}
						{#if filteredCourses.length === 0}
							{#if searchQuery}
								<div class="student-empty-state">
									<p class="student-empty-message">No courses found matching "{searchQuery}"</p>
									<button
										onclick={() => (searchQuery = '')}
										class="student-clear-search-button"
									>
										Clear search
									</button>
								</div>
							{:else}
								<p class="student-empty-message">No courses available.</p>
							{/if}
						{:else}
							<div class="student-courses-table-container">
								<table class="student-courses-table">
									<thead class="student-courses-table-header">
										<tr>
											<th class="student-courses-table-header-cell">
												Category
											</th>
											<th class="student-courses-table-header-cell">
												Title
											</th>
											<th class="student-courses-table-header-cell student-courses-table-header-cell--right">
												Seats
											</th>
											<th class="student-courses-table-header-cell student-courses-table-header-cell--right">
												Action
											</th>
										</tr>
									</thead>
									<tbody class="student-courses-table-body">
										{#each tableGroups as group (group.period)}
											<tr class="student-courses-table-period-header">
												<td class="student-courses-table-period-cell" colspan="4">
													{group.period}
												</td>
											</tr>

											{#each group.courses as course (course.id)}
												<tr class="student-courses-table-row {course.availability_status === 'selected' ? 'student-courses-table-row--selected' : ''}">
													<td class="student-courses-table-cell">{course.category_id}</td>
													<td class="student-courses-table-cell student-courses-table-cell--title">{course.name}</td>
													<td class="student-courses-table-cell student-courses-table-cell--right">
														{course.current_enrollment}/{course.max_students}
													</td>
													<td class="student-courses-table-cell student-courses-table-cell--right">
														{#if course.availability_status === 'selected'}
															{#if pendingDeselection?.course === course}
																<div class="student-table-deselection-prompt">{pendingDeselection.message}</div>
																<div class="student-table-deselection-actions">
																	<button
																		onclick={confirmDeselectCourse}
																		disabled={submitting}
																		class="student-table-action-button student-table-action-button--danger"
																	>
																		{submitting ? 'Removing...' : 'Confirm Remove'}
																	</button>
																	<button
																		onclick={() => (pendingDeselection = null)}
																		disabled={submitting}
																		class="student-table-action-button student-table-action-button--secondary"
																	>
																		Cancel
																	</button>
																</div>
															{:else if canDeselect(course) && selectionEnabled}
																<button
																	onclick={() => deselectCourse(course)}
																	disabled={submitting}
																	class="student-table-action-button student-table-action-button--danger"
																>
																	{submitting ? 'Removing...' : 'Remove'}
																</button>
															{:else if !selectionEnabled}
																<span class="student-table-disabled-button">
																	Selections Closed
																</span>
															{:else}
																<span class="student-table-disabled-button">
																	Cannot deselect
																</span>
															{/if}
														{:else if canSelect(course) && selectionEnabled}
															<button
																onclick={() => selectCourse(course)}
																disabled={submitting}
																class="student-table-action-button student-table-action-button--select"
															>
																{submitting ? 'Selecting...' : 'Select'}
															</button>
														{:else if !selectionEnabled}
															<span class="student-table-disabled-button">
																Selections Closed
															</span>
														{:else}
															<button
																disabled
																class="student-table-disabled-button"
															>
																Unavailable
															</button>
														{/if}
													</td>
												</tr>
											{/each}
										{/each}
									</tbody>
								</table>
							</div>
						{/if}
					{:else}
						{#if Object.keys(coursesByPeriod).length === 0}
							{#if searchQuery}
								<div class="student-empty-state">
									<p class="student-empty-message">No courses found matching "{searchQuery}"</p>
									<button
										onclick={() => (searchQuery = '')}
										class="student-clear-search-button"
									>
										Clear search
									</button>
								</div>
							{:else}
								<p class="student-empty-message">No courses available.</p>
							{/if}
						{:else}
							{#each Object.entries(coursesByPeriod) as [period, periodCourses] (period)}
								<div class="student-period-section">
									<h3 class="student-period-title">{period}</h3>
									<div class="student-courses-grid">
										{#each periodCourses as course (course.id)}
											<div class="student-course-card {course.availability_status === 'selected' ? 'student-course-card--selected' : ''}">
												<div class="student-course-card-content">
													<div class="student-course-card-header">
														<div class="student-course-card-category-bar">
															{course.category_id}
															<div class="student-course-card-badges">
																<span class="student-course-card-enrollment-badge">
																	{course.current_enrollment}/{course.max_students}
																</span>
																<span class="student-course-card-status-badge {getAvailabilityColor(course)}">
																	{getAvailabilityText(course)}
																</span>
															</div>
														</div>

														<h4 class="student-course-card-title">{course.name}</h4>

														<div class="student-course-card-teacher-location">
															<span class="student-course-card-teacher">{course.teacher}</span>
															<span class="student-course-card-location">{course.location}</span>
														</div>
													</div>

													{#if course.description}
														<p class="student-course-card-description">{course.description}</p>
													{/if}
												</div>

												<div class="student-course-card-spacer"></div>

												<div class="student-course-card-actions">
													{#if course.availability_status === 'selected'}
														{#if pendingDeselection?.course === course}
															<div class="student-card-deselection-container">
																<div class="student-card-deselection-prompt">
																	{pendingDeselection.message}
																</div>
																<div class="student-card-deselection-actions">
																	<button
																		onclick={confirmDeselectCourse}
																		disabled={submitting}
																		class="student-card-action-button student-card-action-button--danger"
																	>
																		{submitting ? 'Removing...' : 'Confirm Remove'}
																	</button>
																	<button
																		onclick={() => pendingDeselection = null}
																		disabled={submitting}
																		class="student-card-action-button student-card-action-button--secondary"
																	>
																		Cancel
																	</button>
																</div>
															</div>
														{:else if canDeselect(course) && selectionEnabled}
															<button
																onclick={() => deselectCourse(course)}
																disabled={submitting}
																class="student-card-action-button student-card-action-button--danger student-card-action-button--full"
															>
																{submitting ? 'Removing...' : 'Remove'}
															</button>
														{:else if !selectionEnabled}
															<span class="student-card-disabled-button">
																Selections Closed
															</span>
														{:else}
															<span class="student-card-disabled-button">
																Cannot deselect
															</span>
														{/if}
													{:else if canSelect(course) && selectionEnabled}
														<button
															onclick={() => selectCourse(course)}
															disabled={submitting}
															class="student-card-action-button student-card-action-button--select student-card-action-button--full"
														>
															{submitting ? 'Selecting...' : 'Select'}
														</button>
													{:else if !selectionEnabled}
														<span class="student-card-disabled-button">
															Selections Closed
														</span>
													{:else}
														<button
															disabled
															class="student-card-disabled-button"
														>
															Unavailable
														</button>
													{/if}
												</div>
											</div>
										{/each}
									</div>
								</div>
							{/each}
						{/if}
					{/if}
				</div>
			{:else if activeTab === 'review'}
				<div class="student-review-tab-content">
					{#if requirementsStatus}
						{@const reqCheck = checkRequirementsMet()}
						<div class="student-requirements-status {reqCheck.met ? 'student-requirements-status--met' : 'student-requirements-status--unmet'}">
							<h3 class="student-requirements-status-title">
								Requirements Status: {reqCheck.met ? 'Met' : 'Not Met'}
							</h3>

							{#if !reqCheck.met}
								<ul class="student-requirements-issues-list">
									{#each reqCheck.issues as issue}
										<li class="student-requirements-issue">{issue}</li>
									{/each}
								</ul>
							{:else}
								<p class="student-requirements-satisfied">All requirements satisfied!</p>
							{/if}
						</div>
					{/if}

					<div class="student-selections-list">
						{#each selectionsByPeriod as selection (selection.period_id)}
							<div class="student-selection-item {selection.course_id ? 'student-selection-item--filled' : 'student-selection-item--empty'}">
								<div class="student-selection-item-content">
									<div class="student-selection-item-header">
										<h4 class="student-selection-period-title">{selection.period_id}</h4>
										{#if selection.course_id && selection.invitation_type}
											<span class="student-selection-invitation-badge {selection.invitation_type === 'force'
													? 'student-selection-invitation-badge--forced'
													: selection.invitation_type === 'invite'
														? 'student-selection-invitation-badge--invited'
														: 'student-selection-invitation-badge--selected'}">
												{selection.invitation_type === 'force'
													? 'Forced'
													: selection.invitation_type === 'invite'
														? 'Invited'
														: 'Selected'}
											</span>
										{/if}
									</div>
									{#if selection.course_id}
										<div class="student-selection-course-details">
											<h5 class="student-selection-course-name">{selection.course_name}</h5>
											<div class="student-selection-course-info">
												<div class="student-selection-info-item"><strong>Teacher:</strong> {selection.teacher}</div>
												<div class="student-selection-info-item"><strong>Location:</strong> {selection.location}</div>
												<div class="student-selection-info-item"><strong>Category:</strong> {selection.category_id}</div>
											</div>
										</div>
									{:else}
										<p class="student-selection-flex-time">Flex time</p>
									{/if}
								</div>
							</div>
						{/each}
					</div>
				</div>
			{/if}
		</div>
	</main>
</div>
