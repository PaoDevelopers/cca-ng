// types.ts - Comprehensive type definitions for the CCA application

// Authentication types
export interface AuthState {
	authenticated: boolean;
	role?: 'admin' | 'student';
	username?: string;
	user_id?: number;
}

export interface LoginRequest {
	username: string;
	password: string;
}

export interface LoginResponse {
	role: 'admin' | 'student';
	username?: string;
	user_id?: number;
}

// User types
export type LegalSex = 'F' | 'M' | 'X';

export interface Student {
	id: number;
	name: string;
	grade: number;
	legal_sex: LegalSex;
	password_hash?: string; // Only present in admin views
}

export interface Admin {
	id: number;
	username: string;
	password_hash?: string; // Only present in admin views
}

// Course and selection types
export type MembershipType = 'free-choice' | 'invite-only';
export type SexRestriction = 'F' | 'M' | 'X' | 'Any';
export type InvitationType = 'no' | 'invite' | 'force';

export interface Course {
	id: string;
	name: string;
	description: string;
	period: string;
	max_students: number;
	grade_restrictions: number[];
	sex_restriction: SexRestriction;
	membership_type: MembershipType;
	teacher: string;
	location: string;
	category: string;
	current_students?: number; // Current enrollment count
}

export interface Choice {
	student_id: number;
	course_id: string;
	invitation_type: InvitationType;
}

export interface Selection {
	student_id: number;
	student_name: string;
	student_grade: number;
	course_id: string;
	course_name: string;
	period_id: string;
	invitation_type: InvitationType;
}

// Requirements types
export interface Requirement {
	id: number;
	grade: number;
	min_courses_total: number;
	category_requirements: CategoryRequirement[];
	forbidden_combinations: ForbiddenCombination[];
}

export interface CategoryRequirement {
	category: string;
	min_courses: number;
}

export interface ForbiddenCombination {
	periods: string[];
}

// CSV import types
export interface CourseCSVRow {
	id: string;
	name: string;
	description: string;
	period: string;
	max_students: string;
	grade_restrictions: string;
	sex_restriction: SexRestriction;
	membership_type: MembershipType;
	teacher: string;
	location: string;
	category: string;
}

export interface StudentCSVRow {
	id: string;
	name: string;
	grade: string;
	legal_sex: LegalSex;
	password: string;
}

export interface InvitationCSVRow {
	student_id: string;
	course_id: string;
	invitation_type: 'invite' | 'force';
}

// SSE types
export interface SSEMessage {
	type: string;
	data?: unknown;
}

// API response types
export interface APIResponse<T = unknown> {
	success: boolean;
	data?: T;
	error?: string;
}

export interface ValidationError {
	field: string;
	message: string;
	row?: number; // For CSV validation errors
}

export interface CSVPreview<T> {
	valid: boolean;
	rows: T[];
	errors: ValidationError[];
}

// Component prop types
export interface Props {
	// Basic props that most components might have
	class?: string;
	disabled?: boolean;
}

// Event handler types
export interface LoginEventDetail {
	role: 'admin' | 'student';
}

export interface PasswordChangeEventDetail {
	success: boolean;
}

// Selection status types
export interface StudentSelectionStatus {
	student_id: number;
	student_name: string;
	grade: number;
	selections: StudentSelection[];
	meets_requirements: boolean;
	requirement_issues: string[];
}

export interface StudentSelection {
	course_id: string;
	course_name: string;
	period: string;
	invitation_type: InvitationType;
}

// Export/report types
export interface SelectionExport {
	student_id: number;
	student_name: string;
	grade: number;
	legal_sex: LegalSex;
	course_id: string;
	course_name: string;
	period: string;
	invitation_type: InvitationType;
}

// Invitation types
export interface Invitation {
	student_id: number;
	course_id: string;
	period_id: string;
	invitation_type: InvitationType;
	student_name: string;
	student_grade: number;
	course_name: string;
	course_period: string;
}

// Grade selection controls
export interface GradeSelectionControl {
	grade: number;
	enabled: boolean;
}

// Utility types for forms
export interface FormState<T> {
	data: T;
	loading: boolean;
	error: string;
	touched: Record<keyof T, boolean>;
}

// Type guards
export function isStudent(user: Student | Admin): user is Student {
	return 'grade' in user && 'legal_sex' in user;
}

export function isAdmin(user: Student | Admin): user is Admin {
	return 'username' in user && !('grade' in user);
}

export function isValidLegalSex(value: string): value is LegalSex {
	return value === 'F' || value === 'M' || value === 'X';
}

export function isValidSexRestriction(value: string): value is SexRestriction {
	return value === 'F' || value === 'M' || value === 'X' || value === 'Any';
}

export function isValidMembershipType(value: string): value is MembershipType {
	return value === 'free-choice' || value === 'invite-only';
}

export function isValidInvitationType(value: string): value is InvitationType {
	return value === 'no' || value === 'invite' || value === 'force';
}

// API endpoint types - helps with fetch calls
export type APIEndpoint = 
	| '/api/status'
	| '/api/login'
	| '/api/logout'
	| '/api/change-password'
	| '/api/admin/grades'
	| '/api/admin/periods'
	| '/api/admin/categories'
	| '/api/admin/courses'
	| '/api/admin/students'
	| '/api/admin/requirements'
	| '/api/admin/invitations'
	| '/api/admin/selections'
	| '/api/admin/export'
	| '/api/student/courses'
	| '/api/student/selections'
	| '/api/student/select'
	| '/api/student/deselect'
	| '/api/events';

// HTTP method types
export type HTTPMethod = 'GET' | 'POST' | 'PUT' | 'DELETE';

// Generic fetch wrapper type
export interface FetchOptions {
	method?: HTTPMethod;
	body?: BodyInit | null;
	headers?: Record<string, string>;
}