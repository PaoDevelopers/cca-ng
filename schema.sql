CREATE TABLE schema_version (
	singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton),
	version BIGINT NOT NULL CHECK (version > 0)
);
INSERT INTO schema_version (version) VALUES (1);


-- Enums

CREATE TYPE legal_sex AS ENUM ('F', 'M', 'X');
CREATE TYPE sex_restriction AS ENUM ('ANY', 'F', 'M', 'X');
CREATE TYPE invitation_type AS ENUM ('no', 'invite', 'force');
CREATE TYPE membership_type AS ENUM ('free', 'invite_only');


-- Catalog tables

CREATE TABLE grades (
	grade BIGINT PRIMARY KEY -- yes, any bigint, including negative. this is correct.
);

CREATE TABLE periods (
	id TEXT PRIMARY KEY,
	CONSTRAINT chk_periods_id_nonblank CHECK (btrim(id) <> '')
);

CREATE TABLE categories (
	id TEXT PRIMARY KEY,
	CONSTRAINT chk_categories_id_nonblank CHECK (btrim(id) <> '')
);


-- Users

CREATE TABLE admins (
	id BIGSERIAL PRIMARY KEY,
	username TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	session_token TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT chk_admins_username_nonblank CHECK (btrim(username) <> '')
);

CREATE TABLE students (
	id BIGINT PRIMARY KEY,
	name TEXT NOT NULL,
	grade BIGINT NOT NULL REFERENCES grades(grade) ON UPDATE RESTRICT ON DELETE RESTRICT,
	legal_sex legal_sex NOT NULL,
	password_hash TEXT NOT NULL,
	session_token TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT chk_students_name_nonblank CHECK (btrim(name) <> '')
);

-- cookies encode the user type so there won't be collisions on session_token


-- Courses and their allowed grades

CREATE TABLE courses (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	period_id TEXT NOT NULL REFERENCES periods(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
	max_students BIGINT NOT NULL CHECK (max_students >= 0),
	sex_restriction sex_restriction NOT NULL DEFAULT 'ANY',
	membership membership_type NOT NULL DEFAULT 'free',
	teacher TEXT NOT NULL,
	location TEXT NOT NULL,
	category_id TEXT NOT NULL REFERENCES categories(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
	-- this UNIQUE is intentionally kept even though id is PK, so the composite FK from choices can ensure stored period_id matches the course’s period.
	UNIQUE (id, period_id),
	CONSTRAINT chk_courses_name_nonblank CHECK (btrim(name) <> ''), -- teacher/location/description can be blank
	CONSTRAINT chk_courses_id_nonblank CHECK (btrim(id) <> '')
);

CREATE TABLE course_allowed_grades (
	course_id TEXT NOT NULL REFERENCES courses(id) ON UPDATE CASCADE ON DELETE CASCADE,
	grade BIGINT NOT NULL REFERENCES grades(grade) ON UPDATE RESTRICT ON DELETE RESTRICT,
	PRIMARY KEY (course_id, grade)
);


-- Choices (student selections and/or invitations)

-- PK(student_id, period_id) enforces one course per period per student.
-- UNIQUE(student_id, course_id) prevents duplicate selection/invite of same course.
-- Store period_id and constrain via composite FK to ensure it matches the course period.
-- This is exempt from the 3NF rule, since it still preserves integrity well
-- and avoids too many triggers.
CREATE TABLE choices (
	student_id BIGINT NOT NULL REFERENCES students(id) ON UPDATE CASCADE ON DELETE CASCADE,
	course_id TEXT NOT NULL,
	period_id TEXT NOT NULL,
	invitation_type invitation_type NOT NULL DEFAULT 'no',
	PRIMARY KEY (student_id, period_id),
	UNIQUE (student_id, course_id),
	FOREIGN KEY (course_id, period_id) REFERENCES courses(id, period_id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- Enforce sex/grade/membership and selection window only when invitation_type = 'no'.
-- Invites/forces bypass these checks by design.
CREATE FUNCTION enforce_eligibility_when_no_invitation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
	v_student_grade BIGINT;
	v_student_sex   legal_sex;
	v_sex_restr     sex_restriction;
	v_membership    membership_type;
	v_has_grade_list  boolean;
	v_grade_allowed   boolean;
BEGIN
	-- Skip checks for invites/forces.
	IF NEW.invitation_type <> 'no' THEN
		RETURN NEW;
	END IF;

	-- Student facts
	SELECT s.grade, s.legal_sex
	INTO v_student_grade, v_student_sex
	FROM students s
	WHERE s.id = NEW.student_id;

	IF v_student_grade IS NULL THEN
		RAISE EXCEPTION 'Student % not found', NEW.student_id
			USING ERRCODE = 'foreign_key_violation';
	END IF;

	-- Course restrictions
	SELECT c.sex_restriction, c.membership
	INTO v_sex_restr, v_membership
	FROM courses c
	WHERE c.id = NEW.course_id;

	IF v_sex_restr IS NULL THEN
		RAISE EXCEPTION 'Course % not found', NEW.course_id
			USING ERRCODE = 'foreign_key_violation';
	END IF;

	-- Invite-only courses require an invitation (normal selections blocked)
	IF v_membership = 'invite_only' THEN
		RAISE EXCEPTION 'Course % is invite-only; invitation required', NEW.course_id
			USING ERRCODE = 'check_violation';
	END IF;

	-- Sex restriction (ANY passes)
	IF v_sex_restr <> 'ANY' AND v_student_sex::text <> v_sex_restr::text THEN
		RAISE EXCEPTION 'Student % legal sex % not allowed for course % (requires %)',
			NEW.student_id, v_student_sex, NEW.course_id, v_sex_restr
			USING ERRCODE = 'check_violation';
	END IF;

	-- Grade restriction:
	-- If no rows in course_allowed_grades, treat as "no grade restriction".
	SELECT
		EXISTS (SELECT 1 FROM course_allowed_grades g WHERE g.course_id = NEW.course_id),
		EXISTS (SELECT 1 FROM course_allowed_grades g WHERE g.course_id = NEW.course_id AND g.grade = v_student_grade)
	INTO v_has_grade_list, v_grade_allowed;

	IF v_has_grade_list AND NOT v_grade_allowed THEN
		RAISE EXCEPTION 'Student % grade % not allowed for course %',
			NEW.student_id, v_student_grade, NEW.course_id
			USING ERRCODE = 'check_violation';
	END IF;

	-- Selection window gate for normal selections (enabled only)
	PERFORM 1
	FROM grade_selection_controls g
	WHERE g.grade = v_student_grade
	  AND g.enabled = TRUE;

	IF NOT FOUND THEN
		RAISE EXCEPTION 'Selections are closed for grade %', v_student_grade
			USING ERRCODE = 'check_violation';
	END IF;

	RETURN NEW;
END
$$;

-- Fire on INSERT for normal selections.
CREATE TRIGGER trg_choices_eligibility_ins
BEFORE INSERT ON choices
FOR EACH ROW
EXECUTE FUNCTION enforce_eligibility_when_no_invitation();

-- Fire on UPDATE when a row becomes/acts like a normal selection,
-- or when its destination course changes.
CREATE TRIGGER trg_choices_eligibility_upd
BEFORE UPDATE OF course_id, invitation_type ON choices
FOR EACH ROW
WHEN (
	NEW.invitation_type = 'no'
	AND (OLD.invitation_type IS DISTINCT FROM 'no' OR OLD.course_id IS DISTINCT FROM NEW.course_id)
)
EXECUTE FUNCTION enforce_eligibility_when_no_invitation();


-- Enforce capacity only when the resulting row is a normal choice (invitation_type = 'no').
-- Invites/forces bypass capacity; counts include all choices so invites reserve spots.
CREATE FUNCTION enforce_capacity_when_no_invitation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
	v_max   bigint;
	v_count bigint;
BEGIN
	-- Only enforce for normal selections
	IF NEW.invitation_type <> 'no' THEN
		RETURN NEW;
	END IF;

	-- Serialize checks per course
	SELECT max_students
	INTO v_max
	FROM courses
	WHERE id = NEW.course_id
	FOR UPDATE;

	IF v_max IS NULL THEN
		RAISE EXCEPTION 'Course % not found', NEW.course_id
			USING ERRCODE = 'foreign_key_violation';
	END IF;

	-- Count all choices for the destination course
	SELECT COUNT(*)::bigint
	INTO v_count
	FROM choices
	WHERE course_id = NEW.course_id;

	-- For UPDATE within the same course, the current row is already counted;
	-- neutralize it so reclassifying invite->no doesn't spuriously fail.
	IF TG_OP = 'UPDATE' AND OLD.course_id = NEW.course_id THEN
		v_count := v_count - 1;
	END IF;

	IF v_count >= v_max THEN
		RAISE EXCEPTION 'Course % is at capacity (% >= %)', NEW.course_id, v_count, v_max
			USING ERRCODE = 'check_violation';
	END IF;

	RETURN NEW;
END
$$;

CREATE TRIGGER trg_choices_capacity_ins
BEFORE INSERT ON choices
FOR EACH ROW
EXECUTE FUNCTION enforce_capacity_when_no_invitation();

CREATE TRIGGER trg_choices_capacity_upd
BEFORE UPDATE OF course_id, invitation_type ON choices
FOR EACH ROW
WHEN (
	NEW.invitation_type = 'no'
	AND (OLD.invitation_type IS DISTINCT FROM 'no' OR OLD.course_id IS DISTINCT FROM NEW.course_id)
)
EXECUTE FUNCTION enforce_capacity_when_no_invitation();


-- Selection control per grade
CREATE TABLE grade_selection_controls (
	grade BIGINT PRIMARY KEY REFERENCES grades(grade)
		ON UPDATE RESTRICT ON DELETE RESTRICT,
	enabled BOOLEAN NOT NULL DEFAULT FALSE
);


-- Per-grade minimum total selections

CREATE TABLE grade_requirements (
	grade BIGINT PRIMARY KEY REFERENCES grades(grade) ON UPDATE RESTRICT ON DELETE CASCADE,
	min_total BIGINT NOT NULL DEFAULT 0 CHECK (min_total >= 0)
);

-- Category-group minima

CREATE TABLE grade_requirement_groups (
	id BIGSERIAL PRIMARY KEY,
	grade BIGINT NOT NULL REFERENCES grades(grade) ON UPDATE RESTRICT ON DELETE CASCADE,
	label TEXT NOT NULL DEFAULT '',
	min_count BIGINT NOT NULL CHECK (min_count >= 0)
);

CREATE TABLE grade_requirement_group_categories (
	req_group_id BIGINT NOT NULL REFERENCES grade_requirement_groups(id) ON UPDATE CASCADE ON DELETE CASCADE,
	category_id TEXT NOT NULL REFERENCES categories(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
	PRIMARY KEY (req_group_id, category_id)
);



-- Views

-- Enrollment counts per course (counts all choices irrespective of invitation type)
CREATE VIEW v_course_enrollment_counts AS
SELECT
	c.id AS course_id,
	COUNT(ch.student_id)::BIGINT AS current_enrollment
FROM courses c
LEFT JOIN choices ch
	ON ch.course_id = c.id
GROUP BY c.id;

-- Export projection matching required CSV columns and ordering
CREATE VIEW v_export_selections AS
SELECT
	s.id               AS student_id,
	s.name             AS student_name,
	s.grade            AS grade,
	s.legal_sex        AS legal_sex,
	c.id               AS course_id,
	c.name             AS course_name,
	c.period_id        AS period,
	ch.invitation_type AS invitation_type
FROM choices ch
JOIN students s ON s.id = ch.student_id
JOIN courses  c ON c.id = ch.course_id
ORDER BY s.id, c.period_id, c.id;


-- Indexes

-- Partial unique indexes for session tokens (one active session per user)
CREATE UNIQUE INDEX ux_admins_session_token
	ON admins(session_token)
	WHERE session_token IS NOT NULL;

CREATE UNIQUE INDEX ux_students_session_token
	ON students(session_token)
	WHERE session_token IS NOT NULL;

-- Students filtering
CREATE INDEX idx_students_grade ON students(grade);
CREATE INDEX idx_students_legal_sex ON students(legal_sex);

-- Courses filtering
CREATE INDEX idx_courses_period ON courses(period_id);
CREATE INDEX idx_courses_category ON courses(category_id);
CREATE INDEX idx_courses_membership ON courses(membership);

-- Choices lookup and aggregation
CREATE INDEX idx_choices_course ON choices(course_id);
CREATE INDEX idx_choices_course_invtype ON choices(course_id, invitation_type);
CREATE INDEX idx_choices_student ON choices(student_id);

-- invitations admin views / scans
CREATE INDEX idx_choices_invites
	ON choices(invitation_type)
	WHERE invitation_type <> 'no';

-- allowed grades fast paths
CREATE INDEX idx_cag_grade ON course_allowed_grades(grade);
CREATE INDEX idx_cag_course ON course_allowed_grades(course_id);
