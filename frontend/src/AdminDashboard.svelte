<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import ChangePasswordForm from './ChangePasswordForm.svelte';
	import CatalogManager from './CatalogManager.svelte';
	import CourseManager from './CourseManager.svelte';
	import RequirementsManager from './RequirementsManager.svelte';
	import StudentsManager from './StudentsManager.svelte';
	import InvitationsManager from './InvitationsManager.svelte';
	import SelectionControlsManager from './SelectionControlsManager.svelte';
	import SelectionsManager from './SelectionsManager.svelte';
	import { initializeSSE, getSSEClient, disconnectSSE, type SSEMessage } from './sse';
	import type { AuthState, Course, Student, Invitation, GradeSelectionControl, Selection } from './types.ts';
	import { api, handleAPIError } from './api.ts';

	interface Props {
		authState: AuthState;
		onLogout: () => void;
	}

	const { authState, onLogout }: Props = $props();

	let showChangePassword = $state<boolean>(false);
	let activeTab = $state<string>('catalog');
	let loading = $state<boolean>(false);
	let error = $state<string>('');
	let connectionStatus = $state<boolean>(false);
	let selectionsLoading = $state<boolean>(false);

	let grades = $state<number[]>([]);
	let periods = $state<string[]>([]);
	let categories = $state<string[]>([]);
	let courses = $state<Course[]>([]);
	let students = $state<Student[]>([]);
	let invitations = $state<Invitation[]>([]);
	let selectionControls = $state<GradeSelectionControl[]>([]);
	let selections = $state<Selection[]>([]);

	function handleSSEConnection(message: SSEMessage): void {
		console.log(' AdminDashboard: SSE connected:', message.data);
		connectionStatus = true;
	}

	function handleGradesInvalidation(_message: SSEMessage): void {
		console.log(' AdminDashboard: Grades invalidated, refetching...');
		loadGrades().catch((err) => {
			console.error(' AdminDashboard: Failed to reload grades:', err);
		});
	}

	function handlePeriodsInvalidation(_message: SSEMessage): void {
		console.log(' AdminDashboard: Periods invalidated, refetching...');
		loadPeriods().catch((err) => {
			console.error(' AdminDashboard: Failed to reload periods:', err);
		});
	}

	function handleCategoriesInvalidation(_message: SSEMessage): void {
		console.log(' AdminDashboard: Categories invalidated, refetching...');
		loadCategories().catch((err) => {
			console.error(' AdminDashboard: Failed to reload categories:', err);
		});
	}

	function handleCoursesInvalidation(_message: SSEMessage): void {
		console.log(' AdminDashboard: Courses invalidated, refetching...');
		loadCourses().catch((err) => {
			console.error(' AdminDashboard: Failed to reload courses:', err);
		});
	}

	function handleRequirementsInvalidation(_message: SSEMessage): void {
		console.log(' AdminDashboard: Requirements invalidated, refetching...');
	}

	function handleStudentsInvalidation(_message: SSEMessage): void {
		console.log(' AdminDashboard: Students invalidated, refetching...');
		loadStudents().catch((err) => {
			console.error(' AdminDashboard: Failed to reload students:', err);
		});
	}

	function handleInvitationsInvalidation(_message: SSEMessage): void {
		console.log(' AdminDashboard: Invitations invalidated, refetching...');
		loadInvitations().catch((err) => {
			console.error(' AdminDashboard: Failed to reload invitations:', err);
		});
	}

	function handleSelectionControlsInvalidation(_message: SSEMessage): void {
		console.log(' AdminDashboard: Selection controls invalidated, refetching...');
		loadSelectionControls().catch((err) => {
			console.error(' AdminDashboard: Failed to reload selection controls:', err);
		});
	}

	function handleSelectionsInvalidation(_message: SSEMessage): void {
		console.log(' AdminDashboard: Selections invalidated, refetching...');
		loadSelections().catch((err) => {
			console.error(' AdminDashboard: Failed to reload selections:', err);
		});
	}

	onMount(() => {
		console.log(' AdminDashboard: onMount() called');
		console.log(' AdminDashboard: Initializing SSE connection');
		initializeSSE();
		
		const client = getSSEClient();
		console.log(' AdminDashboard: SSE client', { client: !!client });
		if (client) {
			console.log(' AdminDashboard: Registering SSE event handlers');
			client.on('connected', handleSSEConnection);
			client.on('invalidate_grades', handleGradesInvalidation);
			client.on('invalidate_periods', handlePeriodsInvalidation);
			client.on('invalidate_categories', handleCategoriesInvalidation);
			client.on('invalidate_courses', handleCoursesInvalidation);
			client.on('invalidate_requirements', handleRequirementsInvalidation);
			client.on('invalidate_students', handleStudentsInvalidation);
			client.on('invalidate_invitations', handleInvitationsInvalidation);
			client.on('invalidate_selection_controls', handleSelectionControlsInvalidation);
			client.on('invalidate_selections', handleSelectionsInvalidation);
			
			const statusInterval = setInterval(() => {
				const status = client.getConnectionStatus();
				if (connectionStatus !== status) {
					console.log(' AdminDashboard: Connection status changed', { old: connectionStatus, new: status });
					connectionStatus = status;
				}
			}, 1000);
			
			onDestroy(() => {
				console.log(' AdminDashboard: Cleaning up status interval');
				clearInterval(statusInterval);
			});
		}

		console.log(' AdminDashboard: Loading initial catalog data');
		loadCatalogData().catch((err) => {
			console.error(' AdminDashboard: Failed to load initial catalog data:', err);
		});
	});

	onDestroy(() => {
		const client = getSSEClient();
		if (client) {
			client.off('connected', handleSSEConnection);
			client.off('invalidate_grades', handleGradesInvalidation);
			client.off('invalidate_periods', handlePeriodsInvalidation);
			client.off('invalidate_categories', handleCategoriesInvalidation);
			client.off('invalidate_courses', handleCoursesInvalidation);
			client.off('invalidate_requirements', handleRequirementsInvalidation);
			client.off('invalidate_students', handleStudentsInvalidation);
			client.off('invalidate_invitations', handleInvitationsInvalidation);
			client.off('invalidate_selection_controls', handleSelectionControlsInvalidation);
			client.off('invalidate_selections', handleSelectionsInvalidation);
		}
		disconnectSSE();
	});

	function handleLogout(): void {
		disconnectSSE();
		onLogout();
	}

	async function loadGrades(): Promise<void> {
		console.log(' AdminDashboard: loadGrades() called');
		try {
			const result = await api.admin.getGrades();
			console.log(' AdminDashboard: loadGrades() success', { result, isArray: Array.isArray(result) });
			grades = Array.isArray(result) ? result : [];
		} catch (err) {
			console.error(' AdminDashboard: loadGrades() failed:', handleAPIError(err));
			grades = [];
		}
	}

	async function loadPeriods(): Promise<void> {
		console.log(' AdminDashboard: loadPeriods() called');
		try {
			const result = await api.admin.getPeriods();
			console.log(' AdminDashboard: loadPeriods() success', { result, isArray: Array.isArray(result) });
			periods = Array.isArray(result) ? result : [];
		} catch (err) {
			console.error(' AdminDashboard: loadPeriods() failed:', handleAPIError(err));
			periods = [];
		}
	}

	async function loadCategories(): Promise<void> {
		console.log(' AdminDashboard: loadCategories() called');
		try {
			const result = await api.admin.getCategories();
			console.log(' AdminDashboard: loadCategories() success', { result, isArray: Array.isArray(result) });
			categories = Array.isArray(result) ? result : [];
		} catch (err) {
			console.error(' AdminDashboard: loadCategories() failed:', handleAPIError(err));
			categories = [];
		}
	}

	async function loadCatalogData(): Promise<void> {
		console.log(' AdminDashboard: loadCatalogData() called');
		loading = true;
		error = '';
		
		try {
			console.log(' AdminDashboard: Starting parallel catalog data load');
			await Promise.all([
				loadGrades(),
				loadPeriods(),
				loadCategories()
			]);
			console.log(' AdminDashboard: All catalog data loaded successfully');
		} catch (err) {
			console.error(' AdminDashboard: loadCatalogData() failed:', handleAPIError(err));
			error = handleAPIError(err);
		} finally {
			loading = false;
			console.log(' AdminDashboard: loadCatalogData() finished, loading =', loading);
		}
	}

	async function loadCourses(): Promise<void> {
		console.log(' AdminDashboard: loadCourses() called');
		try {
			const result = await api.admin.getCourses();
			console.log(' AdminDashboard: loadCourses() success', { result, length: Array.isArray(result) ? result.length : 'not array' });
			courses = Array.isArray(result) ? result : [];
		} catch (err) {
			console.error(' AdminDashboard: loadCourses() failed:', handleAPIError(err));
			courses = [];
		}
	}

	async function loadStudents(): Promise<void> {
		console.log(' AdminDashboard: loadStudents() called');
		try {
			const result = await api.admin.getStudents();
			console.log(' AdminDashboard: loadStudents() success', { result, length: Array.isArray(result) ? result.length : 'not array' });
			students = Array.isArray(result) ? result : [];
		} catch (err) {
			console.error(' AdminDashboard: loadStudents() failed:', handleAPIError(err));
			students = [];
		}
	}

	async function loadInvitations(): Promise<void> {
		console.log(' AdminDashboard: loadInvitations() called');
		try {
			const result = await api.admin.getInvitations();
			console.log(' AdminDashboard: loadInvitations() success', { result, length: Array.isArray(result) ? result.length : 'not array' });
			invitations = Array.isArray(result) ? result : [];
		} catch (err) {
			console.error(' AdminDashboard: loadInvitations() failed:', handleAPIError(err));
			invitations = [];
		}
	}

	async function loadSelectionControls(): Promise<void> {
		console.log(' AdminDashboard: loadSelectionControls() called');
		try {
			const result = await api.admin.getSelectionControls();
			console.log(' AdminDashboard: loadSelectionControls() success', { result, length: Array.isArray(result) ? result.length : 'not array' });
			selectionControls = Array.isArray(result) ? result : [];
		} catch (err) {
			console.error(' AdminDashboard: loadSelectionControls() failed:', handleAPIError(err));
			selectionControls = [];
		}
	}

	async function loadSelections(): Promise<void> {
		console.log(' AdminDashboard: loadSelections() called');
		try {
			const result = await api.admin.getSelections();
			console.log(' AdminDashboard: loadSelections() success', { result, length: Array.isArray(result) ? result.length : 'not array' });
			selections = Array.isArray(result) ? result : [];
		} catch (err) {
			console.error(' AdminDashboard: loadSelections() failed:', handleAPIError(err));
			selections = [];
		}
	}

	function switchTab(tab: string): void {
		console.log(' AdminDashboard: switchTab() called', { from: activeTab, to: tab });
		activeTab = tab;
		error = '';
		selectionsLoading = false;
		
		console.log(' AdminDashboard: Tab switched, starting data loading for:', tab);
		
		if (tab === 'courses') {
			console.log(' AdminDashboard: Loading courses data for courses tab');
			loadCourses().catch((err) => {
				console.error(' AdminDashboard: Failed to load courses on tab switch:', handleAPIError(err));
			});
		} else if (tab === 'students') {
			console.log(' AdminDashboard: Loading students data for students tab');
			loadStudents().catch((err) => {
				console.error(' AdminDashboard: Failed to load students on tab switch:', handleAPIError(err));
			});
		} else if (tab === 'invitations') {
			console.log(' AdminDashboard: Loading invitations data for invitations tab');
			loadInvitations().catch((err) => {
				console.error(' AdminDashboard: Failed to load invitations on tab switch:', handleAPIError(err));
			});
		} else if (tab === 'selection-controls') {
			console.log(' AdminDashboard: Loading selection controls data for selection-controls tab');
			loadSelectionControls().catch((err) => {
				console.error(' AdminDashboard: Failed to load selection controls on tab switch:', handleAPIError(err));
			});
		} else if (tab === 'selections') {
			console.log(' AdminDashboard: Starting selections tab data loading');
			selectionsLoading = true;
			console.log(' AdminDashboard: selectionsLoading set to true');
			Promise.all([
				loadStudents(),
				loadCourses(),
				loadSelections()
			]).then(() => {
				console.log(' AdminDashboard: All selections tab data loaded successfully');
				console.log(' AdminDashboard: Final data state:', {
					studentsLength: students.length,
					coursesLength: courses.length,
					selectionsLength: selections.length
				});
				selectionsLoading = false;
				console.log(' AdminDashboard: selectionsLoading set to false');
			}).catch((err) => {
				console.error(' AdminDashboard: Failed to load data for selections tab:', handleAPIError(err));
				error = handleAPIError(err);
				selectionsLoading = false;
				console.log(' AdminDashboard: selectionsLoading set to false (error case)');
			});
		}
		
		console.log(' AdminDashboard: switchTab() finished', { activeTab, selectionsLoading });
	}
</script>

<div class="admin-dashboard">
	<header class="admin-header">
		<div class="admin-header-content">
			<div class="admin-header-layout">
				<h1 class="admin-header-title">Admin Dashboard</h1>
				<div class="admin-header-actions">
					<!-- Real-time connection status -->
					<div class="connection-status">
						<div class={`connection-status-indicator ${connectionStatus ? 'connection-status-indicator--online' : 'connection-status-indicator--offline'}`}></div>
						<span class="connection-status-text">
							{connectionStatus ? 'Live' : 'Offline'}
						</span>
					</div>
					<span class="user-info">
						Welcome, {authState.username}
					</span>
					<button
						onclick={() => showChangePassword = !showChangePassword}
						class="admin-header-button admin-header-button--secondary"
					>
						Change Password
					</button>
					<button
						onclick={handleLogout}
						class="admin-header-button admin-header-button--primary"
					>
						Logout
					</button>
				</div>
			</div>
		</div>
	</header>

	<main class="admin-main-content">
		{#if showChangePassword}
			<div class="change-password-section">
				<ChangePasswordForm onClose={() => showChangePassword = false} />
			</div>
		{/if}

		{#if error}
			<div class="error-message">
				{error}
			</div>
		{/if}

		<!-- Tab Navigation -->
		<div>
			<div>
				<nav class="tab-navigation">
					<button
						onclick={() => switchTab('catalog')}
						class={`tab-button ${
							activeTab === 'catalog'
								? 'tab-button--active'
								: 'tab-button--inactive'
						}`}
					>
						Attributes
					</button>
					<button
						onclick={() => switchTab('courses')}
						class={`tab-button ${
							activeTab === 'courses'
								? 'tab-button--active'
								: 'tab-button--inactive'
						}`}
					>
						Courses
					</button>
					<button
						onclick={() => switchTab('requirements')}
						class={`tab-button ${
							activeTab === 'requirements'
								? 'tab-button--active'
								: 'tab-button--inactive'
						}`}
					>
						Requirements
					</button>
					<button
						onclick={() => switchTab('students')}
						class={`tab-button ${
							activeTab === 'students'
								? 'tab-button--active'
								: 'tab-button--inactive'
						}`}
					>
						Students
					</button>
					<button
						onclick={() => switchTab('invitations')}
						class={`tab-button ${
							activeTab === 'invitations'
								? 'tab-button--active'
								: 'tab-button--inactive'
						}`}
					>
						Invitations
					</button>
					<button
						onclick={() => switchTab('selection-controls')}
						class={`tab-button ${
							activeTab === 'selection-controls'
								? 'tab-button--active'
								: 'tab-button--inactive'
						}`}
					>
						Controls
					</button>
					<button
						onclick={() => switchTab('selections')}
						class={`tab-button ${
							activeTab === 'selections'
								? 'tab-button--active'
								: 'tab-button--inactive'
						}`}
					>
						Selections
					</button>
					<div class="tab-spacer"></div>
				</nav>
			</div>

			<div class="tab-content">
				{#if loading}
					<div class="loading-container">
						<div class="loading-text">Loading...</div>
					</div>
				{:else if activeTab === 'catalog'}
					<CatalogManager
						title="Periods"
						items={periods.map(p => ({ id: p }))}
						apiEndpoint="/api/admin/periods"
						itemName="period"
						idType="string"
						onUpdate={loadCatalogData}
					/>

					<CatalogManager
						title="Categories"
						items={categories.map(c => ({ id: c }))}
						apiEndpoint="/api/admin/categories"
						itemName="category"
						idType="string"
						onUpdate={loadCatalogData}
					/>

					<CatalogManager
						title="Grades"
						items={grades.map(g => ({ id: g }))}
						apiEndpoint="/api/admin/grades"
						itemName="grade"
						idType="number"
						onUpdate={loadCatalogData}
					/>
				{:else if activeTab === 'courses'}
					<div>
						<CourseManager
							courses={courses}
							periods={periods}
							categories={categories}
							grades={grades}
							onUpdate={loadCourses}
						/>
					</div>
				{:else if activeTab === 'requirements'}
					<div>
						<RequirementsManager
							grades={grades}
							categories={categories}
							onUpdate={() => {}}
						/>
					</div>
				{:else if activeTab === 'students'}
					<div>
						<StudentsManager
							students={students}
							grades={grades}
							onUpdate={loadStudents}
						/>
					</div>
				{:else if activeTab === 'invitations'}
					<div>
						<InvitationsManager
							students={students}
							courses={courses}
							onUpdate={loadInvitations}
						/>
					</div>
				{:else if activeTab === 'selection-controls'}
					<div>
						<SelectionControlsManager
							grades={grades}
							onUpdate={loadSelectionControls}
						/>
					</div>
				{:else if activeTab === 'selections'}
					<div>
						{#if selectionsLoading}
							<div class="selections-loading-container">
								<div class="selections-loading-text">Loading selections data...</div>
							</div>
						{:else if error}
							<div class="error-message">
								{error}
							</div>
						{:else}
							<SelectionsManager
								students={students}
								courses={courses}
								onUpdate={loadSelections}
							/>
						{/if}
					</div>
				{/if}
			</div>
		</div>
	</main>
</div>
