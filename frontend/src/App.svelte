<script lang="ts">
	import LoginForm from './LoginForm.svelte';
	import AdminDashboard from './AdminDashboard.svelte';
	import StudentDashboard from './StudentDashboard.svelte';
	import { initializeSSE, disconnectSSE } from './sse.ts';
	import type { AuthState, LoginEventDetail } from './types.ts';
	import { api, handleAPIError } from './api.ts';

	let authState = $state<AuthState>({ authenticated: false });
	let loading = $state<boolean>(true);

	console.log('£ App: Component initialized', { authState, loading });

	$effect(() => {
		console.log('£ App: SSE effect triggered', { authenticated: authState.authenticated });
		if (authState.authenticated) {
			console.log('£ App: User authenticated, initializing SSE');
			initializeSSE();
		} else {
			console.log('£ App: User not authenticated, disconnecting SSE');
			disconnectSSE();
		}
		
		return () => {
			console.log('£ App: SSE effect cleanup, disconnecting SSE');
			disconnectSSE();
		};
	});

	$effect(() => {
		console.log('£ App: Auth check effect triggered, checking auth status');
		checkAuthStatus().then(() => {
			console.log('£ App: Auth check completed successfully');
			loading = false;
		}).catch((error) => {
			console.error(' App: Failed to check auth status during initialization:', error);
			loading = false;
		});
	});

	async function checkAuthStatus(): Promise<void> {
		console.log('£ App: checkAuthStatus() called');
		try {
			const result = await api.checkAuthStatus();
			console.log('£ App: checkAuthStatus() successful', result);
			authState = result;
		} catch (error) {
			console.error(' App: checkAuthStatus() failed:', handleAPIError(error));
			authState = { authenticated: false };
		}
	}

	function handleLogin(event: LoginEventDetail): void {
		console.log('£ App: handleLogin() called', event);
		authState = {
			authenticated: true,
			role: event.role
		};
		console.log('£ App: Auth state updated after login', { authState });
		checkAuthStatus().catch((error) => {
			console.error(' App: Failed to refresh auth status after login:', handleAPIError(error));
		});
	}

	async function handleLogout(): Promise<void> {
		console.log('£ App: handleLogout() called');
		try {
			await api.logout();
			console.log('£ App: Logout API call successful');
		} catch (error) {
			console.error(' App: Logout error:', handleAPIError(error));
		}
		console.log('£ App: Disconnecting SSE and updating auth state');
		disconnectSSE();
		authState = { authenticated: false };
		console.log('£ App: Auth state updated after logout', { authState });
	}
</script>

<div class="app-container">
	{#if loading}
		<div class="app-loading-screen">
			<div class="app-loading-text">Loading...</div>
		</div>
	{:else if !authState.authenticated}
		<LoginForm onLogin={handleLogin} />
	{:else if authState.role === 'admin'}
		<AdminDashboard {authState} onLogout={handleLogout} />
	{:else if authState.role === 'student'}
		<StudentDashboard {authState} onLogout={handleLogout} />
	{:else}
		<div class="app-error-screen">
			<div class="app-error-text">Unknown user role</div>
		</div>
	{/if}
</div>
