package handlers

import (
	"html/template"
	"net/http"
	"time"

	"task-dashboard/internal/models"
	"task-dashboard/internal/store"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	store     *store.RedisStore
	templates *template.Template
}

func NewAuthHandler(store *store.RedisStore, templates *template.Template) *AuthHandler {
	return &AuthHandler{
		store:     store,
		templates: templates,
	}
}

func (h *AuthHandler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	err := h.templates.ExecuteTemplate(w, "login.html", nil)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler) ShowRegister(w http.ResponseWriter, r *http.Request) {
	err := h.templates.ExecuteTemplate(w, "register.html", nil)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	name := r.FormValue("name")

	// Check if user exists
	if _, err := h.store.GetUserByEmail(r.Context(), email); err == nil {
		w.Header().Set("HX-Retarget", "#error-message")
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte(`<div id="error-message" class="error">Email already registered</div>`))
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Create user
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		CreatedAt:    time.Now(),
	}

	if err := h.store.CreateUser(r.Context(), user); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if err := h.store.SetEmailIndex(r.Context(), email, user.ID); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/login")
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.store.GetUserByEmail(r.Context(), email)
	if err != nil {
		w.Header().Set("HK-Retarget", "#error-message")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`<div id="error-message" class="error">Invalid credentials</div>`))
		return

	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		w.Header().Set("HX-Retarget", "#error-message")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`<div id="error-message" class="error">Invalid credentials</div>`))
		return
	}

	// Create session
	session := &models.Session{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := h.store.CreateSession(r.Context(), session); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Path:     "/",
	})

	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		err = h.store.DeleteSession(r.Context(), cookie.Value)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
