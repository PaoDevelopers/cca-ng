package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	config  *Config
	pool    *pgxpool.Pool
	queries *Queries
	sseHub  *SSEHub
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	pool, queries, err := connectDatabase(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()
	if err := ensureDefaultAdmin(ctx, queries); err != nil {
		log.Fatalf("Failed to ensure default admin: %v", err)
	}

	app := &App{
		config:  config,
		pool:    pool,
		queries: queries,
		sseHub:  NewSSEHub(),
	}

	mux := http.NewServeMux()

	// Auth endpoints
	mux.HandleFunc("/api/login", app.loginHandler)
	mux.HandleFunc("/api/logout", app.logoutHandler)
	mux.HandleFunc("/api/change-password", app.changePasswordHandler)
	mux.HandleFunc("/api/status", app.apiStatusHandler)

	// SSE endpoint
	mux.HandleFunc("/api/events", app.sseHandler)

	// Admin endpoints
	mux.HandleFunc("/api/admin/grades", app.adminOnly(app.gradesHandler))
	mux.HandleFunc("/api/admin/periods", app.adminOnly(app.periodsHandler))
	mux.HandleFunc("/api/admin/categories", app.adminOnly(app.categoriesHandler))
	mux.HandleFunc("/api/admin/courses", app.adminOnly(app.coursesHandler))
	mux.HandleFunc("/api/admin/courses/delete-all", app.adminOnly(app.deleteAllCoursesHandler))
	mux.HandleFunc("/api/admin/courses/csv/example", app.adminOnly(app.coursesCSVExampleHandler))
	mux.HandleFunc("/api/admin/courses/csv/download", app.adminOnly(app.coursesCSVDownloadHandler))
	mux.HandleFunc("/api/admin/courses/csv/preview", app.adminOnly(app.coursesCSVPreviewHandler))
	mux.HandleFunc("/api/admin/courses/csv/upload", app.adminOnly(app.coursesCSVUploadHandler))

	// Requirements endpoints
	mux.HandleFunc("/api/admin/requirements", app.adminOnly(app.requirementsHandler))
	mux.HandleFunc("/api/admin/requirement-groups", app.adminOnly(app.requirementGroupsHandler))

	// Students endpoints
	mux.HandleFunc("/api/admin/students", app.adminOnly(app.studentsHandler))
	mux.HandleFunc("/api/admin/students/delete-all", app.adminOnly(app.deleteAllStudentsHandler))
	mux.HandleFunc("/api/admin/students/csv/example", app.adminOnly(app.studentsCSVExampleHandler))
	mux.HandleFunc("/api/admin/students/csv/download", app.adminOnly(app.studentsCSVDownloadHandler))
	mux.HandleFunc("/api/admin/students/csv/preview", app.adminOnly(app.studentsCSVPreviewHandler))
	mux.HandleFunc("/api/admin/students/csv/upload", app.adminOnly(app.studentsCSVUploadHandler))

	// Invitations endpoints
	mux.HandleFunc("/api/admin/invitations", app.adminOnly(app.invitationsHandler))
	mux.HandleFunc("/api/admin/invitations/csv/example", app.adminOnly(app.invitationsCSVExampleHandler))
	mux.HandleFunc("/api/admin/invitations/csv/download", app.adminOnly(app.invitationsCSVDownloadHandler))
	mux.HandleFunc("/api/admin/invitations/csv/preview", app.adminOnly(app.invitationsCSVPreviewHandler))
	mux.HandleFunc("/api/admin/invitations/csv/upload", app.adminOnly(app.invitationsCSVUploadHandler))

	// Selection controls endpoints
	mux.HandleFunc("/api/admin/selection-controls", app.adminOnly(app.selectionControlsHandler))

	// Selections management endpoints
	mux.HandleFunc("/api/admin/selections", app.adminOnly(app.adminSelectionsHandler))

	// Student-specific endpoints
	mux.HandleFunc("/api/student/courses", app.studentOnly(app.studentCoursesHandler))
	mux.HandleFunc("/api/student/selections", app.studentOnly(app.studentSelectionsHandler))
	mux.HandleFunc("/api/student/selections/by-period", app.studentOnly(app.studentSelectionsByPeriodHandler))
	mux.HandleFunc("/api/student/select-course", app.studentOnly(app.studentSelectCourseHandler))
	mux.HandleFunc("/api/student/requirements", app.studentOnly(app.studentRequirementsHandler))
	mux.HandleFunc("/api/student/selection-status", app.studentOnly(app.studentSelectionStatusHandler))

	// Main handler for everything else
	mux.HandleFunc("/", app.rootHandler)

	log.Printf("Starting server on %s", config.Server.Address)
	log.Fatal(http.ListenAndServe(config.Server.Address, mux))
}
