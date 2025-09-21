package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserInfo struct {
	ID       int64
	Username string
	Role     string // "admin" or "student"
}

func generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (app *App) hashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func (app *App) authenticateRequest(r *http.Request) (*UserInfo, error) {
	cookie, err := r.Cookie(app.config.Server.SessionName)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	// Try admin first
	admin, err := app.queries.GetAdminBySessionToken(ctx, pgtype.Text{String: cookie.Value, Valid: true})
	if err == nil {
		return &UserInfo{
			ID:       admin.ID,
			Username: admin.Username,
			Role:     "admin",
		}, nil
	}

	// Try student
	student, err := app.queries.GetStudentBySessionToken(ctx, pgtype.Text{String: cookie.Value, Valid: true})
	if err == nil {
		return &UserInfo{
			ID:       student.ID,
			Username: fmt.Sprintf("%d", student.ID), // Use ID as username for students
			Role:     "student",
		}, nil
	}

	return nil, fmt.Errorf("invalid session token")
}

func (app *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Try admin login first
	admin, err := app.queries.GetAdminByUsername(ctx, username)
	if err == nil {
		match, err := argon2id.ComparePasswordAndHash(password, admin.PasswordHash)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if match {
			token, err := generateSessionToken()
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			err = app.queries.UpdateAdminSessionToken(ctx, UpdateAdminSessionTokenParams{
				ID:           admin.ID,
				SessionToken: pgtype.Text{String: token, Valid: true},
			})
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     app.config.Server.SessionName,
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				Secure:   r.TLS != nil,
				SameSite: http.SameSiteLaxMode,
				Expires:  time.Now().Add(24 * time.Hour),
			})

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"role": "admin"}`))
			return
		}
	}

	// Try student login - username should be numeric for student ID
	var studentID int64
	if _, err := fmt.Sscanf(username, "%d", &studentID); err == nil {
		student, err := app.queries.GetStudentByID(ctx, studentID)
		if err == nil {
			match, err := argon2id.ComparePasswordAndHash(password, student.PasswordHash)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if match {
				token, err := generateSessionToken()
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				err = app.queries.UpdateStudentSessionToken(ctx, UpdateStudentSessionTokenParams{
					ID:           student.ID,
					SessionToken: pgtype.Text{String: token, Valid: true},
				})
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				http.SetCookie(w, &http.Cookie{
					Name:     app.config.Server.SessionName,
					Value:    token,
					Path:     "/",
					HttpOnly: true,
					Secure:   r.TLS != nil,
					SameSite: http.SameSiteLaxMode,
					Expires:  time.Now().Add(24 * time.Hour),
				})

				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"role": "student"}`))
				return
			}
		}
	}

	http.Error(w, "Invalid username or password", http.StatusUnauthorized)
}

func (app *App) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := app.authenticateRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()

	if user.Role == "admin" {
		app.queries.ClearAdminSessionToken(ctx, user.ID)
	} else {
		app.queries.ClearStudentSessionToken(ctx, user.ID)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     app.config.Server.SessionName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(-time.Hour),
	})

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true}`))
}

func (app *App) changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := app.authenticateRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	currentPassword := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")

	if currentPassword == "" || newPassword == "" {
		http.Error(w, "Current password and new password are required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	if user.Role == "admin" {
		// Get the session token from the cookie
		cookie, err := r.Cookie(app.config.Server.SessionName)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		admin, err := app.queries.GetAdminBySessionToken(ctx, pgtype.Text{String: cookie.Value, Valid: true})
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		match, err := argon2id.ComparePasswordAndHash(currentPassword, admin.PasswordHash)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !match {
			http.Error(w, "Current password is incorrect", http.StatusBadRequest)
			return
		}

		hash, err := argon2id.CreateHash(newPassword, argon2id.DefaultParams)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		err = app.queries.UpdateAdminPassword(ctx, UpdateAdminPasswordParams{
			ID:           user.ID,
			PasswordHash: hash,
		})
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		student, err := app.queries.GetStudentByID(ctx, user.ID)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		match, err := argon2id.ComparePasswordAndHash(currentPassword, student.PasswordHash)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !match {
			http.Error(w, "Current password is incorrect", http.StatusBadRequest)
			return
		}

		hash, err := argon2id.CreateHash(newPassword, argon2id.DefaultParams)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		err = app.queries.UpdateStudentPassword(ctx, UpdateStudentPasswordParams{
			ID:           user.ID,
			PasswordHash: hash,
		})
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true}`))
}
