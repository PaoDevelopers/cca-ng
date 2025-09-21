package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// adminOnly middleware to ensure only admin users can access endpoints
func (app *App) adminOnly(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := app.authenticateRequest(r)
		if err != nil || user.Role != "admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}

// Grades handlers
func (app *App) gradesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	switch r.Method {
	case http.MethodGet:
		grades, err := app.queries.GetAllGrades(ctx)
		if err != nil {
			http.Error(w, "Failed to fetch grades", http.StatusInternalServerError)
			return
		}

		// Ensure we return empty array instead of null when there are no grades
		if grades == nil {
			grades = []int64{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(grades)

	case http.MethodPost:
		gradeStr := r.FormValue("grade")
		if gradeStr == "" {
			http.Error(w, "Grade is required", http.StatusBadRequest)
			return
		}

		grade, err := strconv.ParseInt(gradeStr, 10, 64)
		if err != nil {
			http.Error(w, "Grade must be a valid number", http.StatusBadRequest)
			return
		}

		err = app.queries.CreateGrade(ctx, grade)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				http.Error(w, "Grade already exists", http.StatusConflict)
				return
			}
			http.Error(w, "Failed to create grade", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "grade": grade})

		// Broadcast grades invalidation to all clients
		app.BroadcastGradesInvalidation()

	case http.MethodDelete:
		gradeStr := r.URL.Query().Get("grade")
		if gradeStr == "" {
			http.Error(w, "Grade is required", http.StatusBadRequest)
			return
		}

		grade, err := strconv.ParseInt(gradeStr, 10, 64)
		if err != nil {
			http.Error(w, "Grade must be a valid number", http.StatusBadRequest)
			return
		}

		err = app.queries.DeleteGrade(ctx, grade)
		if err != nil {
			if strings.Contains(err.Error(), "foreign key") {
				http.Error(w, "Cannot delete grade: it is still in use", http.StatusConflict)
				return
			}
			http.Error(w, "Failed to delete grade", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})

		// Broadcast grades invalidation to all clients
		app.BroadcastGradesInvalidation()

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Periods handlers
func (app *App) periodsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	switch r.Method {
	case http.MethodGet:
		periods, err := app.queries.GetAllPeriods(ctx)
		if err != nil {
			http.Error(w, "Failed to fetch periods", http.StatusInternalServerError)
			return
		}

		// Ensure we return empty array instead of null when there are no periods
		if periods == nil {
			periods = []string{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(periods)

	case http.MethodPost:
		periodID := strings.TrimSpace(r.FormValue("id"))
		if periodID == "" {
			http.Error(w, "Period ID is required", http.StatusBadRequest)
			return
		}

		err := app.queries.CreatePeriod(ctx, periodID)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				http.Error(w, "Period already exists", http.StatusConflict)
				return
			}
			http.Error(w, "Failed to create period", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "id": periodID})

		// Broadcast periods invalidation to all clients
		app.BroadcastPeriodsInvalidation()

	case http.MethodDelete:
		periodID := r.URL.Query().Get("id")
		if periodID == "" {
			http.Error(w, "Period ID is required", http.StatusBadRequest)
			return
		}

		err := app.queries.DeletePeriod(ctx, periodID)
		if err != nil {
			if strings.Contains(err.Error(), "foreign key") {
				http.Error(w, "Cannot delete period: it is still in use", http.StatusConflict)
				return
			}
			http.Error(w, "Failed to delete period", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})

		// Broadcast periods invalidation to all clients
		app.BroadcastPeriodsInvalidation()

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Categories handlers
func (app *App) categoriesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	switch r.Method {
	case http.MethodGet:
		categories, err := app.queries.GetAllCategories(ctx)
		if err != nil {
			http.Error(w, "Failed to fetch categories", http.StatusInternalServerError)
			return
		}

		// Ensure we return empty array instead of null when there are no categories
		if categories == nil {
			categories = []string{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(categories)

	case http.MethodPost:
		categoryID := strings.TrimSpace(r.FormValue("id"))
		if categoryID == "" {
			http.Error(w, "Category ID is required", http.StatusBadRequest)
			return
		}

		err := app.queries.CreateCategory(ctx, categoryID)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				http.Error(w, "Category already exists", http.StatusConflict)
				return
			}
			http.Error(w, "Failed to create category", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "id": categoryID})

		// Broadcast categories invalidation to all clients
		app.BroadcastCategoriesInvalidation()

	case http.MethodDelete:
		categoryID := r.URL.Query().Get("id")
		if categoryID == "" {
			http.Error(w, "Category ID is required", http.StatusBadRequest)
			return
		}

		err := app.queries.DeleteCategory(ctx, categoryID)
		if err != nil {
			if strings.Contains(err.Error(), "foreign key") {
				http.Error(w, "Cannot delete category: it is still in use", http.StatusConflict)
				return
			}
			http.Error(w, "Failed to delete category", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})

		// Broadcast categories invalidation to all clients
		app.BroadcastCategoriesInvalidation()

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Courses handlers
func (app *App) coursesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	switch r.Method {
	case http.MethodGet:
		courses, err := app.queries.GetAllCourses(ctx)
		if err != nil {
			http.Error(w, "Failed to fetch courses", http.StatusInternalServerError)
			return
		}

		// Ensure we return empty array instead of null when there are no courses
		if courses == nil {
			courses = []Course{}
		}

		// Also get allowed grades for each course
		coursesWithGrades := make([]map[string]interface{}, 0, len(courses))
		for _, course := range courses {
			allowedGrades, err := app.queries.GetCourseAllowedGrades(ctx, course.ID)
			if err != nil {
				http.Error(w, "Failed to fetch course allowed grades", http.StatusInternalServerError)
				return
			}

			// Ensure allowedGrades is never nil, always an empty slice if no grades
			if allowedGrades == nil {
				allowedGrades = []int64{}
			}

			courseData := map[string]interface{}{
				"id":              course.ID,
				"name":            course.Name,
				"description":     course.Description,
				"period_id":       course.PeriodID,
				"max_students":    course.MaxStudents,
				"sex_restriction": course.SexRestriction,
				"membership":      course.Membership,
				"teacher":         course.Teacher,
				"location":        course.Location,
				"category_id":     course.CategoryID,
				"allowed_grades":  allowedGrades,
			}
			coursesWithGrades = append(coursesWithGrades, courseData)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(coursesWithGrades)

	case http.MethodPost:
		err := app.handleCreateOrUpdateCourse(ctx, w, r, false)
		if err != nil {
			return
		}
		// Broadcast courses invalidation to all clients
		app.BroadcastCoursesInvalidation()

	case http.MethodPut:
		err := app.handleCreateOrUpdateCourse(ctx, w, r, true)
		if err != nil {
			return
		}
		// Broadcast courses invalidation to all clients
		app.BroadcastCoursesInvalidation()

	case http.MethodDelete:
		courseID := r.URL.Query().Get("id")
		if courseID == "" {
			http.Error(w, "Course ID is required", http.StatusBadRequest)
			return
		}

		err := app.queries.DeleteCourse(ctx, courseID)
		if err != nil {
			if strings.Contains(err.Error(), "foreign key") {
				http.Error(w, "Cannot delete course: it is still in use", http.StatusConflict)
				return
			}
			http.Error(w, "Failed to delete course", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})

		// Broadcast courses invalidation to all clients
		app.BroadcastCoursesInvalidation()

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) handleCreateOrUpdateCourse(ctx context.Context, w http.ResponseWriter, r *http.Request, isUpdate bool) error {
	courseID := strings.TrimSpace(r.FormValue("id"))
	name := strings.TrimSpace(r.FormValue("name"))
	description := r.FormValue("description")
	periodID := strings.TrimSpace(r.FormValue("period_id"))
	categoryID := strings.TrimSpace(r.FormValue("category_id"))
	teacher := r.FormValue("teacher")
	location := r.FormValue("location")
	sexRestriction := r.FormValue("sex_restriction")
	membership := r.FormValue("membership")
	maxStudentsStr := r.FormValue("max_students")
	allowedGradesStr := r.FormValue("allowed_grades")

	// Validation
	if courseID == "" || name == "" || periodID == "" || categoryID == "" {
		http.Error(w, "Course ID, name, period_id, and category_id are required", http.StatusBadRequest)
		return fmt.Errorf("validation failed")
	}

	maxStudents, err := strconv.ParseInt(maxStudentsStr, 10, 64)
	if err != nil || maxStudents < 0 {
		http.Error(w, "max_students must be a non-negative number", http.StatusBadRequest)
		return fmt.Errorf("validation failed")
	}

	// Parse allowed grades
	var allowedGrades []int64
	if allowedGradesStr != "" {
		gradeStrs := strings.Split(allowedGradesStr, ",")
		allowedGrades = make([]int64, 0, len(gradeStrs))
		for _, gradeStr := range gradeStrs {
			grade, err := strconv.ParseInt(strings.TrimSpace(gradeStr), 10, 64)
			if err != nil {
				http.Error(w, "Invalid grade in allowed_grades", http.StatusBadRequest)
				return fmt.Errorf("validation failed")
			}
			allowedGrades = append(allowedGrades, grade)
		}
	}

	// Set defaults
	if sexRestriction == "" {
		sexRestriction = "ANY"
	}
	if membership == "" {
		membership = "free"
	}

	// Create or update course
	if isUpdate {
		err = app.queries.UpdateCourse(ctx, UpdateCourseParams{
			ID:             courseID,
			Name:           name,
			Description:    description,
			PeriodID:       periodID,
			MaxStudents:    maxStudents,
			SexRestriction: SexRestriction(sexRestriction),
			Membership:     MembershipType(membership),
			Teacher:        teacher,
			Location:       location,
			CategoryID:     categoryID,
		})
		if err != nil {
			http.Error(w, "Failed to update course", http.StatusInternalServerError)
			return err
		}
	} else {
		err = app.queries.CreateCourse(ctx, CreateCourseParams{
			ID:             courseID,
			Name:           name,
			Description:    description,
			PeriodID:       periodID,
			MaxStudents:    maxStudents,
			SexRestriction: SexRestriction(sexRestriction),
			Membership:     MembershipType(membership),
			Teacher:        teacher,
			Location:       location,
			CategoryID:     categoryID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				http.Error(w, "Course ID already exists", http.StatusConflict)
				return err
			}
			http.Error(w, "Failed to create course", http.StatusInternalServerError)
			return err
		}
	}

	// Update allowed grades
	err = app.queries.DeleteCourseAllowedGrades(ctx, courseID)
	if err != nil {
		http.Error(w, "Failed to update course allowed grades", http.StatusInternalServerError)
		return err
	}

	if len(allowedGrades) > 0 {
		gradeParams := make([]InsertCourseAllowedGradesParams, len(allowedGrades))
		for i, grade := range allowedGrades {
			gradeParams[i] = InsertCourseAllowedGradesParams{
				CourseID: courseID,
				Grade:    grade,
			}
		}
		_, err = app.queries.InsertCourseAllowedGrades(ctx, gradeParams)
		if err != nil {
			http.Error(w, "Failed to set course allowed grades", http.StatusInternalServerError)
			return err
		}
	}

	status := http.StatusCreated
	if isUpdate {
		status = http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "id": courseID})
	return nil
}

// Delete all courses
func (app *App) deleteAllCoursesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	err := app.queries.DeleteAllCourses(ctx)
	if err != nil {
		http.Error(w, "Failed to delete all courses", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})

	// Broadcast courses invalidation to all clients
	app.BroadcastCoursesInvalidation()
}

// CSV handlers for courses
func (app *App) coursesCSVExampleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"courses_example.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"id", "name", "description", "period_id", "max_students",
		"sex_restriction", "membership", "teacher", "location", "category_id", "allowed_grades",
	}
	writer.Write(header)

	// Write example row
	example := []string{
		"EXAMPLE1", "Example Course", "This is an example course description",
		"Period1", "20", "ANY", "free", "John Doe", "Room 101", "Sport", "9,10,11,12",
	}
	writer.Write(example)
}

func (app *App) coursesCSVDownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	courses, err := app.queries.GetAllCourses(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch courses", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"courses.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"id", "name", "description", "period_id", "max_students",
		"sex_restriction", "membership", "teacher", "location", "category_id", "allowed_grades",
	}
	writer.Write(header)

	// Write courses
	for _, course := range courses {
		allowedGrades, err := app.queries.GetCourseAllowedGrades(ctx, course.ID)
		if err != nil {
			http.Error(w, "Failed to fetch course allowed grades", http.StatusInternalServerError)
			return
		}

		var gradesStr string
		if len(allowedGrades) > 0 {
			gradeStrs := make([]string, len(allowedGrades))
			for i, grade := range allowedGrades {
				gradeStrs[i] = fmt.Sprintf("%d", grade)
			}
			gradesStr = strings.Join(gradeStrs, ",")
		}

		row := []string{
			course.ID,
			course.Name,
			course.Description,
			course.PeriodID,
			fmt.Sprintf("%d", course.MaxStudents),
			string(course.SexRestriction),
			string(course.Membership),
			course.Teacher,
			course.Location,
			course.CategoryID,
			gradesStr,
		}
		writer.Write(row)
	}
}

func (app *App) coursesCSVPreviewHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("coursesCSVPreviewHandler: start method=%s remote=%s", r.Method, r.RemoteAddr)
	if r.Method != http.MethodPost {
		log.Printf("coursesCSVPreviewHandler: method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("coursesCSVPreviewHandler: calling r.FormFile")
	file, _, err := r.FormFile("csv")
	if err != nil {
		log.Printf("coursesCSVPreviewHandler: failed to read uploaded file: %v", err)
		http.Error(w, "Failed to read uploaded file", http.StatusBadRequest)
		return
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Printf("coursesCSVPreviewHandler: warning: error closing file: %v", cerr)
		} else {
			log.Printf("coursesCSVPreviewHandler: file closed")
		}
	}()
	log.Printf("coursesCSVPreviewHandler: file opened successfully")

	reader := csv.NewReader(file)
	log.Printf("coursesCSVPreviewHandler: calling reader.ReadAll")
	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("coursesCSVPreviewHandler: ReadAll error: %v", err)
		http.Error(w, "Failed to parse CSV file: "+err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("coursesCSVPreviewHandler: ReadAll succeeded, records=%d", len(records))

	if len(records) == 0 {
		log.Printf("coursesCSVPreviewHandler: csv file empty")
		http.Error(w, "CSV file is empty", http.StatusBadRequest)
		return
	}

	// Validate header
	expectedHeader := []string{
		"id", "name", "description", "period_id", "max_students",
		"sex_restriction", "membership", "teacher", "location", "category_id", "allowed_grades",
	}

	header := records[0]
	log.Printf("coursesCSVPreviewHandler: header read: %v", header)
	if len(header) != len(expectedHeader) {
		log.Printf("coursesCSVPreviewHandler: header length mismatch: got=%d want=%d", len(header), len(expectedHeader))
		http.Error(w, fmt.Sprintf("CSV header must have %d columns", len(expectedHeader)), http.StatusBadRequest)
		return
	}

	for i, expected := range expectedHeader {
		if header[i] != expected {
			log.Printf("coursesCSVPreviewHandler: header column mismatch at index %d: got=%q expected=%q", i, header[i], expected)
			http.Error(w, fmt.Sprintf("Column %d should be '%s', got '%s'", i+1, expected, header[i]), http.StatusBadRequest)
			return
		}
	}
	log.Printf("coursesCSVPreviewHandler: header validated successfully")

	ctx := context.Background()
	log.Printf("coursesCSVPreviewHandler: created background context for validation")

	// Get existing periods, categories, and grades for validation
	log.Printf("coursesCSVPreviewHandler: fetching periods")
	periods, err := app.queries.GetAllPeriods(ctx)
	if err != nil {
		log.Printf("coursesCSVPreviewHandler: GetAllPeriods error: %v", err)
		http.Error(w, "Failed to fetch periods for validation", http.StatusInternalServerError)
		return
	}
	log.Printf("coursesCSVPreviewHandler: fetched %d periods", len(periods))
	periodMap := make(map[string]bool)
	for _, period := range periods {
		periodMap[period] = true
	}

	log.Printf("coursesCSVPreviewHandler: fetching categories")
	categories, err := app.queries.GetAllCategories(ctx)
	if err != nil {
		log.Printf("coursesCSVPreviewHandler: GetAllCategories error: %v", err)
		http.Error(w, "Failed to fetch categories for validation", http.StatusInternalServerError)
		return
	}
	log.Printf("coursesCSVPreviewHandler: fetched %d categories", len(categories))
	categoryMap := make(map[string]bool)
	for _, category := range categories {
		categoryMap[category] = true
	}

	log.Printf("coursesCSVPreviewHandler: fetching grades")
	grades, err := app.queries.GetAllGrades(ctx)
	if err != nil {
		log.Printf("coursesCSVPreviewHandler: GetAllGrades error: %v", err)
		http.Error(w, "Failed to fetch grades for validation", http.StatusInternalServerError)
		return
	}
	log.Printf("coursesCSVPreviewHandler: fetched %d grades", len(grades))
	gradeMap := make(map[int64]bool)
	for _, grade := range grades {
		gradeMap[grade] = true
	}

	// Preview data structure
	type PreviewRow struct {
		RowNumber  int                    `json:"row_number"`
		Data       map[string]interface{} `json:"data"`
		Errors     []string               `json:"errors"`
		IsValid    bool                   `json:"is_valid"`
		WillUpdate bool                   `json:"will_update"`
	}

	preview := make([]PreviewRow, 0, len(records)-1)
	var hasErrors bool

	log.Printf("coursesCSVPreviewHandler: starting to process %d data rows", len(records)-1)

	// Process each row
	for rowIdx, record := range records[1:] {
		rowNumber := rowIdx + 2
		log.Printf("coursesCSVPreviewHandler: processing row %d", rowNumber)
		errors := []string{}
		data := make(map[string]interface{})

		if len(record) != len(expectedHeader) {
			errMsg := fmt.Sprintf("Has %d columns, expected %d", len(record), len(expectedHeader))
			log.Printf("coursesCSVPreviewHandler: row %d column count mismatch: %s", rowNumber, errMsg)
			errors = append(errors, errMsg)
		} else {
			courseID := strings.TrimSpace(record[0])
			name := strings.TrimSpace(record[1])
			description := record[2]
			periodID := strings.TrimSpace(record[3])
			maxStudentsStr := strings.TrimSpace(record[4])
			sexRestriction := strings.TrimSpace(record[5])
			membership := strings.TrimSpace(record[6])
			teacher := record[7]
			location := record[8]
			categoryID := strings.TrimSpace(record[9])
			allowedGradesStr := strings.TrimSpace(record[10])

			// Store data
			data["id"] = courseID
			data["name"] = name
			data["description"] = description
			data["period_id"] = periodID
			data["max_students"] = maxStudentsStr
			data["sex_restriction"] = sexRestriction
			data["membership"] = membership
			data["teacher"] = teacher
			data["location"] = location
			data["category_id"] = categoryID
			data["allowed_grades"] = allowedGradesStr

			// Validation
			if courseID == "" {
				log.Printf("coursesCSVPreviewHandler: row %d validation: Course ID is required", rowNumber)
				errors = append(errors, "Course ID is required")
			}
			if name == "" {
				log.Printf("coursesCSVPreviewHandler: row %d validation: Name is required", rowNumber)
				errors = append(errors, "Name is required")
			}
			if periodID == "" {
				log.Printf("coursesCSVPreviewHandler: row %d validation: Period ID is required", rowNumber)
				errors = append(errors, "Period ID is required")
			} else if !periodMap[periodID] {
				log.Printf("coursesCSVPreviewHandler: row %d validation: Period '%s' does not exist", rowNumber, periodID)
				errors = append(errors, fmt.Sprintf("Period '%s' does not exist", periodID))
			}
			if categoryID == "" {
				log.Printf("coursesCSVPreviewHandler: row %d validation: Category ID is required", rowNumber)
				errors = append(errors, "Category ID is required")
			} else if !categoryMap[categoryID] {
				log.Printf("coursesCSVPreviewHandler: row %d validation: Category '%s' does not exist", rowNumber, categoryID)
				errors = append(errors, fmt.Sprintf("Category '%s' does not exist", categoryID))
			}

			maxStudents, err := strconv.ParseInt(maxStudentsStr, 10, 64)
			if err != nil || maxStudents < 0 {
				log.Printf("coursesCSVPreviewHandler: row %d validation: invalid max_students %q (err=%v)", rowNumber, maxStudentsStr, err)
				errors = append(errors, "Max students must be a non-negative number")
			} else {
				data["max_students"] = maxStudents
			}

			// Validate sex restriction
			if sexRestriction != "" && sexRestriction != "ANY" && sexRestriction != "F" && sexRestriction != "M" && sexRestriction != "X" {
				log.Printf("coursesCSVPreviewHandler: row %d validation: invalid sex_restriction %q", rowNumber, sexRestriction)
				errors = append(errors, "Sex restriction must be empty, ANY, F, M, or X")
			}
			if sexRestriction == "" {
				data["sex_restriction"] = "ANY"
			}

			// Validate membership
			if membership != "" && membership != "free" && membership != "invite_only" {
				log.Printf("coursesCSVPreviewHandler: row %d validation: invalid membership %q", rowNumber, membership)
				errors = append(errors, "Membership must be empty, free, or invite_only")
			}
			if membership == "" {
				data["membership"] = "free"
			}

			// Parse and validate allowed grades
			var allowedGrades []int64
			if allowedGradesStr != "" {
				gradeStrs := strings.Split(allowedGradesStr, ",")
				allowedGrades = make([]int64, 0, len(gradeStrs))
				for _, gradeStr := range gradeStrs {
					grade, err := strconv.ParseInt(strings.TrimSpace(gradeStr), 10, 64)
					if err != nil {
						log.Printf("coursesCSVPreviewHandler: row %d validation: invalid grade %q in allowed_grades", rowNumber, gradeStr)
						errors = append(errors, fmt.Sprintf("Invalid grade '%s' in allowed_grades", gradeStr))
					} else if !gradeMap[grade] {
						log.Printf("coursesCSVPreviewHandler: row %d validation: grade %d does not exist", rowNumber, grade)
						errors = append(errors, fmt.Sprintf("Grade %d does not exist", grade))
					} else {
						allowedGrades = append(allowedGrades, grade)
					}
				}
				data["allowed_grades"] = allowedGrades
			} else {
				data["allowed_grades"] = []int64{}
			}

			// Check if course exists
			willUpdate := false
			if courseID != "" {
				_, err := app.queries.GetCourseByID(ctx, courseID)
				willUpdate = (err == nil)
				log.Printf("coursesCSVPreviewHandler: row %d GetCourseByID(courseID=%s) err=%v willUpdate=%v", rowNumber, courseID, err, willUpdate)
			}
			data["will_update"] = willUpdate
		}

		isValid := len(errors) == 0
		if !isValid {
			hasErrors = true
			log.Printf("coursesCSVPreviewHandler: row %d has errors: %v", rowNumber, errors)
		} else {
			log.Printf("coursesCSVPreviewHandler: row %d is valid", rowNumber)
		}

		preview = append(preview, PreviewRow{
			RowNumber:  rowNumber,
			Data:       data,
			Errors:     errors,
			IsValid:    isValid,
			WillUpdate: data["will_update"] == true,
		})
	}

	log.Printf("coursesCSVPreviewHandler: finished processing rows: preview_count=%d hasErrors=%v", len(preview), hasErrors)

	w.Header().Set("Content-Type", "application/json")
	log.Printf("coursesCSVPreviewHandler: encoding response JSON")
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    !hasErrors,
		"has_errors": hasErrors,
		"preview":    preview,
		"total_rows": len(records) - 1,
	})
	if err != nil {
		// Log the encoding error (we can't respond again if headers/body already sent, but log it)
		log.Printf("coursesCSVPreviewHandler: error encoding response JSON: %v", err)
	} else {
		log.Printf("coursesCSVPreviewHandler: response encoded successfully")
	}
}

func (app *App) coursesCSVUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("csv")
	if err != nil {
		http.Error(w, "Failed to read uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(w, "Failed to parse CSV file: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(records) == 0 {
		http.Error(w, "CSV file is empty", http.StatusBadRequest)
		return
	}

	// Validate header
	expectedHeader := []string{
		"id", "name", "description", "period_id", "max_students",
		"sex_restriction", "membership", "teacher", "location", "category_id", "allowed_grades",
	}

	header := records[0]
	if len(header) != len(expectedHeader) {
		http.Error(w, fmt.Sprintf("CSV header must have %d columns", len(expectedHeader)), http.StatusBadRequest)
		return
	}

	for i, expected := range expectedHeader {
		if header[i] != expected {
			http.Error(w, fmt.Sprintf("Column %d should be '%s', got '%s'", i+1, expected, header[i]), http.StatusBadRequest)
			return
		}
	}

	ctx := context.Background()

	// Validate and process each row
	for rowIdx, record := range records[1:] {
		if len(record) != len(expectedHeader) {
			http.Error(w, fmt.Sprintf("Row %d has %d columns, expected %d", rowIdx+2, len(record), len(expectedHeader)), http.StatusBadRequest)
			return
		}

		courseID := strings.TrimSpace(record[0])
		name := strings.TrimSpace(record[1])
		description := record[2]
		periodID := strings.TrimSpace(record[3])
		maxStudentsStr := strings.TrimSpace(record[4])
		sexRestriction := strings.TrimSpace(record[5])
		membership := strings.TrimSpace(record[6])
		teacher := record[7]
		location := record[8]
		categoryID := strings.TrimSpace(record[9])
		allowedGradesStr := strings.TrimSpace(record[10])

		// Validation
		if courseID == "" || name == "" || periodID == "" || categoryID == "" {
			http.Error(w, fmt.Sprintf("Row %d: Course ID, name, period_id, and category_id are required", rowIdx+2), http.StatusBadRequest)
			return
		}

		maxStudents, err := strconv.ParseInt(maxStudentsStr, 10, 64)
		if err != nil || maxStudents < 0 {
			http.Error(w, fmt.Sprintf("Row %d: max_students must be a non-negative number", rowIdx+2), http.StatusBadRequest)
			return
		}

		// Parse allowed grades
		var allowedGrades []int64
		if allowedGradesStr != "" {
			gradeStrs := strings.Split(allowedGradesStr, ",")
			allowedGrades = make([]int64, 0, len(gradeStrs))
			for _, gradeStr := range gradeStrs {
				grade, err := strconv.ParseInt(strings.TrimSpace(gradeStr), 10, 64)
				if err != nil {
					http.Error(w, fmt.Sprintf("Row %d: Invalid grade '%s' in allowed_grades", rowIdx+2, gradeStr), http.StatusBadRequest)
					return
				}
				allowedGrades = append(allowedGrades, grade)
			}
		}

		// Set defaults
		if sexRestriction == "" {
			sexRestriction = "ANY"
		}
		if membership == "" {
			membership = "free"
		}

		// Check if course already exists
		_, err = app.queries.GetCourseByID(ctx, courseID)
		exists := err == nil

		if exists {
			// Update existing course
			err = app.queries.UpdateCourse(ctx, UpdateCourseParams{
				ID:             courseID,
				Name:           name,
				Description:    description,
				PeriodID:       periodID,
				MaxStudents:    maxStudents,
				SexRestriction: SexRestriction(sexRestriction),
				Membership:     MembershipType(membership),
				Teacher:        teacher,
				Location:       location,
				CategoryID:     categoryID,
			})
			if err != nil {
				http.Error(w, fmt.Sprintf("Row %d: Failed to update course: %v", rowIdx+2, err), http.StatusInternalServerError)
				return
			}
		} else {
			// Create new course
			err = app.queries.CreateCourse(ctx, CreateCourseParams{
				ID:             courseID,
				Name:           name,
				Description:    description,
				PeriodID:       periodID,
				MaxStudents:    maxStudents,
				SexRestriction: SexRestriction(sexRestriction),
				Membership:     MembershipType(membership),
				Teacher:        teacher,
				Location:       location,
				CategoryID:     categoryID,
			})
			if err != nil {
				http.Error(w, fmt.Sprintf("Row %d: Failed to create course: %v", rowIdx+2, err), http.StatusInternalServerError)
				return
			}
		}

		// Update allowed grades
		err = app.queries.DeleteCourseAllowedGrades(ctx, courseID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Row %d: Failed to update course allowed grades: %v", rowIdx+2, err), http.StatusInternalServerError)
			return
		}

		if len(allowedGrades) > 0 {
			gradeParams := make([]InsertCourseAllowedGradesParams, len(allowedGrades))
			for i, grade := range allowedGrades {
				gradeParams[i] = InsertCourseAllowedGradesParams{
					CourseID: courseID,
					Grade:    grade,
				}
			}
			_, err = app.queries.InsertCourseAllowedGrades(ctx, gradeParams)
			if err != nil {
				http.Error(w, fmt.Sprintf("Row %d: Failed to set course allowed grades: %v", rowIdx+2, err), http.StatusInternalServerError)
				return
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"imported": len(records) - 1,
	})

	// Broadcast courses invalidation to all clients
	app.BroadcastCoursesInvalidation()
}

// Students handlers
func (app *App) studentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	switch r.Method {
	case http.MethodGet:
		students, err := app.queries.GetAllStudents(ctx)
		if err != nil {
			http.Error(w, "Failed to fetch students", http.StatusInternalServerError)
			return
		}

		// Ensure we return empty array instead of null when there are no students
		if students == nil {
			students = []GetAllStudentsRow{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(students)

	case http.MethodPost:
		err := app.handleCreateStudent(ctx, w, r)
		if err != nil {
			return
		}
		app.BroadcastStudentsInvalidation()

	case http.MethodPut:
		err := app.handleUpdateStudent(ctx, w, r)
		if err != nil {
			return
		}
		app.BroadcastStudentsInvalidation()

	case http.MethodDelete:
		studentIDStr := r.URL.Query().Get("id")
		if studentIDStr == "" {
			http.Error(w, "Student ID is required", http.StatusBadRequest)
			return
		}

		studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Student ID must be a valid number", http.StatusBadRequest)
			return
		}

		err = app.queries.DeleteStudent(ctx, studentID)
		if err != nil {
			if strings.Contains(err.Error(), "foreign key") {
				http.Error(w, "Cannot delete student: they have existing selections", http.StatusConflict)
				return
			}
			http.Error(w, "Failed to delete student", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
		app.BroadcastStudentsInvalidation()

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) handleCreateStudent(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	studentIDStr := strings.TrimSpace(r.FormValue("id"))
	name := strings.TrimSpace(r.FormValue("name"))
	gradeStr := r.FormValue("grade")
	legalSex := r.FormValue("legal_sex")
	password := r.FormValue("password")

	// Validation
	if studentIDStr == "" || name == "" || gradeStr == "" || legalSex == "" || password == "" {
		http.Error(w, "Student ID, name, grade, legal_sex, and password are required", http.StatusBadRequest)
		return fmt.Errorf("validation failed")
	}

	studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Student ID must be a valid number", http.StatusBadRequest)
		return fmt.Errorf("validation failed")
	}

	grade, err := strconv.ParseInt(gradeStr, 10, 64)
	if err != nil {
		http.Error(w, "Grade must be a valid number", http.StatusBadRequest)
		return fmt.Errorf("validation failed")
	}

	// Validate legal_sex enum
	if legalSex != "F" && legalSex != "M" && legalSex != "X" {
		http.Error(w, "Legal sex must be F, M, or X", http.StatusBadRequest)
		return fmt.Errorf("validation failed")
	}

	// Hash password
	hashedPassword, err := app.hashPassword(password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return err
	}

	// Create student
	student, err := app.queries.CreateStudent(ctx, CreateStudentParams{
		ID:           studentID,
		Name:         name,
		Grade:        grade,
		LegalSex:     LegalSex(legalSex),
		PasswordHash: hashedPassword,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			http.Error(w, "Student ID already exists", http.StatusConflict)
			return err
		}
		if strings.Contains(err.Error(), "foreign key") {
			http.Error(w, "Invalid grade - grade does not exist", http.StatusBadRequest)
			return err
		}
		http.Error(w, "Failed to create student", http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"student": student,
	})
	return nil
}

func (app *App) handleUpdateStudent(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	studentIDStr := strings.TrimSpace(r.FormValue("id"))
	name := strings.TrimSpace(r.FormValue("name"))
	gradeStr := r.FormValue("grade")
	legalSex := r.FormValue("legal_sex")

	// Validation
	if studentIDStr == "" || name == "" || gradeStr == "" || legalSex == "" {
		http.Error(w, "Student ID, name, grade, and legal_sex are required", http.StatusBadRequest)
		return fmt.Errorf("validation failed")
	}

	studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Student ID must be a valid number", http.StatusBadRequest)
		return fmt.Errorf("validation failed")
	}

	grade, err := strconv.ParseInt(gradeStr, 10, 64)
	if err != nil {
		http.Error(w, "Grade must be a valid number", http.StatusBadRequest)
		return fmt.Errorf("validation failed")
	}

	// Validate legal_sex enum
	if legalSex != "F" && legalSex != "M" && legalSex != "X" {
		http.Error(w, "Legal sex must be F, M, or X", http.StatusBadRequest)
		return fmt.Errorf("validation failed")
	}

	// Update student
	err = app.queries.UpdateStudent(ctx, UpdateStudentParams{
		ID:       studentID,
		Name:     name,
		Grade:    grade,
		LegalSex: LegalSex(legalSex),
	})
	if err != nil {
		if strings.Contains(err.Error(), "foreign key") {
			http.Error(w, "Invalid grade - grade does not exist", http.StatusBadRequest)
			return err
		}
		http.Error(w, "Failed to update student", http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
	return nil
}

// Delete all students
func (app *App) deleteAllStudentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	err := app.queries.DeleteAllStudents(ctx)
	if err != nil {
		http.Error(w, "Failed to delete all students", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
	app.BroadcastStudentsInvalidation()
}

// CSV handlers for students
func (app *App) studentsCSVExampleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"students_example.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{"id", "name", "grade", "legal_sex", "password"}
	writer.Write(header)

	// Write example row
	example := []string{"12345", "John Doe", "9", "M", "defaultpassword"}
	writer.Write(example)
}

func (app *App) studentsCSVDownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	students, err := app.queries.GetAllStudents(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch students", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"students.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{"id", "name", "grade", "legal_sex", "password"}
	writer.Write(header)

	// Write students (password field will be empty for security)
	for _, student := range students {
		row := []string{
			fmt.Sprintf("%d", student.ID),
			student.Name,
			fmt.Sprintf("%d", student.Grade),
			string(student.LegalSex),
			"", // Don't export passwords
		}
		writer.Write(row)
	}
}

func (app *App) studentsCSVPreviewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("csv")
	if err != nil {
		http.Error(w, "Failed to read uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(w, "Failed to parse CSV file: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(records) == 0 {
		http.Error(w, "CSV file is empty", http.StatusBadRequest)
		return
	}

	// Validate header
	expectedHeader := []string{"id", "name", "grade", "legal_sex", "password"}
	header := records[0]
	if len(header) != len(expectedHeader) {
		http.Error(w, fmt.Sprintf("CSV header must have %d columns", len(expectedHeader)), http.StatusBadRequest)
		return
	}

	for i, expected := range expectedHeader {
		if header[i] != expected {
			http.Error(w, fmt.Sprintf("Column %d should be '%s', got '%s'", i+1, expected, header[i]), http.StatusBadRequest)
			return
		}
	}

	ctx := context.Background()

	// Get existing grades for validation
	grades, err := app.queries.GetAllGrades(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch grades for validation", http.StatusInternalServerError)
		return
	}
	gradeMap := make(map[int64]bool)
	for _, grade := range grades {
		gradeMap[grade] = true
	}

	// Preview data structure
	type PreviewRow struct {
		RowNumber  int                    `json:"row_number"`
		Data       map[string]interface{} `json:"data"`
		Errors     []string               `json:"errors"`
		IsValid    bool                   `json:"is_valid"`
		WillUpdate bool                   `json:"will_update"`
	}

	preview := make([]PreviewRow, 0, len(records)-1)
	var hasErrors bool

	// Process each row
	for rowIdx, record := range records[1:] {
		rowNumber := rowIdx + 2
		errors := []string{}
		data := make(map[string]interface{})

		if len(record) != len(expectedHeader) {
			errors = append(errors, fmt.Sprintf("Has %d columns, expected %d", len(record), len(expectedHeader)))
		} else {
			studentIDStr := strings.TrimSpace(record[0])
			name := strings.TrimSpace(record[1])
			gradeStr := strings.TrimSpace(record[2])
			legalSex := strings.TrimSpace(record[3])
			password := strings.TrimSpace(record[4])

			// Store data
			data["id"] = studentIDStr
			data["name"] = name
			data["grade"] = gradeStr
			data["legal_sex"] = legalSex
			data["password"] = password

			// Validation
			if studentIDStr == "" {
				errors = append(errors, "Student ID is required")
			} else {
				studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
				if err != nil {
					errors = append(errors, "Student ID must be a valid number")
				} else {
					data["id"] = studentID
				}
			}

			if name == "" {
				errors = append(errors, "Name is required")
			}

			if gradeStr == "" {
				errors = append(errors, "Grade is required")
			} else {
				grade, err := strconv.ParseInt(gradeStr, 10, 64)
				if err != nil {
					errors = append(errors, "Grade must be a valid number")
				} else if !gradeMap[grade] {
					errors = append(errors, fmt.Sprintf("Grade %d does not exist", grade))
				} else {
					data["grade"] = grade
				}
			}

			if legalSex != "F" && legalSex != "M" && legalSex != "X" {
				errors = append(errors, "Legal sex must be F, M, or X")
			}

			if password == "" {
				errors = append(errors, "Password is required")
			}

			// Check if student exists
			willUpdate := false
			if studentIDStr != "" {
				studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
				if err == nil {
					_, err := app.queries.GetStudentByID(ctx, studentID)
					willUpdate = (err == nil)
				}
			}
			data["will_update"] = willUpdate
		}

		isValid := len(errors) == 0
		if !isValid {
			hasErrors = true
		}

		preview = append(preview, PreviewRow{
			RowNumber:  rowNumber,
			Data:       data,
			Errors:     errors,
			IsValid:    isValid,
			WillUpdate: data["will_update"] == true,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    !hasErrors,
		"has_errors": hasErrors,
		"preview":    preview,
		"total_rows": len(records) - 1,
	})
}

func (app *App) studentsCSVUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("csv")
	if err != nil {
		http.Error(w, "Failed to read uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(w, "Failed to parse CSV file: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(records) == 0 {
		http.Error(w, "CSV file is empty", http.StatusBadRequest)
		return
	}

	// Validate header
	expectedHeader := []string{"id", "name", "grade", "legal_sex", "password"}
	header := records[0]
	if len(header) != len(expectedHeader) {
		http.Error(w, fmt.Sprintf("CSV header must have %d columns", len(expectedHeader)), http.StatusBadRequest)
		return
	}

	for i, expected := range expectedHeader {
		if header[i] != expected {
			http.Error(w, fmt.Sprintf("Column %d should be '%s', got '%s'", i+1, expected, header[i]), http.StatusBadRequest)
			return
		}
	}

	ctx := context.Background()

	// Validate and process each row
	for rowIdx, record := range records[1:] {
		if len(record) != len(expectedHeader) {
			http.Error(w, fmt.Sprintf("Row %d has %d columns, expected %d", rowIdx+2, len(record), len(expectedHeader)), http.StatusBadRequest)
			return
		}

		studentIDStr := strings.TrimSpace(record[0])
		name := strings.TrimSpace(record[1])
		gradeStr := strings.TrimSpace(record[2])
		legalSex := strings.TrimSpace(record[3])
		password := strings.TrimSpace(record[4])

		// Validation
		if studentIDStr == "" || name == "" || gradeStr == "" || legalSex == "" || password == "" {
			http.Error(w, fmt.Sprintf("Row %d: All fields are required", rowIdx+2), http.StatusBadRequest)
			return
		}

		studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Row %d: Student ID must be a valid number", rowIdx+2), http.StatusBadRequest)
			return
		}

		grade, err := strconv.ParseInt(gradeStr, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Row %d: Grade must be a valid number", rowIdx+2), http.StatusBadRequest)
			return
		}

		if legalSex != "F" && legalSex != "M" && legalSex != "X" {
			http.Error(w, fmt.Sprintf("Row %d: Legal sex must be F, M, or X", rowIdx+2), http.StatusBadRequest)
			return
		}

		// Check if student already exists
		_, err = app.queries.GetStudentByID(ctx, studentID)
		exists := err == nil

		if exists {
			// Update existing student (but not password)
			err = app.queries.UpdateStudent(ctx, UpdateStudentParams{
				ID:       studentID,
				Name:     name,
				Grade:    grade,
				LegalSex: LegalSex(legalSex),
			})
			if err != nil {
				http.Error(w, fmt.Sprintf("Row %d: Failed to update student: %v", rowIdx+2, err), http.StatusInternalServerError)
				return
			}
		} else {
			// Create new student
			hashedPassword, err := app.hashPassword(password)
			if err != nil {
				http.Error(w, fmt.Sprintf("Row %d: Failed to hash password", rowIdx+2), http.StatusInternalServerError)
				return
			}

			_, err = app.queries.CreateStudent(ctx, CreateStudentParams{
				ID:           studentID,
				Name:         name,
				Grade:        grade,
				LegalSex:     LegalSex(legalSex),
				PasswordHash: hashedPassword,
			})
			if err != nil {
				http.Error(w, fmt.Sprintf("Row %d: Failed to create student: %v", rowIdx+2, err), http.StatusInternalServerError)
				return
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"imported": len(records) - 1,
	})

	app.BroadcastStudentsInvalidation()
}

// Requirements handlers
func (app *App) requirementsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	switch r.Method {
	case http.MethodGet:
		// Get grade from query parameter (optional)
		gradeStr := r.URL.Query().Get("grade")

		if gradeStr != "" {
			// Get requirements for specific grade
			grade, err := strconv.ParseInt(gradeStr, 10, 64)
			if err != nil {
				http.Error(w, "Grade must be a valid number", http.StatusBadRequest)
				return
			}

			// Get grade requirement
			gradeReq, err := app.queries.GetGradeRequirement(ctx, grade)
			if err != nil {
				// If not found, return default of 0
				gradeReq = GradeRequirement{Grade: grade, MinTotal: 0}
			}

			// Get requirement groups for this grade
			groups, err := app.queries.GetGradeRequirementGroups(ctx, grade)
			if err != nil {
				http.Error(w, "Failed to fetch requirement groups", http.StatusInternalServerError)
				return
			}

			// Get categories for each group
			groupsWithCategories := make([]map[string]interface{}, 0, len(groups))
			for _, group := range groups {
				categories, err := app.queries.GetGradeRequirementGroupCategories(ctx, group.ID)
				if err != nil {
					http.Error(w, "Failed to fetch group categories", http.StatusInternalServerError)
					return
				}

				groupData := map[string]interface{}{
					"id":         group.ID,
					"label":      group.Label,
					"min_count":  group.MinCount,
					"categories": categories,
				}
				groupsWithCategories = append(groupsWithCategories, groupData)
			}

			response := map[string]interface{}{
				"grade":     grade,
				"min_total": gradeReq.MinTotal,
				"groups":    groupsWithCategories,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			// Get all grade requirements
			gradeReqs, err := app.queries.GetAllGradeRequirements(ctx)
			if err != nil {
				http.Error(w, "Failed to fetch grade requirements", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(gradeReqs)
		}

	case http.MethodPost:
		// Set grade requirement
		gradeStr := r.FormValue("grade")
		minTotalStr := r.FormValue("min_total")

		if gradeStr == "" || minTotalStr == "" {
			http.Error(w, "Grade and min_total are required", http.StatusBadRequest)
			return
		}

		grade, err := strconv.ParseInt(gradeStr, 10, 64)
		if err != nil {
			http.Error(w, "Grade must be a valid number", http.StatusBadRequest)
			return
		}

		minTotal, err := strconv.ParseInt(minTotalStr, 10, 64)
		if err != nil || minTotal < 0 {
			http.Error(w, "min_total must be a non-negative number", http.StatusBadRequest)
			return
		}

		err = app.queries.SetGradeRequirement(ctx, SetGradeRequirementParams{
			Grade:    grade,
			MinTotal: minTotal,
		})
		if err != nil {
			http.Error(w, "Failed to set grade requirement", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
		app.BroadcastRequirementsInvalidation()

	case http.MethodDelete:
		gradeStr := r.URL.Query().Get("grade")
		if gradeStr == "" {
			http.Error(w, "Grade is required", http.StatusBadRequest)
			return
		}

		grade, err := strconv.ParseInt(gradeStr, 10, 64)
		if err != nil {
			http.Error(w, "Grade must be a valid number", http.StatusBadRequest)
			return
		}

		err = app.queries.DeleteGradeRequirement(ctx, grade)
		if err != nil {
			http.Error(w, "Failed to delete grade requirement", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
		app.BroadcastRequirementsInvalidation()

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Requirement groups handlers
func (app *App) requirementGroupsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	switch r.Method {
	case http.MethodPost:
		gradeStr := r.FormValue("grade")
		label := r.FormValue("label")
		minCountStr := r.FormValue("min_count")
		categoriesStr := r.FormValue("categories") // comma-separated

		if gradeStr == "" || label == "" || minCountStr == "" {
			http.Error(w, "Grade, label, and min_count are required", http.StatusBadRequest)
			return
		}

		grade, err := strconv.ParseInt(gradeStr, 10, 64)
		if err != nil {
			http.Error(w, "Grade must be a valid number", http.StatusBadRequest)
			return
		}

		minCount, err := strconv.ParseInt(minCountStr, 10, 64)
		if err != nil || minCount < 0 {
			http.Error(w, "min_count must be a non-negative number", http.StatusBadRequest)
			return
		}

		// Parse categories
		var categories []string
		if categoriesStr != "" {
			categories = strings.Split(categoriesStr, ",")
			for i, cat := range categories {
				categories[i] = strings.TrimSpace(cat)
			}
		}

		// Create the group
		groupID, err := app.queries.CreateGradeRequirementGroup(ctx, CreateGradeRequirementGroupParams{
			Grade:    grade,
			Label:    label,
			MinCount: minCount,
		})
		if err != nil {
			http.Error(w, "Failed to create requirement group", http.StatusInternalServerError)
			return
		}

		// Add categories to the group
		if len(categories) > 0 {
			categoryParams := make([]InsertGradeRequirementGroupCategoriesParams, len(categories))
			for i, category := range categories {
				categoryParams[i] = InsertGradeRequirementGroupCategoriesParams{
					ReqGroupID: groupID,
					CategoryID: category,
				}
			}
			_, err = app.queries.InsertGradeRequirementGroupCategories(ctx, categoryParams)
			if err != nil {
				http.Error(w, "Failed to set group categories", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"group_id": groupID,
		})
		app.BroadcastRequirementsInvalidation()

	case http.MethodPut:
		groupIDStr := r.FormValue("id")
		label := r.FormValue("label")
		minCountStr := r.FormValue("min_count")
		categoriesStr := r.FormValue("categories") // comma-separated

		if groupIDStr == "" || label == "" || minCountStr == "" {
			http.Error(w, "Group ID, label, and min_count are required", http.StatusBadRequest)
			return
		}

		groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Group ID must be a valid number", http.StatusBadRequest)
			return
		}

		minCount, err := strconv.ParseInt(minCountStr, 10, 64)
		if err != nil || minCount < 0 {
			http.Error(w, "min_count must be a non-negative number", http.StatusBadRequest)
			return
		}

		// Parse categories
		var categories []string
		if categoriesStr != "" {
			categories = strings.Split(categoriesStr, ",")
			for i, cat := range categories {
				categories[i] = strings.TrimSpace(cat)
			}
		}

		// Update the group
		err = app.queries.UpdateGradeRequirementGroup(ctx, UpdateGradeRequirementGroupParams{
			ID:       groupID,
			Label:    label,
			MinCount: minCount,
		})
		if err != nil {
			http.Error(w, "Failed to update requirement group", http.StatusInternalServerError)
			return
		}

		// Update categories for the group
		err = app.queries.DeleteGradeRequirementGroupCategories(ctx, groupID)
		if err != nil {
			http.Error(w, "Failed to update group categories", http.StatusInternalServerError)
			return
		}

		if len(categories) > 0 {
			categoryParams := make([]InsertGradeRequirementGroupCategoriesParams, len(categories))
			for i, category := range categories {
				categoryParams[i] = InsertGradeRequirementGroupCategoriesParams{
					ReqGroupID: groupID,
					CategoryID: category,
				}
			}
			_, err = app.queries.InsertGradeRequirementGroupCategories(ctx, categoryParams)
			if err != nil {
				http.Error(w, "Failed to set group categories", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
		app.BroadcastRequirementsInvalidation()

	case http.MethodDelete:
		groupIDStr := r.URL.Query().Get("id")
		if groupIDStr == "" {
			http.Error(w, "Group ID is required", http.StatusBadRequest)
			return
		}

		groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Group ID must be a valid number", http.StatusBadRequest)
			return
		}

		err = app.queries.DeleteGradeRequirementGroup(ctx, groupID)
		if err != nil {
			http.Error(w, "Failed to delete requirement group", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
		app.BroadcastRequirementsInvalidation()

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Invitations handlers
func (app *App) invitationsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	switch r.Method {
	case http.MethodGet:
		invitations, err := app.queries.GetAllInvitations(ctx)
		if err != nil {
			http.Error(w, "Failed to fetch invitations", http.StatusInternalServerError)
			return
		}

		// Ensure we return empty array instead of null when there are no invitations
		if invitations == nil {
			invitations = []GetAllInvitationsRow{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(invitations)

	case http.MethodPost:
		err := app.handleCreateOrUpdateInvitation(ctx, w, r, false)
		if err != nil {
			return
		}
		app.BroadcastInvitationsInvalidation()

	case http.MethodPut:
		err := app.handleCreateOrUpdateInvitation(ctx, w, r, true)
		if err != nil {
			return
		}
		app.BroadcastInvitationsInvalidation()

	case http.MethodDelete:
		studentIDStr := r.URL.Query().Get("student_id")
		courseID := r.URL.Query().Get("course_id")

		if studentIDStr == "" || courseID == "" {
			http.Error(w, "Student ID and Course ID are required", http.StatusBadRequest)
			return
		}

		studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Student ID must be a valid number", http.StatusBadRequest)
			return
		}

		err = app.queries.DeleteInvitation(ctx, DeleteInvitationParams{
			StudentID: studentID,
			CourseID:  courseID,
		})
		if err != nil {
			http.Error(w, "Failed to delete invitation", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
		app.BroadcastInvitationsInvalidation()

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) handleCreateOrUpdateInvitation(ctx context.Context, w http.ResponseWriter, r *http.Request, isUpdate bool) error {
	studentIDStr := strings.TrimSpace(r.FormValue("student_id"))
	courseID := strings.TrimSpace(r.FormValue("course_id"))
	invitationType := strings.TrimSpace(r.FormValue("invitation_type"))

	// Validation
	if studentIDStr == "" || courseID == "" || invitationType == "" {
		http.Error(w, "Student ID, Course ID, and invitation type are required", http.StatusBadRequest)
		return fmt.Errorf("validation failed")
	}

	studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Student ID must be a valid number", http.StatusBadRequest)
		return fmt.Errorf("validation failed")
	}

	if invitationType != "invite" && invitationType != "force" {
		http.Error(w, "Invitation type must be 'invite' or 'force'", http.StatusBadRequest)
		return fmt.Errorf("validation failed")
	}

	// Get the course to find its period
	course, err := app.queries.GetCourseByID(ctx, courseID)
	if err != nil {
		http.Error(w, "Course not found", http.StatusBadRequest)
		return fmt.Errorf("course not found")
	}

	// Verify student exists
	_, err = app.queries.GetStudentByID(ctx, studentID)
	if err != nil {
		http.Error(w, "Student not found", http.StatusBadRequest)
		return fmt.Errorf("student not found")
	}

	// Create or update the invitation
	err = app.queries.CreateOrUpdateInvitation(ctx, CreateOrUpdateInvitationParams{
		StudentID:      studentID,
		CourseID:       courseID,
		PeriodID:       course.PeriodID,
		InvitationType: InvitationType(invitationType),
	})
	if err != nil {
		http.Error(w, "Failed to create/update invitation", http.StatusInternalServerError)
		return err
	}

	status := http.StatusCreated
	if isUpdate {
		status = http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"student_id": studentID,
		"course_id":  courseID,
	})
	return nil
}

// CSV handlers for invitations
func (app *App) invitationsCSVExampleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"invitations_example.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{"student_id", "course_id", "invitation_type"}
	writer.Write(header)

	// Write example row
	example := []string{"12345", "COURSE1", "invite"}
	writer.Write(example)
}

func (app *App) invitationsCSVDownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	invitations, err := app.queries.GetAllInvitations(ctx)
	if err != nil {
		http.Error(w, "Failed to fetch invitations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"invitations.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{"student_id", "course_id", "invitation_type"}
	writer.Write(header)

	// Write invitations
	for _, invitation := range invitations {
		row := []string{
			fmt.Sprintf("%d", invitation.StudentID),
			invitation.CourseID,
			string(invitation.InvitationType),
		}
		writer.Write(row)
	}
}

func (app *App) invitationsCSVPreviewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("csv")
	if err != nil {
		http.Error(w, "Failed to read uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(w, "Failed to parse CSV file: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(records) == 0 {
		http.Error(w, "CSV file is empty", http.StatusBadRequest)
		return
	}

	// Validate header
	expectedHeader := []string{"student_id", "course_id", "invitation_type"}
	header := records[0]
	if len(header) != len(expectedHeader) {
		http.Error(w, fmt.Sprintf("CSV header must have %d columns", len(expectedHeader)), http.StatusBadRequest)
		return
	}

	for i, expected := range expectedHeader {
		if header[i] != expected {
			http.Error(w, fmt.Sprintf("Column %d should be '%s', got '%s'", i+1, expected, header[i]), http.StatusBadRequest)
			return
		}
	}

	ctx := context.Background()

	// Preview data structure
	type PreviewRow struct {
		RowNumber  int                    `json:"row_number"`
		Data       map[string]interface{} `json:"data"`
		Errors     []string               `json:"errors"`
		IsValid    bool                   `json:"is_valid"`
		WillUpdate bool                   `json:"will_update"`
	}

	preview := make([]PreviewRow, 0, len(records)-1)
	var hasErrors bool

	// Process each row
	for rowIdx, record := range records[1:] {
		rowNumber := rowIdx + 2
		errors := []string{}
		data := make(map[string]interface{})

		if len(record) != len(expectedHeader) {
			errors = append(errors, fmt.Sprintf("Has %d columns, expected %d", len(record), len(expectedHeader)))
		} else {
			studentIDStr := strings.TrimSpace(record[0])
			courseID := strings.TrimSpace(record[1])
			invitationType := strings.TrimSpace(record[2])

			// Store data
			data["student_id"] = studentIDStr
			data["course_id"] = courseID
			data["invitation_type"] = invitationType

			// Validation
			if studentIDStr == "" {
				errors = append(errors, "Student ID is required")
			} else {
				studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
				if err != nil {
					errors = append(errors, "Student ID must be a valid number")
				} else {
					data["student_id"] = studentID
					// Check if student exists
					_, err := app.queries.GetStudentByID(ctx, studentID)
					if err != nil {
						errors = append(errors, fmt.Sprintf("Student %d does not exist", studentID))
					}
				}
			}

			if courseID == "" {
				errors = append(errors, "Course ID is required")
			} else {
				// Check if course exists
				_, err := app.queries.GetCourseByID(ctx, courseID)
				if err != nil {
					errors = append(errors, fmt.Sprintf("Course '%s' does not exist", courseID))
				}
			}

			if invitationType != "invite" && invitationType != "force" {
				errors = append(errors, "Invitation type must be 'invite' or 'force'")
			}

			// Check if invitation exists (for updates)
			willUpdate := false
			if studentIDStr != "" && courseID != "" {
				studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
				if err == nil {
					_, err := app.queries.GetInvitation(ctx, GetInvitationParams{
						StudentID: studentID,
						CourseID:  courseID,
					})
					willUpdate = (err == nil)
				}
			}
			data["will_update"] = willUpdate
		}

		isValid := len(errors) == 0
		if !isValid {
			hasErrors = true
		}

		preview = append(preview, PreviewRow{
			RowNumber:  rowNumber,
			Data:       data,
			Errors:     errors,
			IsValid:    isValid,
			WillUpdate: data["will_update"] == true,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    !hasErrors,
		"has_errors": hasErrors,
		"preview":    preview,
		"total_rows": len(records) - 1,
	})
}

func (app *App) invitationsCSVUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("csv")
	if err != nil {
		http.Error(w, "Failed to read uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		http.Error(w, "Failed to parse CSV file: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(records) == 0 {
		http.Error(w, "CSV file is empty", http.StatusBadRequest)
		return
	}

	// Validate header
	expectedHeader := []string{"student_id", "course_id", "invitation_type"}
	header := records[0]
	if len(header) != len(expectedHeader) {
		http.Error(w, fmt.Sprintf("CSV header must have %d columns", len(expectedHeader)), http.StatusBadRequest)
		return
	}

	for i, expected := range expectedHeader {
		if header[i] != expected {
			http.Error(w, fmt.Sprintf("Column %d should be '%s', got '%s'", i+1, expected, header[i]), http.StatusBadRequest)
			return
		}
	}

	ctx := context.Background()

	// Validate and process each row
	for rowIdx, record := range records[1:] {
		if len(record) != len(expectedHeader) {
			http.Error(w, fmt.Sprintf("Row %d has %d columns, expected %d", rowIdx+2, len(record), len(expectedHeader)), http.StatusBadRequest)
			return
		}

		studentIDStr := strings.TrimSpace(record[0])
		courseID := strings.TrimSpace(record[1])
		invitationType := strings.TrimSpace(record[2])

		// Validation
		if studentIDStr == "" || courseID == "" || invitationType == "" {
			http.Error(w, fmt.Sprintf("Row %d: All fields are required", rowIdx+2), http.StatusBadRequest)
			return
		}

		studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Row %d: Student ID must be a valid number", rowIdx+2), http.StatusBadRequest)
			return
		}

		if invitationType != "invite" && invitationType != "force" {
			http.Error(w, fmt.Sprintf("Row %d: Invitation type must be 'invite' or 'force'", rowIdx+2), http.StatusBadRequest)
			return
		}

		// Get the course to find its period
		course, err := app.queries.GetCourseByID(ctx, courseID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Row %d: Course '%s' not found", rowIdx+2, courseID), http.StatusBadRequest)
			return
		}

		// Verify student exists
		_, err = app.queries.GetStudentByID(ctx, studentID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Row %d: Student %d not found", rowIdx+2, studentID), http.StatusBadRequest)
			return
		}

		// Create or update the invitation
		err = app.queries.CreateOrUpdateInvitation(ctx, CreateOrUpdateInvitationParams{
			StudentID:      studentID,
			CourseID:       courseID,
			PeriodID:       course.PeriodID,
			InvitationType: InvitationType(invitationType),
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Row %d: Failed to create/update invitation: %v", rowIdx+2, err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"imported": len(records) - 1,
	})

	app.BroadcastInvitationsInvalidation()
}

// Grade selection controls handlers
func (app *App) selectionControlsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	switch r.Method {
	case http.MethodGet:
		controls, err := app.queries.GetAllGradeSelectionControls(ctx)
		if err != nil {
			http.Error(w, "Failed to fetch selection controls", http.StatusInternalServerError)
			return
		}

		// Ensure we return empty array instead of null when there are no selection controls
		if controls == nil {
			controls = []GradeSelectionControl{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(controls)

	case http.MethodPost:
		gradeStr := r.FormValue("grade")
		enabledStr := r.FormValue("enabled")

		if gradeStr == "" || enabledStr == "" {
			http.Error(w, "Grade and enabled are required", http.StatusBadRequest)
			return
		}

		grade, err := strconv.ParseInt(gradeStr, 10, 64)
		if err != nil {
			http.Error(w, "Grade must be a valid number", http.StatusBadRequest)
			return
		}

		enabled, err := strconv.ParseBool(enabledStr)
		if err != nil {
			http.Error(w, "Enabled must be true or false", http.StatusBadRequest)
			return
		}

		err = app.queries.CreateOrUpdateGradeSelectionControl(ctx, CreateOrUpdateGradeSelectionControlParams{
			Grade:   grade,
			Enabled: enabled,
		})
		if err != nil {
			http.Error(w, "Failed to create/update selection control", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"grade":   grade,
		})

		// Broadcast to admins for UI updates
		app.BroadcastSelectionControlsInvalidation()

		// Broadcast to students of the affected grade about their selection status change
		app.BroadcastGradeSelectionStatusChange(grade, enabled)

	case http.MethodDelete:
		gradeStr := r.URL.Query().Get("grade")
		if gradeStr == "" {
			http.Error(w, "Grade is required", http.StatusBadRequest)
			return
		}

		grade, err := strconv.ParseInt(gradeStr, 10, 64)
		if err != nil {
			http.Error(w, "Grade must be a valid number", http.StatusBadRequest)
			return
		}

		err = app.queries.DeleteGradeSelectionControl(ctx, grade)
		if err != nil {
			http.Error(w, "Failed to delete selection control", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})

		// Broadcast to admins for UI updates
		app.BroadcastSelectionControlsInvalidation()

		// Broadcast to students of the affected grade that selections are now disabled
		app.BroadcastGradeSelectionStatusChange(grade, false)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Tab 7: Admin selections management
func (app *App) adminSelectionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	switch r.Method {
	case http.MethodGet:
		selections, err := app.queries.GetAllSelections(ctx)
		if err != nil {
			http.Error(w, "Failed to fetch selections", http.StatusInternalServerError)
			return
		}

		// Ensure we return empty array instead of null when there are no selections
		if selections == nil {
			selections = []GetAllSelectionsRow{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(selections)

	case http.MethodPost:
		// Add a new selection
		studentIDStr := r.FormValue("student_id")
		courseID := r.FormValue("course_id")
		periodID := r.FormValue("period_id")
		invitationTypeStr := r.FormValue("invitation_type")

		if studentIDStr == "" || courseID == "" || periodID == "" || invitationTypeStr == "" {
			http.Error(w, "student_id, course_id, period_id, and invitation_type are required", http.StatusBadRequest)
			return
		}

		studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid student_id", http.StatusBadRequest)
			return
		}

		// Validate invitation type
		if invitationTypeStr != "no" && invitationTypeStr != "invite" && invitationTypeStr != "force" {
			http.Error(w, "invitation_type must be 'no', 'invite', or 'force'", http.StatusBadRequest)
			return
		}

		// Get admin user for logging
		adminUser, _ := app.authenticateRequest(r)
		actorUsername := "unknown"
		if adminUser != nil {
			actorUsername = adminUser.Username
		}

		err = app.queries.AdminAddSelection(ctx, AdminAddSelectionParams{
			StudentID:      studentID,
			CourseID:       courseID,
			PeriodID:       periodID,
			InvitationType: InvitationType(invitationTypeStr),
		})
		if err != nil {
			if strings.Contains(err.Error(), "foreign key") {
				http.Error(w, "Invalid student_id, course_id, or period_id", http.StatusBadRequest)
			} else if strings.Contains(err.Error(), "capacity") {
				http.Error(w, "Course is at capacity", http.StatusConflict)
			} else if strings.Contains(err.Error(), "not allowed") {
				http.Error(w, "Student does not meet course requirements", http.StatusConflict)
			} else if strings.Contains(err.Error(), "PRIMARY KEY") || strings.Contains(err.Error(), "duplicate") {
				http.Error(w, "Student already has a selection for this period or course", http.StatusConflict)
			} else {
				http.Error(w, "Failed to add selection: "+err.Error(), http.StatusInternalServerError)
			}
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
		app.BroadcastSelectionAction("added", studentID, courseID, invitationTypeStr, actorUsername)

		// If admin is modifying a student's selections, notify the student
		app.BroadcastStudentSelectionsInvalidation(studentID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})

	case http.MethodPut:
		// Update an existing selection
		studentIDStr := r.FormValue("student_id")
		oldCourseID := r.FormValue("old_course_id")
		newCourseID := r.FormValue("new_course_id")
		newPeriodID := r.FormValue("new_period_id")
		newInvitationTypeStr := r.FormValue("new_invitation_type")

		if studentIDStr == "" || oldCourseID == "" || newCourseID == "" || newPeriodID == "" || newInvitationTypeStr == "" {
			http.Error(w, "All fields are required", http.StatusBadRequest)
			return
		}

		studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid student_id", http.StatusBadRequest)
			return
		}

		// Validate invitation type
		if newInvitationTypeStr != "no" && newInvitationTypeStr != "invite" && newInvitationTypeStr != "force" {
			http.Error(w, "invitation_type must be 'no', 'invite', or 'force'", http.StatusBadRequest)
			return
		}

		// Get admin user for logging
		adminUser, _ := app.authenticateRequest(r)
		actorUsername := "unknown"
		if adminUser != nil {
			actorUsername = adminUser.Username
		}

		err = app.queries.AdminUpdateSelection(ctx, AdminUpdateSelectionParams{
			StudentID:      studentID,
			CourseID:       oldCourseID,
			CourseID_2:     newCourseID,
			PeriodID:       newPeriodID,
			InvitationType: InvitationType(newInvitationTypeStr),
		})
		if err != nil {
			if strings.Contains(err.Error(), "foreign key") {
				http.Error(w, "Invalid student_id, course_id, or period_id", http.StatusBadRequest)
			} else if strings.Contains(err.Error(), "capacity") {
				http.Error(w, "Course is at capacity", http.StatusConflict)
			} else if strings.Contains(err.Error(), "not allowed") {
				http.Error(w, "Student does not meet course requirements", http.StatusConflict)
			} else if strings.Contains(err.Error(), "PRIMARY KEY") || strings.Contains(err.Error(), "duplicate") {
				http.Error(w, "Student already has a selection for this period or course", http.StatusConflict)
			} else {
				http.Error(w, "Failed to update selection: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Get updated enrollment counts for both old and new courses (if different)
		oldEnrollment, err := app.queries.GetCourseCurrentEnrollment(ctx, oldCourseID)
		if err != nil {
			log.Printf("Warning: Failed to get enrollment count for old course %s: %v", oldCourseID, err)
			oldEnrollment = 0
		}

		// Broadcast enrollment update for old course
		app.BroadcastCourseEnrollmentUpdate(oldCourseID, oldEnrollment)

		// If the course changed, also update the new course enrollment
		if oldCourseID != newCourseID {
			newEnrollment, err := app.queries.GetCourseCurrentEnrollment(ctx, newCourseID)
			if err != nil {
				log.Printf("Warning: Failed to get enrollment count for new course %s: %v", newCourseID, err)
				newEnrollment = 0
			}
			app.BroadcastCourseEnrollmentUpdate(newCourseID, newEnrollment)
		}

		// Broadcast selection action to admins
		app.BroadcastSelectionAction("updated", studentID, newCourseID, newInvitationTypeStr, actorUsername)

		// If admin is modifying a student's selections, notify the student
		app.BroadcastStudentSelectionsInvalidation(studentID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})

	case http.MethodDelete:
		// Delete a selection
		studentIDStr := r.URL.Query().Get("student_id")
		courseID := r.URL.Query().Get("course_id")

		if studentIDStr == "" || courseID == "" {
			http.Error(w, "student_id and course_id are required", http.StatusBadRequest)
			return
		}

		studentID, err := strconv.ParseInt(studentIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid student_id", http.StatusBadRequest)
			return
		}

		// Get admin user for logging
		adminUser, _ := app.authenticateRequest(r)
		actorUsername := "unknown"
		if adminUser != nil {
			actorUsername = adminUser.Username
		}

		err = app.queries.AdminDeleteSelection(ctx, AdminDeleteSelectionParams{
			StudentID: studentID,
			CourseID:  courseID,
		})
		if err != nil {
			http.Error(w, "Failed to delete selection", http.StatusInternalServerError)
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
		app.BroadcastSelectionAction("removed", studentID, courseID, "no", actorUsername)

		// If admin is modifying a student's selections, notify the student
		app.BroadcastStudentSelectionsInvalidation(studentID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
