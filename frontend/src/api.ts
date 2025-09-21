// api.ts - Type-safe API utility functions
import type { 
	APIEndpoint, 
	FetchOptions, 
	Course, 
	Student, 
	AuthState,
	LoginResponse,
	Invitation,
	GradeSelectionControl,
	Selection
} from './types.ts';

// Generic API error class
export class APIError extends Error {
	constructor(
		message: string,
		public status?: number,
		public response?: Response
	) {
		super(message);
		this.name = 'APIError';
	}
}

// Generic fetch wrapper with proper typing
export async function apiCall<T>(
	endpoint: APIEndpoint | string, 
	options: FetchOptions = {}
): Promise<T> {
	console.log(' API: Starting API call', { endpoint, method: options.method || 'GET' });
	const { method = 'GET', body = null, headers = {} } = options;

	console.log(' API: Fetch options', { method, hasBody: !!body, headers });

	const response = await fetch(endpoint, {
		method,
		body,
		headers: {
			...headers
		}
	});

	console.log(' API: Response received', { 
		endpoint, 
		status: response.status, 
		ok: response.ok,
		statusText: response.statusText
	});

	if (!response.ok) {
		const errorText = await response.text().catch(() => 'Unknown error');
		console.error(' API: Request failed', { endpoint, status: response.status, errorText });
		throw new APIError(errorText || `HTTP ${response.status}`, response.status, response);
	}

	const data = await response.json() as Promise<T>;
	console.log(' API: Response data parsed', { endpoint, data });
	return data;
}

// Specific API functions with proper typing
export const api = {
	// Authentication
	async checkAuthStatus(): Promise<AuthState> {
		console.log(' API: checkAuthStatus() called');
		const result = await apiCall<AuthState>('/api/status');
		console.log(' API: checkAuthStatus() result', result);
		return result;
	},

	async login(username: string, password: string): Promise<LoginResponse> {
		console.log(' API: login() called', { username: username.substring(0, 3) + '***' });
		const formData = new FormData();
		formData.append('username', username);
		formData.append('password', password);
		
		const result = await apiCall<LoginResponse>('/api/login', {
			method: 'POST',
			body: formData
		});
		console.log(' API: login() result', result);
		return result;
	},

	async logout(): Promise<void> {
		console.log(' API: logout() called');
		await apiCall<void>('/api/logout', { method: 'POST' });
		console.log(' API: logout() completed');
	},

	async changePassword(currentPassword: string, newPassword: string): Promise<void> {
		console.log(' API: changePassword() called');
		const formData = new FormData();
		formData.append('current_password', currentPassword);
		formData.append('new_password', newPassword);
		
		await apiCall<void>('/api/change-password', {
			method: 'POST',
			body: formData
		});
		console.log(' API: changePassword() completed');
	},

	// Admin endpoints
	admin: {
		async getGrades(): Promise<number[]> {
			console.log(' API: admin.getGrades() called');
			const result = await apiCall<number[]>('/api/admin/grades');
			console.log(' API: admin.getGrades() result', result);
			return result;
		},

		async getPeriods(): Promise<string[]> {
			console.log(' API: admin.getPeriods() called');
			const result = await apiCall<string[]>('/api/admin/periods');
			console.log(' API: admin.getPeriods() result', result);
			return result;
		},

		async getCategories(): Promise<string[]> {
			console.log(' API: admin.getCategories() called');
			const result = await apiCall<string[]>('/api/admin/categories');
			console.log(' API: admin.getCategories() result', result);
			return result;
		},

		async getCourses(): Promise<Course[]> {
			console.log(' API: admin.getCourses() called');
			const result = await apiCall<Course[]>('/api/admin/courses');
			console.log(' API: admin.getCourses() result', { result, length: Array.isArray(result) ? result.length : 'not array' });
			return result;
		},

		async getStudents(): Promise<Student[]> {
			console.log(' API: admin.getStudents() called');
			const result = await apiCall<Student[]>('/api/admin/students');
			console.log(' API: admin.getStudents() result', { result, length: Array.isArray(result) ? result.length : 'not array' });
			return result;
		},

		// Add methods for POST/PUT/DELETE operations
		async addGrade(grade: number): Promise<void> {
			const formData = new FormData();
			formData.append('grade', grade.toString());
			
			await apiCall<void>('/api/admin/grades', {
				method: 'POST',
				body: formData
			});
		},

		async deleteGrade(grade: number): Promise<void> {
			await apiCall<void>(`/api/admin/grades/${grade}`, {
				method: 'DELETE'
			});
		},

		async addPeriod(period: string): Promise<void> {
			const formData = new FormData();
			formData.append('period', period);
			
			await apiCall<void>('/api/admin/periods', {
				method: 'POST',
				body: formData
			});
		},

		async deletePeriod(period: string): Promise<void> {
			await apiCall<void>(`/api/admin/periods/${encodeURIComponent(period)}`, {
				method: 'DELETE'
			});
		},

		async addCategory(category: string): Promise<void> {
			const formData = new FormData();
			formData.append('category', category);
			
			await apiCall<void>('/api/admin/categories', {
				method: 'POST',
				body: formData
			});
		},

		async deleteCategory(category: string): Promise<void> {
			await apiCall<void>(`/api/admin/categories/${encodeURIComponent(category)}`, {
				method: 'DELETE'
			});
		},

		// Invitations
		async getInvitations(): Promise<Invitation[]> {
			return apiCall<Invitation[]>('/api/admin/invitations');
		},

		async createInvitation(formData: FormData): Promise<void> {
			await apiCall<void>('/api/admin/invitations', {
				method: 'POST',
				body: formData
			});
		},

		async updateInvitation(formData: FormData): Promise<void> {
			await apiCall<void>('/api/admin/invitations', {
				method: 'PUT',
				body: formData
			});
		},

		async deleteInvitation(studentId: number, courseId: string): Promise<void> {
			await apiCall<void>(`/api/admin/invitations?student_id=${studentId}&course_id=${encodeURIComponent(courseId)}`, {
				method: 'DELETE'
			});
		},

		async previewInvitationsCSV(formData: FormData): Promise<any> {
			return apiCall<any>('/api/admin/invitations/csv/preview', {
				method: 'POST',
				body: formData
			});
		},

		async uploadInvitationsCSV(formData: FormData): Promise<void> {
			await apiCall<void>('/api/admin/invitations/csv/upload', {
				method: 'POST',
				body: formData
			});
		},

		// Selection controls
		async getSelectionControls(): Promise<GradeSelectionControl[]> {
			return apiCall<GradeSelectionControl[]>('/api/admin/selection-controls');
		},

		async createSelectionControl(formData: FormData): Promise<void> {
			await apiCall<void>('/api/admin/selection-controls', {
				method: 'POST',
				body: formData
			});
		},

		async deleteSelectionControl(grade: number): Promise<void> {
			await apiCall<void>(`/api/admin/selection-controls?grade=${grade}`, {
				method: 'DELETE'
			});
		},

		// Selections management
		async getSelections(): Promise<Selection[]> {
			console.log(' API: admin.getSelections() called');
			const result = await apiCall<Selection[]>('/api/admin/selections');
			console.log(' API: admin.getSelections() result', { result, length: Array.isArray(result) ? result.length : 'not array' });
			return result;
		},

		async addSelection(formData: FormData): Promise<void> {
			console.log(' API: admin.addSelection() called', { formData });
			await apiCall<void>('/api/admin/selections', {
				method: 'POST',
				body: formData
			});
			console.log(' API: admin.addSelection() completed');
		},

		async updateSelection(formData: FormData): Promise<void> {
			console.log(' API: admin.updateSelection() called', { formData });
			await apiCall<void>('/api/admin/selections', {
				method: 'PUT',
				body: formData
			});
			console.log(' API: admin.updateSelection() completed');
		},

		async deleteSelection(studentId: number, courseId: string): Promise<void> {
			console.log(' API: admin.deleteSelection() called', { studentId, courseId });
			await apiCall<void>(`/api/admin/selections?student_id=${studentId}&course_id=${encodeURIComponent(courseId)}`, {
				method: 'DELETE'
			});
			console.log(' API: admin.deleteSelection() completed');
		}
	},

	// Student endpoints  
	student: {
		async getCourses(): Promise<Course[]> {
			return apiCall<Course[]>('/api/student/courses');
		},

		async getSelections(): Promise<Course[]> {
			return apiCall<Course[]>('/api/student/selections');
		},

		async selectCourse(courseId: string): Promise<void> {
			const formData = new FormData();
			formData.append('course_id', courseId);
			
			await apiCall<void>('/api/student/select', {
				method: 'POST',
				body: formData
			});
		},

		async deselectCourse(courseId: string): Promise<void> {
			const formData = new FormData();
			formData.append('course_id', courseId);
			
			await apiCall<void>('/api/student/deselect', {
				method: 'POST',
				body: formData
			});
		}
	}
};

// Helper to handle common API error scenarios
export function handleAPIError(error: unknown): string {
	if (error instanceof APIError) {
		return error.message;
	} else if (error instanceof Error) {
		return error.message;
	} else {
		return 'An unexpected error occurred';
	}
}