<script lang="ts">
	import { api, handleAPIError } from './api.ts';

	interface Props {
		onClose: () => void;
	}

	const { onClose }: Props = $props();

	let currentPassword = $state<string>('');
	let newPassword = $state<string>('');
	let confirmPassword = $state<string>('');
	let loading = $state<boolean>(false);
	let error = $state<string>('');
	let success = $state<boolean>(false);

	async function handleSubmit(): Promise<void> {
		if (!currentPassword || !newPassword || !confirmPassword) {
			error = 'Please fill in all fields';
			return;
		}

		if (newPassword !== confirmPassword) {
			error = 'New passwords do not match';
			return;
		}

		if (newPassword.length < 4) {
			error = 'New password must be at least 4 characters long';
			return;
		}

		loading = true;
		error = '';

		try {
			await api.changePassword(currentPassword, newPassword);
			success = true;
			currentPassword = '';
			newPassword = '';
			confirmPassword = '';
			setTimeout(() => {
				onClose();
			}, 2000);
		} catch (err) {
			console.error('Password change error:', err);
			error = handleAPIError(err);
		} finally {
			loading = false;
		}
	}

	function handleClose(): void {
		onClose();
	}

	function handleFormSubmit(event: SubmitEvent): void {
		event.preventDefault();
		handleSubmit().catch((err) => {
			console.error('Unexpected error in password form submission:', err);
			error = 'An unexpected error occurred. Please try again.';
		});
	}
</script>

<div class="password-form-card">
	<div class="password-form-header">
		<div class="password-form-header-layout">
			<h3 class="password-form-title">Change Password</h3>
			<button
				onclick={handleClose}
				class="password-form-close-button"
			>
				<span class="password-form-close-label">Close</span>
				<svg class="password-form-close-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>
		</div>
	</div>
	<div class="password-form-body">
		{#if success}
			<div class="password-form-success">
				Password changed successfully! Closing...
			</div>
		{:else}
			<form onsubmit={handleFormSubmit} class="password-form">
				<div class="password-form-group">
					<label for="current-password" class="password-form-label">
						Current Password
					</label>
					<input
						type="password"
						id="current-password"
						bind:value={currentPassword}
						disabled={loading}
						class="password-form-input"
						required
					/>
				</div>

				<div class="password-form-group">
					<label for="new-password" class="password-form-label">
						New Password
					</label>
					<input
						type="password"
						id="new-password"
						bind:value={newPassword}
						disabled={loading}
						class="password-form-input"
						required
					/>
				</div>

				<div class="password-form-group">
					<label for="confirm-password" class="password-form-label">
						Confirm New Password
					</label>
					<input
						type="password"
						id="confirm-password"
						bind:value={confirmPassword}
						disabled={loading}
						class="password-form-input"
						required
					/>
				</div>

				{#if error}
					<div class="password-form-error">
						{error}
					</div>
				{/if}

				<div class="password-form-actions">
					<button
						type="button"
						onclick={handleClose}
						disabled={loading}
						class="password-form-button password-form-button--secondary"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={loading}
						class="password-form-button password-form-button--primary"
					>
						{loading ? 'Changing...' : 'Change Password'}
					</button>
				</div>
			</form>
		{/if}
	</div>
</div>
