package handlers

import (
	"html/template"
	"net/http"

	"task-dashboard/internal/middleware"
	"task-dashboard/internal/store"
)

type DashboardHandler struct {
	store     *store.RedisStore
	templates *template.Template
}

func NewDashboardHandler(store *store.RedisStore, templates *template.Template) *DashboardHandler {
	return &DashboardHandler{
		store:     store,
		templates: templates,
	}
}

func (h *DashboardHandler) ShowDashboard(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	user, err := h.store.GetUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	stats, err := h.store.GetDashboardStats(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	tasks, err := h.store.GetUserTasks(r.Context(), userID, "")
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if err := h.templates.ExecuteTemplate(w, "dashboard.html", map[string]interface{}{
		"User":  user,
		"Stats": stats,
		"Tasks": tasks,
	}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}

func (h *DashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	stats, err := h.store.GetDashboardStats(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if err := h.templates.ExecuteTemplate(w, "stats.html", stats); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
}
