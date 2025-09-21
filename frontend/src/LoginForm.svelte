<script lang="ts">
	import type { LoginEventDetail } from './types.ts';
	import { api, handleAPIError } from './api.ts';

	interface Props {
		onLogin: (event: LoginEventDetail) => void;
	}

	const { onLogin }: Props = $props();

	let username = $state<string>('');
	let password = $state<string>('');
	let loading = $state<boolean>(false);
	let error = $state<string>('');

	async function handleSubmit(): Promise<void> {
		console.log(' LoginForm: handleSubmit() called', { username: username.substring(0, 3) + '***' });
		if (!username.trim() || !password) {
			console.log(' LoginForm: Validation failed - missing credentials');
			error = 'Please enter both username and password';
			return;
		}

		loading = true;
		error = '';
		console.log(' LoginForm: Starting login API call');

		try {
			const result = await api.login(username.trim(), password);
			console.log(' LoginForm: Login successful', { role: result.role });
			onLogin({ role: result.role });
		} catch (err) {
			console.error(' LoginForm: Login error:', err);
			error = handleAPIError(err);
		} finally {
			loading = false;
			console.log(' LoginForm: Login attempt finished, loading =', loading);
		}
	}

	function handleFormSubmit(event: SubmitEvent): void {
		console.log(' LoginForm: handleFormSubmit() called');
		event.preventDefault();
		handleSubmit().catch((err) => {
			console.error(' LoginForm: Unexpected error in form submission:', err);
			error = 'An unexpected error occurred. Please try again.';
		});
	}
</script>

<div class="login-container">
	<div class="login-card-wrapper">
		<div class="login-card">
			<div class="login-header">
				<h1 class="login-title">Course Selection System</h1>
				<p class="login-subtitle">Please log in to continue</p>
			</div>

			<form onsubmit={handleFormSubmit} class="login-form">
				<div>
					<label for="username" class="login-label">
						Username/Student ID
					</label>
					<input
						type="text"
						id="username"
						bind:value={username}
						disabled={loading}
						class="login-input"
						placeholder="Enter username or student ID"
						required
					/>
				</div>

				<div>
					<label for="password" class="login-label">
						Password
					</label>
					<input
						type="password"
						id="password"
						bind:value={password}
						disabled={loading}
						class="login-input"
						placeholder="Enter password"
						required
					/>
				</div>

				{#if error}
					<div class="login-error">
						{error}
					</div>
				{/if}

				<button
					type="submit"
					disabled={loading}
					class="login-submit-button"
				>
					{loading ? 'Logging in...' : 'Log In'}
				</button>
			</form>

			<div class="login-footer">
				<p>Admins: Use your username</p>
				<p>Students: Use your student ID number</p>
			</div>
		</div>
	</div>
</div>
