package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// SSEClient represents a connected SSE client
type SSEClient struct {
	ID       string
	UserID   int64
	Role     string
	Username string
	Channel  chan SSEMessage
	Done     chan bool
}

// SSEMessage represents a server-sent event message
type SSEMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// SSEHub manages all connected SSE clients
type SSEHub struct {
	clients    map[string]*SSEClient
	clientsMux sync.RWMutex
}

// NewSSEHub creates a new SSE hub
func NewSSEHub() *SSEHub {
	return &SSEHub{
		clients: make(map[string]*SSEClient),
	}
}

// AddClient adds a new SSE client
func (h *SSEHub) AddClient(client *SSEClient) {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	h.clients[client.ID] = client
	log.Printf("SSE client connected: %s (user: %s, role: %s)", client.ID, client.Username, client.Role)
}

// RemoveClient removes an SSE client
func (h *SSEHub) RemoveClient(clientID string) {
	h.clientsMux.Lock()
	defer h.clientsMux.Unlock()

	if client, exists := h.clients[clientID]; exists {
		close(client.Channel)
		close(client.Done)
		delete(h.clients, clientID)
		log.Printf("SSE client disconnected: %s", clientID)
	}
}

// BroadcastToAll sends a message to all connected clients
func (h *SSEHub) BroadcastToAll(message SSEMessage) {
	h.clientsMux.RLock()
	defer h.clientsMux.RUnlock()

	for clientID, client := range h.clients {
		select {
		case client.Channel <- message:
			// Message sent successfully
		case <-time.After(5 * time.Second):
			// Client appears to be unresponsive, remove it
			log.Printf("SSE client %s appears unresponsive, removing", clientID)
			go h.RemoveClient(clientID)
		}
	}
}

// BroadcastToAdmins sends a message to all admin clients
func (h *SSEHub) BroadcastToAdmins(message SSEMessage) {
	h.clientsMux.RLock()
	defer h.clientsMux.RUnlock()

	for clientID, client := range h.clients {
		if client.Role != "admin" {
			continue
		}

		select {
		case client.Channel <- message:
			// Message sent successfully
		case <-time.After(5 * time.Second):
			// Client appears to be unresponsive, remove it
			log.Printf("SSE client %s appears unresponsive, removing", clientID)
			go h.RemoveClient(clientID)
		}
	}
}

// BroadcastToStudents sends a message to all student clients
func (h *SSEHub) BroadcastToStudents(message SSEMessage) {
	h.clientsMux.RLock()
	defer h.clientsMux.RUnlock()

	for clientID, client := range h.clients {
		if client.Role != "student" {
			continue
		}

		select {
		case client.Channel <- message:
			// Message sent successfully
		case <-time.After(5 * time.Second):
			// Client appears to be unresponsive, remove it
			log.Printf("SSE client %s appears unresponsive, removing", clientID)
			go h.RemoveClient(clientID)
		}
	}
}

// GetClientCount returns the number of connected clients
func (h *SSEHub) GetClientCount() int {
	h.clientsMux.RLock()
	defer h.clientsMux.RUnlock()

	return len(h.clients)
}

// sseHandler handles server-sent event connections
func (app *App) sseHandler(w http.ResponseWriter, r *http.Request) {
	// Authenticate the request
	user, err := app.authenticateRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Create a new SSE client
	clientID := fmt.Sprintf("%s_%d_%d", user.Username, user.ID, time.Now().UnixNano())
	client := &SSEClient{
		ID:       clientID,
		UserID:   user.ID,
		Role:     user.Role,
		Username: user.Username,
		Channel:  make(chan SSEMessage, 100), // Buffer up to 100 messages
		Done:     make(chan bool),
	}

	// Add client to the hub
	app.sseHub.AddClient(client)
	defer app.sseHub.RemoveClient(clientID)

	// Send initial connection message
	initialMessage := SSEMessage{
		Type: "connected",
		Data: map[string]interface{}{
			"clientId": clientID,
			"time":     time.Now().Unix(),
		},
	}

	if err := app.writeSSEMessage(w, initialMessage); err != nil {
		return
	}

	// Flush the response to establish the connection
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// Handle client connection
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			return
		case <-client.Done:
			// Client cleanup requested
			return
		case message := <-client.Channel:
			// Send message to client
			if err := app.writeSSEMessage(w, message); err != nil {
				log.Printf("Error writing SSE message to client %s: %v", clientID, err)
				return
			}

			// Flush the message
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		case <-time.After(30 * time.Second):
			// Send keepalive ping
			keepalive := SSEMessage{
				Type: "ping",
				Data: map[string]interface{}{
					"time": time.Now().Unix(),
				},
			}

			if err := app.writeSSEMessage(w, keepalive); err != nil {
				return
			}

			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}

// writeSSEMessage writes an SSE message to the response writer
func (app *App) writeSSEMessage(w http.ResponseWriter, message SSEMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal SSE message: %v", err)
	}

	// Write the SSE message format
	_, err = fmt.Fprintf(w, "event: %s\n", message.Type)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "data: %s\n\n", string(data))
	if err != nil {
		return err
	}

	return nil
}

// Broadcast methods for different data types

// BroadcastGradesInvalidation broadcasts a grades invalidation message
func (app *App) BroadcastGradesInvalidation() {
	message := SSEMessage{
		Type: "invalidate_grades",
		Data: map[string]interface{}{
			"time": time.Now().Unix(),
		},
	}
	app.sseHub.BroadcastToAll(message)
}

// BroadcastPeriodsInvalidation broadcasts a periods invalidation message
func (app *App) BroadcastPeriodsInvalidation() {
	message := SSEMessage{
		Type: "invalidate_periods",
		Data: map[string]interface{}{
			"time": time.Now().Unix(),
		},
	}
	app.sseHub.BroadcastToAll(message)
}

// BroadcastCategoriesInvalidation broadcasts a categories invalidation message
func (app *App) BroadcastCategoriesInvalidation() {
	message := SSEMessage{
		Type: "invalidate_categories",
		Data: map[string]interface{}{
			"time": time.Now().Unix(),
		},
	}
	app.sseHub.BroadcastToAll(message)
}

// BroadcastCoursesInvalidation broadcasts a courses invalidation message
func (app *App) BroadcastCoursesInvalidation() {
	message := SSEMessage{
		Type: "invalidate_courses",
		Data: map[string]interface{}{
			"time": time.Now().Unix(),
		},
	}
	app.sseHub.BroadcastToAll(message)
}

// BroadcastRequirementsInvalidation broadcasts a requirements invalidation message
func (app *App) BroadcastRequirementsInvalidation() {
	message := SSEMessage{
		Type: "invalidate_requirements",
		Data: map[string]interface{}{
			"time": time.Now().Unix(),
		},
	}
	app.sseHub.BroadcastToAll(message)
}

// BroadcastStudentsInvalidation broadcasts a students invalidation message
func (app *App) BroadcastStudentsInvalidation() {
	message := SSEMessage{
		Type: "invalidate_students",
		Data: map[string]interface{}{
			"time": time.Now().Unix(),
		},
	}
	app.sseHub.BroadcastToAdmins(message) // Only admins should know about student changes
}

// BroadcastInvitationsInvalidation broadcasts an invitations invalidation message
func (app *App) BroadcastInvitationsInvalidation() {
	message := SSEMessage{
		Type: "invalidate_invitations",
		Data: map[string]interface{}{
			"time": time.Now().Unix(),
		},
	}
	app.sseHub.BroadcastToAdmins(message) // Only admins manage invitations
}

// BroadcastSelectionControlsInvalidation broadcasts a selection controls invalidation message
func (app *App) BroadcastSelectionControlsInvalidation() {
	message := SSEMessage{
		Type: "invalidate_selection_controls",
		Data: map[string]interface{}{
			"time": time.Now().Unix(),
		},
	}
	app.sseHub.BroadcastToAdmins(message) // Only admins manage selection controls
}

// BroadcastToStudentsOfGrade sends a message to all students of a specific grade
func (app *App) BroadcastToStudentsOfGrade(grade int64, message SSEMessage) {
	app.sseHub.clientsMux.RLock()
	defer app.sseHub.clientsMux.RUnlock()

	for clientID, client := range app.sseHub.clients {
		if client.Role != "student" {
			continue
		}

		// Get student's grade - we need to query the database for this
		// We'll store the grade in memory for efficiency, but for now let's query
		// In a production system, we might want to store grade in the SSEClient struct
		ctx := context.Background()
		student, err := app.queries.GetStudentByID(ctx, client.UserID)
		if err != nil {
			log.Printf("Failed to get student grade for SSE client %s: %v", clientID, err)
			continue
		}

		if student.Grade != grade {
			continue
		}

		select {
		case client.Channel <- message:
			// Message sent successfully
		case <-time.After(5 * time.Second):
			// Client appears to be unresponsive, remove it
			log.Printf("SSE client %s appears unresponsive, removing", clientID)
			go app.sseHub.RemoveClient(clientID)
		}
	}
}

// BroadcastGradeSelectionStatusChange broadcasts selection status change to students of a specific grade
func (app *App) BroadcastGradeSelectionStatusChange(grade int64, enabled bool) {
	message := SSEMessage{
		Type: "grade_selection_status_change",
		Data: map[string]interface{}{
			"grade":   grade,
			"enabled": enabled,
			"time":    time.Now().Unix(),
		},
	}
	app.BroadcastToStudentsOfGrade(grade, message)
}

// BroadcastSelectionsInvalidation broadcasts a selections invalidation message
func (app *App) BroadcastSelectionsInvalidation() {
	message := SSEMessage{
		Type: "invalidate_selections",
		Data: map[string]interface{}{
			"time": time.Now().Unix(),
		},
	}
	app.sseHub.BroadcastToAll(message) // All users need to know about selection changes
}

// BroadcastCourseEnrollmentUpdate broadcasts fine-grained course enrollment update to all users
func (app *App) BroadcastCourseEnrollmentUpdate(courseID string, currentEnrollment int64) {
	message := SSEMessage{
		Type: "course_enrollment_update",
		Data: map[string]interface{}{
			"course_id":          courseID,
			"current_enrollment": currentEnrollment,
			"time":               time.Now().Unix(),
		},
	}
	app.sseHub.BroadcastToAll(message)
}

// BroadcastSelectionAction broadcasts admin-specific selection action notifications
func (app *App) BroadcastSelectionAction(action string, studentID int64, courseID string, invitationType string, actorUsername string) {
	message := SSEMessage{
		Type: "selection_action",
		Data: map[string]interface{}{
			"action":          action, // "added", "removed", "updated"
			"student_id":      studentID,
			"course_id":       courseID,
			"invitation_type": invitationType,
			"actor":           actorUsername,
			"time":            time.Now().Unix(),
		},
	}
	app.sseHub.BroadcastToAdmins(message)
}

// BroadcastToSpecificStudent sends a message to a specific student user
func (app *App) BroadcastToSpecificStudent(studentID int64, message SSEMessage) {
	app.sseHub.clientsMux.RLock()
	defer app.sseHub.clientsMux.RUnlock()

	for clientID, client := range app.sseHub.clients {
		if client.Role == "student" && client.UserID == studentID {
			select {
			case client.Channel <- message:
				// Message sent successfully
			case <-time.After(5 * time.Second):
				// Client appears to be unresponsive, remove it
				log.Printf("SSE client %s appears unresponsive, removing", clientID)
				go app.sseHub.RemoveClient(clientID)
			}
		}
	}
}

// BroadcastStudentSelectionsInvalidation sends selections invalidation to a specific student
func (app *App) BroadcastStudentSelectionsInvalidation(studentID int64) {
	message := SSEMessage{
		Type: "invalidate_student_selections_by_period",
		Data: map[string]interface{}{
			"student_id": studentID,
			"time":       time.Now().Unix(),
		},
	}
	app.BroadcastToSpecificStudent(studentID, message)
}
