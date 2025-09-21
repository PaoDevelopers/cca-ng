package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

// Middleware to ensure only students can access student endpoints
func (app *App) studentOnly(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := app.authenticateRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if user.Role != "student" {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		handler(w, r)
	}
}

// Get available courses for the student with availability status
func (app *App) studentCoursesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := app.authenticateRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()

	// Get student info
	student, err := app.queries.GetStudentByID(ctx, user.ID)
	if err != nil {
		http.Error(w, "Failed to get student info", http.StatusInternalServerError)
		return
	}

	// Get available courses with availability status
	courses, err := app.queries.GetStudentAvailableCourses(ctx, GetStudentAvailableCoursesParams{
		StudentID: user.ID,
		Column2:   string(student.LegalSex),
		Grade:     student.Grade,
	})
	if err != nil {
		http.Error(w, "Failed to get courses", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courses)
}

// Get student's current selections
func (app *App) studentSelectionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := app.authenticateRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()

	selections, err := app.queries.GetStudentSelections(ctx, user.ID)
	if err != nil {
		http.Error(w, "Failed to get selections", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(selections)
}

// Get student's selections organized by period
func (app *App) studentSelectionsByPeriodHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := app.authenticateRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()

	selections, err := app.queries.GetStudentSelectionsByPeriod(ctx, user.ID)
	if err != nil {
		http.Error(w, "Failed to get selections by period", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(selections)
}

// Handle adding/removing course selections
func (app *App) studentSelectCourseHandler(w http.ResponseWriter, r *http.Request) {
	user, err := app.authenticateRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()
	courseID := r.FormValue("course_id")
	if courseID == "" {
		http.Error(w, "Course ID is required", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPost {
		// Add selection
		course, err := app.queries.GetCourseByID(ctx, courseID)
		if err != nil {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}

		err = app.queries.AddStudentSelection(ctx, AddStudentSelectionParams{
			StudentID: user.ID,
			CourseID:  courseID,
			PeriodID:  course.PeriodID,
		})
		if err != nil {
			// The database triggers will enforce restrictions and capacity
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		// Get updated enrollment count for the course
		enrollment, err := app.queries.GetCourseCurrentEnrollment(ctx, courseID)
		if err != nil {
			log.Printf("Warning: Failed to get enrollment count for course %s: %v", courseID, err)
			enrollment = 0 // fallback
		}

		// Broadcast fine-grained course enrollment update to all users
		app.BroadcastCourseEnrollmentUpdate(courseID, enrollment)

		// Broadcast selection action to admins
		app.BroadcastSelectionAction("added", user.ID, courseID, "no", user.Username)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success": true}`))

	} else if r.Method == http.MethodDelete {
		// Remove selection
		err = app.queries.RemoveStudentSelection(ctx, RemoveStudentSelectionParams{
			StudentID: user.ID,
			CourseID:  courseID,
		})
		if err != nil {
			http.Error(w, "Failed to remove selection", http.StatusInternalServerError)
			return
		}

		// Get updated enrollment count for the course
		enrollment, err := app.queries.GetCourseCurrentEnrollment(ctx, courseID)
		if err != nil {
			log.Printf("Warning: Failed to get enrollment count for course %s: %v", courseID, err)
			enrollment = 0 // fallback
		}

		// Broadcast fine-grained course enrollment update to all users
		app.BroadcastCourseEnrollmentUpdate(courseID, enrollment)

		// Broadcast selection action to admins
		app.BroadcastSelectionAction("removed", user.ID, courseID, "no", user.Username)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success": true}`))

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Get student's requirements compliance status
func (app *App) studentRequirementsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := app.authenticateRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()

	// Get total requirements status
	status, err := app.queries.GetStudentRequirementsStatus(ctx, user.ID)
	if err != nil {
		http.Error(w, "Failed to get requirements status", http.StatusInternalServerError)
		return
	}

	// Get requirement groups status
	groups, err := app.queries.GetStudentRequirementGroupsStatus(ctx, user.ID)
	if err != nil {
		http.Error(w, "Failed to get requirement groups status", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"total_requirements": status,
		"group_requirements": groups,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get student's selection status (whether they can select courses)
func (app *App) studentSelectionStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := app.authenticateRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()

	// Get student info to check their grade
	student, err := app.queries.GetStudentByID(ctx, user.ID)
	if err != nil {
		http.Error(w, "Failed to get student info", http.StatusInternalServerError)
		return
	}

	// Check if selection is enabled for this grade
	selectionControl, err := app.queries.GetGradeSelectionControl(ctx, student.Grade)
	if err != nil {
		// If no control exists, default to disabled
		response := map[string]bool{
			"enabled": false,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]bool{
		"enabled": selectionControl.Enabled,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
