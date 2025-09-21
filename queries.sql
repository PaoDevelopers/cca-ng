-- name: GetSchemaVersion :one
SELECT version FROM schema_version LIMIT 1;

-- name: CreateAdmin :one
INSERT INTO admins (username, password_hash)
VALUES ($1, $2)
RETURNING id, username, created_at;

-- name: GetAdminByUsername :one
SELECT id, username, password_hash, session_token, created_at, updated_at
FROM admins
WHERE username = $1;

-- name: GetAdminBySessionToken :one
SELECT id, username, password_hash, session_token, created_at, updated_at
FROM admins
WHERE session_token = $1;

-- name: UpdateAdminSessionToken :exec
UPDATE admins
SET session_token = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateAdminPassword :exec
UPDATE admins
SET password_hash = $2, updated_at = NOW()
WHERE id = $1;

-- name: ClearAdminSessionToken :exec
UPDATE admins
SET session_token = NULL, updated_at = NOW()
WHERE id = $1;

-- name: CountAdmins :one
SELECT COUNT(*) FROM admins;

-- name: CreateStudent :one
INSERT INTO students (id, name, grade, legal_sex, password_hash)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, grade, legal_sex, created_at;

-- name: GetStudentByID :one
SELECT id, name, grade, legal_sex, password_hash, session_token, created_at, updated_at
FROM students
WHERE id = $1;

-- name: GetStudentBySessionToken :one
SELECT id, name, grade, legal_sex, password_hash, session_token, created_at, updated_at
FROM students
WHERE session_token = $1;

-- name: UpdateStudentSessionToken :exec
UPDATE students
SET session_token = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateStudentPassword :exec
UPDATE students
SET password_hash = $2, updated_at = NOW()
WHERE id = $1;

-- name: ClearStudentSessionToken :exec
UPDATE students
SET session_token = NULL, updated_at = NOW()
WHERE id = $1;

-- Grades management
-- name: GetAllGrades :many
SELECT grade FROM grades ORDER BY grade;

-- name: CreateGrade :exec
INSERT INTO grades (grade) VALUES ($1);

-- name: DeleteGrade :exec
DELETE FROM grades WHERE grade = $1;

-- Periods management
-- name: GetAllPeriods :many
SELECT id FROM periods ORDER BY id;

-- name: CreatePeriod :exec
INSERT INTO periods (id) VALUES ($1);

-- name: DeletePeriod :exec
DELETE FROM periods WHERE id = $1;

-- Categories management
-- name: GetAllCategories :many
SELECT id FROM categories ORDER BY id;

-- name: CreateCategory :exec
INSERT INTO categories (id) VALUES ($1);

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = $1;

-- Courses management
-- name: GetAllCourses :many
SELECT 
	c.id,
	c.name,
	c.description,
	c.period_id,
	c.max_students,
	c.sex_restriction,
	c.membership,
	c.teacher,
	c.location,
	c.category_id
FROM courses c ORDER BY c.id;

-- name: GetCourseByID :one
SELECT 
	c.id,
	c.name,
	c.description,
	c.period_id,
	c.max_students,
	c.sex_restriction,
	c.membership,
	c.teacher,
	c.location,
	c.category_id
FROM courses c WHERE c.id = $1;

-- name: CreateCourse :exec
INSERT INTO courses (
	id, name, description, period_id, max_students, 
	sex_restriction, membership, teacher, location, category_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: UpdateCourse :exec
UPDATE courses SET
	name = $2,
	description = $3,
	period_id = $4,
	max_students = $5,
	sex_restriction = $6,
	membership = $7,
	teacher = $8,
	location = $9,
	category_id = $10
WHERE id = $1;

-- name: DeleteCourse :exec
DELETE FROM courses WHERE id = $1;

-- name: DeleteAllCourses :exec
DELETE FROM courses;

-- Course allowed grades management
-- name: GetCourseAllowedGrades :many
SELECT grade FROM course_allowed_grades WHERE course_id = $1 ORDER BY grade;

-- name: SetCourseAllowedGrades :exec
DELETE FROM course_allowed_grades WHERE course_id = $1;

-- name: InsertCourseAllowedGrades :copyfrom
INSERT INTO course_allowed_grades (course_id, grade) VALUES ($1, $2);

-- name: DeleteCourseAllowedGrades :exec
DELETE FROM course_allowed_grades WHERE course_id = $1;

-- Students management (for Tab 4)
-- name: GetAllStudents :many
SELECT id, name, grade, legal_sex, created_at, updated_at 
FROM students ORDER BY id;

-- name: UpdateStudent :exec
UPDATE students 
SET name = $2, grade = $3, legal_sex = $4, updated_at = NOW()
WHERE id = $1;

-- name: DeleteStudent :exec
DELETE FROM students WHERE id = $1;

-- name: DeleteAllStudents :exec
DELETE FROM students;

-- Requirements management (for Tab 3)
-- name: GetGradeRequirement :one
SELECT grade, min_total FROM grade_requirements WHERE grade = $1;

-- name: SetGradeRequirement :exec
INSERT INTO grade_requirements (grade, min_total) 
VALUES ($1, $2)
ON CONFLICT (grade) DO UPDATE SET min_total = EXCLUDED.min_total;

-- name: GetAllGradeRequirements :many
SELECT grade, min_total FROM grade_requirements ORDER BY grade;

-- name: DeleteGradeRequirement :exec
DELETE FROM grade_requirements WHERE grade = $1;

-- name: GetGradeRequirementGroups :many
SELECT id, grade, label, min_count 
FROM grade_requirement_groups 
WHERE grade = $1 
ORDER BY id;

-- name: CreateGradeRequirementGroup :one
INSERT INTO grade_requirement_groups (grade, label, min_count)
VALUES ($1, $2, $3)
RETURNING id;

-- name: UpdateGradeRequirementGroup :exec
UPDATE grade_requirement_groups 
SET label = $2, min_count = $3
WHERE id = $1;

-- name: DeleteGradeRequirementGroup :exec
DELETE FROM grade_requirement_groups WHERE id = $1;

-- name: GetGradeRequirementGroupCategories :many
SELECT category_id 
FROM grade_requirement_group_categories 
WHERE req_group_id = $1
ORDER BY category_id;

-- name: DeleteGradeRequirementGroupCategories :exec
DELETE FROM grade_requirement_group_categories WHERE req_group_id = $1;

-- name: InsertGradeRequirementGroupCategories :copyfrom
INSERT INTO grade_requirement_group_categories (req_group_id, category_id) 
VALUES ($1, $2);

-- Student-specific queries for course selection

-- name: GetStudentAvailableCourses :many
SELECT 
	c.id,
	c.name,
	c.description,
	c.period_id,
	c.max_students,
	c.sex_restriction,
	c.membership,
	c.teacher,
	c.location,
	c.category_id,
	COALESCE(ec.current_enrollment, 0) AS current_enrollment,
	CASE 
		WHEN ch.student_id IS NOT NULL THEN ch.invitation_type
		ELSE 'no'::invitation_type
	END AS student_invitation_type,
	-- Check if student can select this course (independent of selection window status)
	CASE
		WHEN ch.student_id IS NOT NULL THEN 'selected'
		WHEN c.membership = 'invite_only' THEN 'invite_only'
		WHEN c.max_students > 0 AND COALESCE(ec.current_enrollment, 0) >= c.max_students THEN 'at_capacity'
		WHEN c.sex_restriction != 'ANY' AND c.sex_restriction::text != $2::text THEN 'sex_restriction'
		WHEN EXISTS (
			SELECT 1 FROM course_allowed_grades cag WHERE cag.course_id = c.id
		) AND NOT EXISTS (
			SELECT 1 FROM course_allowed_grades cag WHERE cag.course_id = c.id AND cag.grade = $3
		) THEN 'grade_restriction'
		ELSE 'available'
	END AS availability_status
FROM courses c
LEFT JOIN v_course_enrollment_counts ec ON ec.course_id = c.id
LEFT JOIN choices ch ON ch.course_id = c.id AND ch.student_id = $1
ORDER BY c.period_id, c.name;

-- name: GetStudentSelections :many  
SELECT 
	c.id AS course_id,
	c.name AS course_name,
	c.description,
	c.period_id,
	c.teacher,
	c.location,
	c.category_id,
	ch.invitation_type
FROM choices ch
JOIN courses c ON c.id = ch.course_id
WHERE ch.student_id = $1
ORDER BY c.period_id, c.name;

-- name: AddStudentSelection :exec
INSERT INTO choices (student_id, course_id, period_id, invitation_type)
VALUES ($1, $2, $3, 'no');

-- name: RemoveStudentSelection :exec
DELETE FROM choices 
WHERE student_id = $1 AND course_id = $2;

-- name: GetCourseCurrentEnrollment :one
SELECT COALESCE(current_enrollment, 0) AS current_enrollment
FROM v_course_enrollment_counts
WHERE course_id = $1;

-- name: GetStudentRequirementsStatus :one
SELECT
	s.grade,
	COALESCE(gr.min_total, 0) AS required_total,
	COUNT(ch.student_id)::BIGINT AS current_total
FROM students s
LEFT JOIN grade_requirements gr ON gr.grade = s.grade
LEFT JOIN choices ch ON ch.student_id = s.id
WHERE s.id = $1
GROUP BY s.grade, gr.min_total;

-- name: GetStudentRequirementGroupsStatus :many
SELECT
	grg.id,
	grg.label,
	grg.min_count AS required_count,
	COUNT(ch.student_id)::BIGINT AS current_count,
	array_agg(DISTINCT grgc.category_id ORDER BY grgc.category_id) AS required_categories
FROM students s
JOIN grade_requirement_groups grg ON grg.grade = s.grade
LEFT JOIN grade_requirement_group_categories grgc ON grgc.req_group_id = grg.id
LEFT JOIN choices ch ON ch.student_id = s.id
LEFT JOIN courses c ON c.id = ch.course_id AND c.category_id = grgc.category_id
WHERE s.id = $1
GROUP BY grg.id, grg.label, grg.min_count
ORDER BY grg.id;

-- name: GetStudentSelectionsByPeriod :many
SELECT 
	p.id AS period_id,
	c.id AS course_id,
	c.name AS course_name,
	c.teacher,
	c.location,
	c.category_id,
	ch.invitation_type
FROM periods p
LEFT JOIN choices ch ON ch.period_id = p.id AND ch.student_id = $1
LEFT JOIN courses c ON c.id = ch.course_id
ORDER BY p.id;

-- name: GetGradeSelectionControl :one
SELECT grade, enabled
FROM grade_selection_controls
WHERE grade = $1;

-- Invitations (choices with invitation_type != 'no')
-- name: GetAllInvitations :many
SELECT 
	ch.student_id,
	ch.course_id,
	ch.period_id,
	ch.invitation_type,
	s.name AS student_name,
	s.grade AS student_grade,
	c.name AS course_name,
	c.period_id AS course_period
FROM choices ch
JOIN students s ON s.id = ch.student_id
JOIN courses c ON c.id = ch.course_id
WHERE ch.invitation_type != 'no'
ORDER BY s.grade, s.name, c.period_id, c.name;

-- name: GetInvitation :one
SELECT 
	ch.student_id,
	ch.course_id,
	ch.period_id,
	ch.invitation_type,
	s.name AS student_name,
	s.grade AS student_grade,
	c.name AS course_name
FROM choices ch
JOIN students s ON s.id = ch.student_id
JOIN courses c ON c.id = ch.course_id
WHERE ch.student_id = $1 AND ch.course_id = $2
AND ch.invitation_type != 'no';

-- name: CreateOrUpdateInvitation :exec
INSERT INTO choices (student_id, course_id, period_id, invitation_type)
VALUES ($1, $2, $3, $4)
ON CONFLICT (student_id, course_id) DO UPDATE SET
	invitation_type = EXCLUDED.invitation_type;

-- name: DeleteInvitation :exec
DELETE FROM choices 
WHERE student_id = $1 AND course_id = $2 AND invitation_type != 'no';

-- name: DeleteAllInvitations :exec
DELETE FROM choices WHERE invitation_type != 'no';

-- Grade selection controls management
-- name: GetAllGradeSelectionControls :many
SELECT grade, enabled
FROM grade_selection_controls
ORDER BY grade;

-- name: CreateOrUpdateGradeSelectionControl :exec
INSERT INTO grade_selection_controls (grade, enabled)
VALUES ($1, $2)
ON CONFLICT (grade) DO UPDATE SET
	enabled = EXCLUDED.enabled;

-- name: DeleteGradeSelectionControl :exec
DELETE FROM grade_selection_controls WHERE grade = $1;

-- Admin selections management
-- name: GetAllSelections :many
SELECT 
	ch.student_id,
	s.name AS student_name,
	s.grade AS student_grade,
	ch.course_id,
	c.name AS course_name,
	ch.period_id,
	ch.invitation_type
FROM choices ch
JOIN students s ON s.id = ch.student_id
JOIN courses c ON c.id = ch.course_id
ORDER BY s.id, ch.period_id, c.name;

-- name: AdminAddSelection :exec
INSERT INTO choices (student_id, course_id, period_id, invitation_type)
VALUES ($1, $2, $3, $4);

-- name: AdminUpdateSelection :exec
UPDATE choices 
SET course_id = $3, period_id = $4, invitation_type = $5
WHERE student_id = $1 AND course_id = $2;

-- name: AdminDeleteSelection :exec
DELETE FROM choices 
WHERE student_id = $1 AND course_id = $2;