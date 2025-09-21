<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { api, handleAPIError } from './api';
	import { getSSEClient } from './sse';
	import type { GradeSelectionControl, SSEMessage } from './types';

	interface Props {
		grades: number[];
		onUpdate: () => void;
	}

	const { grades, onUpdate }: Props = $props();

	let controls = $state<GradeSelectionControl[]>([]);
	let loading = $state<boolean>(false);
	let error = $state<string>('');
	
	const sortedGrades = $derived([...grades].sort((a, b) => a - b));

	let gradeEnabled = $state<Map<number, boolean>>(new Map());

	let sseEventHandler: ((message: SSEMessage) => void) | null = null;

	onMount(() => {
		loadSelectionControls();
		
		const sseClient = getSSEClient();
		if (sseClient) {
			sseEventHandler = (message: SSEMessage) => {
				if (message.type === 'invalidate_selection_controls') {
					loadSelectionControls();
				}
			};
			sseClient.on('invalidate_selection_controls', sseEventHandler);
		}
	});

	onDestroy(() => {
		if (sseEventHandler) {
			const sseClient = getSSEClient();
			if (sseClient) {
				sseClient.off('invalidate_selection_controls', sseEventHandler);
			}
		}
	});

	async function loadSelectionControls(): Promise<void> {
		loading = true;
		error = '';
		try {
			controls = await api.admin.getSelectionControls();
			
			const newGradeEnabled = new Map();
			for (const grade of grades) {
				const control = controls && controls.find(c => c.grade === grade);
				newGradeEnabled.set(grade, control?.enabled || false);
			}
			gradeEnabled = newGradeEnabled;
		} catch (err) {
			error = handleAPIError(err);
		} finally {
			loading = false;
		}
	}

	async function toggleEnabled(grade: number): Promise<void> {
		const currentStatus = gradeEnabled.get(grade) || false;
		const newStatus = !currentStatus;

		try {
			const formData = new FormData();
			formData.append('grade', grade.toString());
			formData.append('enabled', newStatus.toString());

			await api.admin.createSelectionControl(formData);
			
			gradeEnabled.set(grade, newStatus);
		} catch (err) {
			error = handleAPIError(err);
			await loadSelectionControls();
		}
	}
</script>

<div class="selection-controls-manager">
	{#if error}
		<div class="selection-controls-error">
			{error}
		</div>
	{/if}

		{#if loading}
			<div class="selection-controls-loading">Loading selection controls...</div>
		{:else if !grades || grades.length === 0}
			<div class="selection-controls-empty">
				No grades found. Please add grades in the Attributes tab first.
			</div>
		{:else}
			<div>
				<div class="selection-controls-list">
					{#each sortedGrades as grade}
						{@const enabled = gradeEnabled.get(grade) || false}
						<div class="selection-control-item">
							<div class="selection-control-info">
								<span class="selection-control-grade-label">Grade {grade}</span>
								<span class="selection-control-status-badge selection-control-status-badge--{enabled ? 'enabled' : 'disabled'}">
									{enabled ? 'Enabled' : 'Disabled'}
								</span>
							</div>
							<button
								onclick={() => toggleEnabled(grade)}
								class="selection-control-toggle-button selection-control-toggle-button--{enabled ? 'disable' : 'enable'}"
							>
								{enabled ? 'Disable' : 'Enable'}
							</button>
						</div>
					{/each}
				</div>
			</div>
		{/if}
</div>
